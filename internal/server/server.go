package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	portString := ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", portString)
	if err != nil {
		return nil, err
	}

	server := &Server{listener: l}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Fatalf("Error accepting connection: %v", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	conn.Write([]byte(`HTTP/1.1 200 OK
Content-Type: text/plain

Hello World!`))
	conn.Close()
}
