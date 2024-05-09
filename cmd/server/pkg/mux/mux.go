package mux

import (
	"context"
	"errors"
	"filesync/enums"
	"filesync/models"
	log "github.com/sirupsen/logrus"
	"net"
	"server/pkg/session"
	"server/services"
	"sync"
	"time"
)

type Request struct {
	Message models.Message
	Ctx     context.Context
}

type HandlerFunc func(chan models.Message, *Request) error

type Mux interface {
	Handle(action enums.MessageType, handler HandlerFunc)
	ServeConn(net.Conn)
	Shutdown()
}

type concreteMux struct {
	handlers      map[enums.MessageType]HandlerFunc
	authenticator services.AuthService
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewMux(authenticator services.AuthService) Mux {
	ctx, cancel := context.WithCancel(context.Background())
	return &concreteMux{
		make(map[enums.MessageType]HandlerFunc),
		authenticator,
		ctx,
		cancel,
	}
}

func (m *concreteMux) Shutdown() {
	m.cancel()
}

func (m *concreteMux) Handle(action enums.MessageType, handlerFunc HandlerFunc) {
	log.Debugf("Registering handler for action %s", action)
	m.handlers[action] = handlerFunc
}

func (m *concreteMux) ServeConn(conn net.Conn) {
	defer conn.Close()

	log.Debugf("Serving connection from %s", conn.RemoteAddr().String())

	sessionData := &session.Session{
		Transactions: &sync.Map{},
	}

	err := m.authenticateClient(conn, sessionData)
	if err != nil {
		return
	}

	ctx, cancel := session.NewContext(m.ctx, sessionData)
	defer cancel()

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

		reqCtx, cancelReq := context.WithTimeout(ctx, time.Second*5)
		req := &Request{
			Message: message,
			Ctx:     reqCtx,
		}
		err = m.handleRequest(resChan, req, cancelReq)
		if err != nil {
			log.Error("Error handling request: ", err)
		}
	}
}

func (m *concreteMux) authenticateClient(conn net.Conn, session *session.Session) error {
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

func (m *concreteMux) handleRequest(resChan chan models.Message, req *Request, cancel context.CancelFunc) error {
	log.Debugf("Handling request with action %s", req.Message.Header.Action)
	handler, ok := m.handlers[req.Message.Header.Action]
	if !ok {
		return errors.New("unknown action")
	}
	var sessionData *session.Session
	sessionData, ok = session.FromContext(req.Ctx)
	if !ok {
		return errors.New("no session data in context")
	}
	var transactionChan chan models.Message
	transactionChan, ok = sessionData.GetTransaction(req.Message.Header.TransactionID)
	if ok {
		log.Debug("Transaction found for request, forwarding message to transaction channel")
		transactionChan <- req.Message
		return nil
	}
	go func() {
		defer sessionData.Transactions.Delete(req.Message.Header.TransactionID)
		defer cancel()

		err := handler(resChan, req)
		if err != nil {
			log.Error("Error handling request: ", err)
		}
	}()
	return nil
}
