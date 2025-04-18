package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type parserState int

const (
	Initialized parserState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state: Initialized,
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

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == Initialized {
		requestLine, byteCount, err := parseRequestLine(string(data))
		if err != nil || byteCount == 0 {
			return byteCount, err
		}
		r.RequestLine = *requestLine
		r.state = Done
		return byteCount, nil
	}

	if r.state == Done {
		return 0, errors.New("trying to read data in a done state")
	}

	return 0, fmt.Errorf("unknown state")
}

func parseRequestLine(message string) (*RequestLine, int, error) {
	if strings.Index(message, "\r\n") == -1 {
		return nil, 0, nil
	}
	messageParts := strings.Split(message, "\r\n")
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
	}, len(message), nil
}
