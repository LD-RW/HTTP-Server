package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "HTTP/1.1"

}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("Malformed HTTP Request Line")
var SEPARATOR = "\r\n"

func parseRequestLine(b string) (*RequestLine, string, error) {
	idx := strings.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, b, nil
	}
	startLine := b[0:idx]
	restOfMsg := b[idx+len(SEPARATOR):]

	parts := strings.Split(startLine, " ")

	if len(parts) != 3 {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}
	httpParts := strings.Split(parts[2], "/")

	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpParts[1],
	}
	return rl, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Error reading request body: "), err)
	}
	str := string(data)
	rl, _, err := parseRequestLine(str)
	if err != nil {
		return nil, err
	}
	return &Request{
		RequestLine: *rl,
	}, err

}
