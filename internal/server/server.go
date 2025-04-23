package server

import (
	"httpfromtcp/internal/response"
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
	content := "Hello World!"
	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	conn.Write([]byte(content))
	conn.Close()
}
