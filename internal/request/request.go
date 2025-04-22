package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type parserState int

const (
	Initialized parserState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		Headers: headers.NewHeaders(),
		state:   Initialized,
	}

	buffer := make([]byte, 8)
	readToIndex := 0
	for request.state == Initialized {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		bytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = Done
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return request, err
		}

		copy(buffer, buffer[bytesParsed:readToIndex])
		readToIndex -= bytesParsed
	}

	idx := bytes.LastIndex(buffer, []byte(crlf))
	if idx != -1 {
		buffer = buffer[idx+2:]
	} else {
		buffer = make([]byte, 8)
	}
	for request.state != Done {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		bytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = Done
				contentLengthHeader := request.Headers["content-length"]
				if contentLengthHeader == "" {
					break
				}
				contentLength, err := strconv.Atoi(contentLengthHeader)
				if err != nil {
					return nil, err
				}
				if len(request.Body) < contentLength {
					return nil, fmt.Errorf("body shorter than content-length")
				}
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return request, err
		}

		copy(buffer, buffer[bytesParsed:readToIndex])
		readToIndex -= bytesParsed
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case Initialized:
		requestLine, byteCount, err := parseRequestLine(string(data))
		if err != nil || byteCount == 0 {
			return byteCount, err
		}
		r.RequestLine = *requestLine
		r.state = ParsingHeaders
		return byteCount, nil
	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return n, err
		}
		if done {
			r.state = ParsingBody
		}
		return n, nil
	case ParsingBody:
		if r.Headers["content-length"] == "" {
			r.state = Done
			return 0, nil
		}
		r.Body = append(r.Body, data...)
		contentLength, err := strconv.Atoi(r.Headers["content-length"])
		if err != nil {
			return 0, err
		}
		if len(r.Body) > contentLength {
			return len(data), fmt.Errorf("body is longer than content-length")
		}
		if len(r.Body) == contentLength {
			r.state = Done
		}
		return len(data), nil
	case Done:
		return 0, fmt.Errorf("trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}

func parseRequestLine(message string) (*RequestLine, int, error) {
	crlfIdx := strings.Index(message, crlf)
	if crlfIdx == -1 {
		return nil, 0, nil
	}
	messageParts := strings.Split(message, crlf)
	if len(messageParts) < 3 {
		return nil, 0, nil
	}

	requestLine := messageParts[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return nil, 0, errors.New("Request line has too many parts")
	}

	if parts[0] != strings.ToUpper(parts[0]) {
		return nil, 0, errors.New("Request method must be uppercase")
	}

	if parts[2] != "HTTP/1.1" {
		return nil, 0, errors.New("Request not supported")
	}

	httpVersion := strings.Split(parts[2], "/")[1]

	return &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: parts[1],
		Method:        parts[0],
	}, crlfIdx + 2, nil
}
