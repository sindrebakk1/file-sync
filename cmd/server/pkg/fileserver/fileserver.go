package fileserver

import (
	"crypto/tls"
	"fmt"
	"net"
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
		authenticator := auth.NewAuthenticator(conn, s.userService, s.authConfig)
		handler := newClientHandler(conn, authenticator)
		go handler.handleClient()
	}
}
