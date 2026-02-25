package headers

import (
	"bytes"
	"errors"
)

/*
	Headers : I thought about not using a map, this will help me with more

abstraction, but as-long-as I'm not changing the implementation, it's ok
*/
type Headers map[string]string

var rn = []byte("\r\n")

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2) // Can you have a ":" in the field value ?
	if len(parts) != 2 {
		return "", "", errors.New("malformed field line")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if bytes.HasPrefix(name, []byte(" ")) { // localhost : 49532
		return "", "", errors.New("malformed field name")
	}
	return string(name), string(value), nil
}

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}
		// empty header
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
		//fmt.Printf("\"%s\"\n", string(data[read:read+idx]))
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		read += idx + len(rn) // n + 2
		h[name] = value       // a map is already a pointer, so I should be able to do that ?
	}
	return read, done, nil

}
