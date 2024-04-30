package mux

import (
	"net"
	"server/enums"
	"server/models"
	"server/pkg/auth"
	"server/services"
	"sync"
)

type Handler func(msg models.Message, responseChan chan models.Message, transactions *sync.Map, fileService services.FileService)

type Mux interface {
	Handle(action enums.MessageType, handler Handler)
	ServeConn(net.Conn)
}

type concreteMux struct {
	handlers      map[enums.MessageType]Handler
	transactions  *sync.Map
	authenticator auth.Authenticator
	fileService   services.FileService
}

func NewMux(authenticator auth.Authenticator) Mux {
	return &concreteMux{
		make(map[enums.MessageType]Handler),
		new(sync.Map),
		authenticator,
		nil,
	}
}

func (m *concreteMux) Handle(action enums.MessageType, handlerFunc Handler) {
	m.handlers[action] = handlerFunc
}

func (m *concreteMux) ServeConn(conn net.Conn) {
	err := m.authenticator.AuthenticateClient(conn)
	if err != nil {
		return
	}
	m.fileService, err = m.authenticator.GetFileService()
	if err != nil {
		return
	}

	responseChan := make(chan models.Message, 5)
	go func() {
		for message := range responseChan {
			_, err = message.Send(conn)
			if err != nil {
				return
			}
		}
	}()

	for {
		var message models.Message
		_, err = message.Receive(conn)
		if err != nil {
			return
		}
		handler, ok := m.handlers[message.Header.Action]
		if !ok {
			return
		}
		go handler(message, responseChan, m.transactions, m.fileService)
	}
}
