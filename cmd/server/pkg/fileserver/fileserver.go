package fileserver

import (
	"crypto/tls"
	"fmt"
	"net"
	"server/pkg/mux"
)

type Server interface {
	ListenAndServe(port int) error
}

type concreteServer struct {
	mux    mux.Mux
	config *tls.Config
}

func NewServer(mux mux.Mux, config *tls.Config) Server {
	return &concreteServer{
		mux,
		config,
	}
}

func (s *concreteServer) ListenAndServe(port int) error {
	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer tcpListener.Close()

	tlsListener := tls.NewListener(tcpListener, s.config)
	defer tlsListener.Close()

	for {
		var conn net.Conn
		conn, err = tlsListener.Accept()
		if err != nil {
			return err
		}
		go s.mux.ServeConn(conn)
	}
}
