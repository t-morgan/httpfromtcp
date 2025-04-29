package response

import (
	"errors"
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

type WriterStatus int

const (
	pendingStatusLine = iota
	pendingHeaders
	pendingBody
	done
)

type Writer struct {
	writer      io.Writer
	writerState WriterStatus
}

func New(w io.Writer) *Writer {
	return &Writer{
		writer:      w,
		writerState: pendingStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != pendingStatusLine {
		return errors.New("status line already written")
	}
	switch statusCode {
	case StatusOK:
		_, err := w.writer.Write([]byte("HTTP/1.1 200 OK"))
		if err != nil {
			return err
		}
	case StatusBadRequest:
		_, err := w.writer.Write([]byte("HTTP/1.1 400 Bad Request"))
		if err != nil {
			return err
		}
	case StatusInternalServerError:
		_, err := w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error"))
		if err != nil {
			return err
		}
	default:
		status := strconv.Itoa(int(statusCode))
		_, err := w.writer.Write([]byte("HTTP/1.1 " + status + " "))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.writerState = pendingHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")
	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != pendingHeaders {
		return errors.New("headers already written or not ready yet")
	}
	for key, value := range headers {
		_, err := w.writer.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.writerState = pendingBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != pendingBody {
		return 0, errors.New("body already written or not ready yet")
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	w.writerState = done
	return n, nil
}
