package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

var separator = []byte(":")
var crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, []byte(crlf))
	if crlfIdx == -1 {
		return n, false, nil
	}

	if crlfIdx == 0 {
		return n, true, nil
	}

	name, value, found := bytes.Cut(data[:crlfIdx], separator)
	if !found {
		return 0, false, fmt.Errorf("invalid header: %v", string(data))
	}
	if unicode.IsSpace(rune(name[len(name)-1])) {
		return 0, false, fmt.Errorf("header name ends in whitespace: %v", string(name))
	}

	h[strings.TrimSpace(string(name))] = strings.TrimSpace(string(value))
	n = crlfIdx + len(crlf)
	
	return n, false, nil
}
