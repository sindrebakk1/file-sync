package fileserver

import (
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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
	defer s.mux.Shutdown()

	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer tcpListener.Close()

	tlsListener := tls.NewListener(tcpListener, s.config)
	defer tlsListener.Close()

	log.Infof("Listening on port %d", port)

	for {
		var conn net.Conn
		conn, err = tlsListener.Accept()
		log.Debugf("Accepted connection from %s, port: %d", conn.RemoteAddr().String(), conn.RemoteAddr().(*net.TCPAddr).Port)
		conn.RemoteAddr().String()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				log.Info("Timed out", err)
				continue
			}
			if conn != nil {
				conn.Close()
			}
			return err
		}

		// Check if the accepted connection's local address port matches the intended port
		if conn.RemoteAddr().(*net.TCPAddr).Port != port {
			log.Debugf("Connection from unexpected port: %d, closing...", conn.LocalAddr().(*net.TCPAddr).Port)
			conn.Close()
			continue
		}

		go s.mux.ServeConn(conn)
	}
}
