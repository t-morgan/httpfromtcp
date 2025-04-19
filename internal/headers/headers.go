package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

var separator = []byte(":")

const crlf = "\r\n"
const validNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%'*+-.^_`|~"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, []byte(crlf))
	if crlfIdx == -1 {
		return n, false, nil
	}

	if crlfIdx == 0 {
		return 2, true, nil
	}

	name, value, found := bytes.Cut(data[:crlfIdx], separator)
	if !found {
		return 0, false, fmt.Errorf("invalid header: %v", string(data))
	}

	nameKey := strings.ToLower(string(name))
	if nameKey != strings.TrimRight(nameKey, " ") {
		return 0, false, fmt.Errorf("header name ends in whitespace: %v", nameKey)
	}

	nameKey = strings.TrimSpace(nameKey)
	for _, b := range nameKey {
		if !contains(validNameChars, byte(b)) {
			return 0, false, fmt.Errorf("invalid header token found: %v", nameKey)
		}
	}

	h.Set(nameKey, strings.TrimSpace(string(value)))
	return crlfIdx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

func contains(s string, b byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return true
		}
	}
	return false
}
