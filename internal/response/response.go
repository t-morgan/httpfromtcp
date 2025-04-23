package response

import (
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		w.Write([]byte("HTTP/1.1 200 OK"))
	case StatusBadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request"))
	case StatusInternalServerError:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error"))
	default:
		status := strconv.Itoa(int(statusCode))
		w.Write([]byte("HTTP/1.1 " + status + " "))
	}
	w.Write([]byte("\r\n"))
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		w.Write([]byte(key + ": " + value + "\r\n"))
	}
	w.Write([]byte("\r\n"))
	return nil
}
