package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		w.WriteStatusLine(response.StatusOK)
		headers := response.GetDefaultHeaders(0)
		headers.Remove("content-length")
		headers.Remove("connection")
		headers.Override("content-type", "application/json")
		headers.Set("transfer-encoding", "chunked")
		w.WriteHeaders(headers)
		binPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		res, err := http.Get("https://httpbin.org/" + binPath)
		if err != nil {
			log.Fatal(err)
		}
		buffer := make([]byte, 1024)

		for {
			n, err := res.Body.Read(buffer)
			if n > 0 {
				w.WriteChunkedBody(buffer[:n])
			}
			if err != nil {
				w.WriteChunkedBodyDone()
				return
			}
		}
	}

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
