package fileserver

import (
	"crypto/tls"
	"encoding/gob"
	"file-sync/pkg/globalenums"
	"file-sync/pkg/globalmodels"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"server/pkg/auth"
	"server/pkg/cache"
	"server/services"
)

type FileServer interface {
	ListenAndServe(port int) error
}

type concreteFileServer struct {
	userService services.UserService
	fileCache   cache.Cache
	metaCache   cache.Cache
	tlsConfig   *tls.Config
	authConfig  *auth.Config
}

func NewFileServer(
	userService services.UserService,
	fileCache cache.Cache,
	metaCache cache.Cache,
	tlsConfig *tls.Config,
	authConfig *auth.Config,
) (FileServer, error) {
	return &concreteFileServer{
		userService,
		fileCache,
		metaCache,
		tlsConfig,
		authConfig,
	}, nil
}

func (s *concreteFileServer) ListenAndServe(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()

	tlsListener := tls.NewListener(listener, s.tlsConfig)
	defer tlsListener.Close()

	for {
		conn, connErr := tlsListener.Accept()
		if connErr != nil {
			return connErr
		}
		go s.handleClient(conn)
	}
}

func (s *concreteFileServer) handleClient(conn net.Conn) {
	defer conn.Close()

	// Create a channel to receive interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	// 1. Authenticate client
	authenticator := auth.NewAuthenticator(encoder, decoder, s.userService, s.authConfig)
	userName, err := authenticator.AuthenticateClient()
	if err != nil {
		log.Error(err)
		return
	}
	// 2. Initialize file service based on user
	//var fileService services.FileService
	_, err = s.userService.GetFileService(userName)
	// 3. Handle requests until interrupt signal is sent
	for {
		var request globalmodels.Message
		err = decoder.Decode(&request)
		if err != nil {
			log.Error("Error decoding request: ", err)
			return
		}

		fmt.Printf("Received request: %v\n", request)

		// Check for interrupt signal
		select {
		case <-interrupt:
			fmt.Println("Received interrupt signal. Shutting down gracefully.")
			return
		default:
			// Continue handling client requests
		}
	}
}

func handleErrorResponse(encoder *gob.Encoder, statusCode globalenums.StatusCode) {
	encoder.Encode(globalmodels.Message{
		StatusCode: statusCode,
	})
}
