package fileserver

import (
	"crypto/tls"
	"encoding/gob"
	"file-sync/pkg/globalmodels"
	"fmt"
	"net"
	"server/pkg/cache"
	"server/services"
)

type FileServer interface {
	ListenAndServe(port int) error
}

type concreteFileServer struct {
	fileCache   cache.Cache
	metaCache   cache.Cache
	fileService services.FileService
	config      *tls.Config
}

func NewFileServer(fileService services.FileService, fileCache cache.Cache, metaCache cache.Cache, config *tls.Config) (FileServer, error) {
	return &concreteFileServer{
		fileCache,
		metaCache,
		fileService,
		config,
	}, nil
}

func (s *concreteFileServer) ListenAndServe(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()

	tlsListener := tls.NewListener(listener, s.config)
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

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var handshakeMessage globalmodels.HandshakeMessage
	// Handshake...
	err := decoder.Decode(&handshakeMessage)
	if err != nil {
		handleErrorResponse(encoder, globalmodels.InvalidMessage)
		return
	}
	// Receive the client's request...
	// Process the request...
	// Send the response...

	// After the handshake, continue with regular communication...
}

func handleErrorResponse(encoder *gob.Encoder, statusCode globalmodels.StatusCode) {
	encoder.Encode(globalmodels.Message{
		StatusCode: statusCode,
	})
}
