package mux

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"server/enums"
	"server/models"
	"server/pkg/session"
	"server/services"
	"sync"
	"time"
)

type Request struct {
	Message models.Message
	Ctx     context.Context
}

type HandlerFunc func(chan models.Message, *Request)

type Mux interface {
	Handle(action enums.MessageType, handler HandlerFunc)
	ServeConn(net.Conn)
}

type concreteMux struct {
	handlers      map[enums.MessageType]HandlerFunc
	authenticator services.Authenticator
	ctx           context.Context
}

func NewMux(authenticator services.Authenticator, ctx context.Context) Mux {
	return &concreteMux{
		make(map[enums.MessageType]HandlerFunc),
		authenticator,
		ctx,
	}
}

func (m *concreteMux) Handle(action enums.MessageType, handlerFunc HandlerFunc) {
	m.handlers[action] = handlerFunc
}

func (m *concreteMux) ServeConn(conn net.Conn) {
	defer conn.Close()

	connCtx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	sessionData := &session.Data{
		Transactions: &sync.Map{},
	}

	err := m.authenticateClient(conn, sessionData)
	if err != nil {
		return
	}

	ctx := session.NewContext(connCtx, sessionData)

	resChan := handleResponses(conn, ctx)

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Warn("Connection timed out, shutting down connection")
			} else if errors.Is(ctx.Err(), context.Canceled) {
				log.Info("Context cancelled, shutting down connection")
			} else {
				log.Error("Unknown context error, shutting down connection", ctx.Err())
			}
			return
		default:
		}

		var message models.Message
		_, err = message.Receive(conn)
		if err != nil {
			log.Error("Error receiving message: ", err)
			return
		}

		if message.Header.Action == enums.Cancel {
			log.Info("Received cancel message, shutting down connection")
			return
		}

		req := &Request{
			Message: message,
			Ctx:     ctx,
		}
		err = m.handleRequest(resChan, req)
		if err != nil {
			log.Error("Error handling request: ", err)
		}
	}
}

func (m *concreteMux) authenticateClient(conn net.Conn, session *session.Data) error {
	err := m.authenticator.AuthenticateClient(conn)
	if err != nil {
		return err
	}
	session.FileService, err = m.authenticator.GetFileService()
	if err != nil {
		return err
	}
	session.Username = m.authenticator.GetUsername()
	return nil
}

func (m *concreteMux) handleRequest(resChan chan models.Message, req *Request) error {
	handler, ok := m.handlers[req.Message.Header.Action]
	if !ok {
		return errors.New("unknown action")
	}
	var sessionData *session.Data
	sessionData, ok = session.FromContext(req.Ctx)
	if !ok {
		return errors.New("no session data in context")
	}
	var transactionChan chan models.Message
	transactionChan, ok = sessionData.GetTransaction(req.Message.Header.TransactionID)
	if ok {
		transactionChan <- req.Message
		return nil
	}

	_ = sessionData.NewTransaction(req.Message.Header.TransactionID)
	go handler(resChan, req)

	return nil
}

// handleResponses listens for responses on the response channel and sends them to the client.
func handleResponses(conn net.Conn, ctx context.Context) chan models.Message {
	responseChan := make(chan models.Message, 5)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(responseChan)
				return
			case message, ok := <-responseChan:
				if !ok {
					log.Error("Response channel closed")
					return
				}
				_, err := message.Send(conn)
				if err != nil {
					log.Error("Error sending response: ", err)
				}
			}
		}
	}()
	return responseChan
}
