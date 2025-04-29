package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteStatusLine(response.StatusBadRequest)
		headers := response.GetDefaultHeaders(len(server.BadRequestHTML))
		headers.SetContentType("text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(server.BadRequestHTML))
	case "/myproblem":
		w.WriteStatusLine(response.StatusInternalServerError)
		headers := response.GetDefaultHeaders(len(server.ServerErrorHTML))
		headers.SetContentType("text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(server.ServerErrorHTML))
	default:
		w.WriteStatusLine(response.StatusOK)
		headers := response.GetDefaultHeaders(len(server.SuccessHTML))
		headers.SetContentType("text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(server.SuccessHTML))
	}
}
