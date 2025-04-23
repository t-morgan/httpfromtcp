package server

import (
	"bytes"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError
type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	portString := ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", portString)
	if err != nil {
		return nil, err
	}

	server := &Server{listener: l, handler: handler}
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
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		writeError(conn, &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Bad Request",
		})
		return
	}

	buffer := bytes.Buffer{}

	handlerError := s.handler(&buffer, req)
	if handlerError != nil {
		writeError(conn, handlerError)
		return
	}

	responseBody := buffer.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(len(responseBody)))
	conn.Write([]byte(responseBody))
}

func writeError(conn io.Writer, handlerError *HandlerError) error {
	err := response.WriteStatusLine(conn, handlerError.StatusCode)
	if err != nil {
		return err
	}

	headers := headers.NewHeaders()
	headers.Set("Content-Type", "text/plain")
	headers.Set("Content-Length", strconv.Itoa(len(handlerError.Message)))

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(handlerError.Message))
	return err
}
