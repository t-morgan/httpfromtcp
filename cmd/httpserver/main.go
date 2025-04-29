package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/headers"
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
		binPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		res, err := http.Get("https://httpbin.org/" + binPath)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteStatusLine(response.StatusOK)
		h := response.GetDefaultHeaders(0)
		h.Remove("content-length")
		h.Remove("connection")
		h.Override("content-type", res.Header.Get("Content-Type"))
		h.Set("transfer-encoding", "chunked")
		h.Set("trailer", "x-content-sha256, x-content-length")
		w.WriteHeaders(h)

		responseBody := []byte{}
		buffer := make([]byte, 1024)

		for {
			n, err := res.Body.Read(buffer)
			if n > 0 {
				w.WriteChunkedBody(buffer[:n])
				responseBody = append(responseBody, buffer[:n]...)
			}
			if err != nil {
				w.WriteChunkedBodyDone()
				break
			}
		}
		trailers := headers.NewHeaders()
		sum := sha256.Sum256(responseBody)
		trailers.Set("x-content-sha256", fmt.Sprintf("%x", sum))
		trailers.Set("x-content-length", fmt.Sprintf("%d", len(responseBody)))
		w.WriteTrailers(trailers)
		return
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
	case "/video":
		data, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Fatal(err)
		}
		w.WriteStatusLine(response.StatusOK)
		headers := response.GetDefaultHeaders(len(data))
		headers.SetContentType("video/mp4")
		w.WriteHeaders(headers)
		w.WriteBody(data)
	default:
		w.WriteStatusLine(response.StatusOK)
		headers := response.GetDefaultHeaders(len(server.SuccessHTML))
		headers.SetContentType("text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(server.SuccessHTML))
	}
}
