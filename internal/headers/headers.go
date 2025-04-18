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
var validNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%'*+-.^_`|~"

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

	nameKey := strings.ToLower(strings.TrimSpace(string(name)))
	for i, b := range nameKey {
		if !contains(validNameChars, byte(b)) {
			return 0, false, fmt.Errorf("invalid character '%c' (0x%x) at index %d", b, b, i)
		}
	}

	h[nameKey] = strings.TrimSpace(string(value))
	n = crlfIdx + len(crlf)
	
	return n, false, nil
}

func contains(s string, b byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return true
		}
	}
	return false
}