package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

/*
Headers manages HTTP field-value pairs. While a custom struct or slice could offer
greater abstraction, a map provides O(1) lookups and is sufficient for this
implementation as long as the underlying storage is encapsulated.
*/

/*
Memory Management Strategy:
Byte slices are utilized throughout the parser to minimize memory allocations.
Unlike immutable strings, byte slices allow in-place manipulation and align
with the io.Reader/io.Writer interfaces provided by net.Conn, reducing
overhead during high-throughput network I/O.
*/
// isToken validates if a byte slice consists only of valid HTTP token characters
// as defined in RFC 7230.
func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true

		}
		if !found {
			return false
		}
	}
	return true
}

// Headers represents a collection of HTTP headers with case-insensitive keys.
type Headers struct {
	// The internal map uses lowercase keys to ensure case-insensitivity.
	headers map[string]string
}

// NewHeaders initializes an empty Headers container.
func NewHeaders() *Headers {
	return &Headers{
		// maps in go are implemented using hash tables
		headers: map[string]string{},
	}
}

// rn represents the CRLF (Carriage Return Line Feed) delimiter used in HTTP/1.1.
var rn = []byte("\r\n")

// parseHeader dissects a raw HTTP field line into a key-value pair.
// It enforces RFC 9112 compliance regarding field formatting and whitespace.
func parseHeader(fieldLine []byte) (string, string, error) {
	// SplitN is used with N=2 to ensure that colons within the field-value
	// (e.g., in a Host header with a port) do not cause incorrect fragmentation.
	parts := bytes.SplitN(fieldLine, []byte(":"), 2) // Can you have a ":" in the field value ?
	// Per RFC 9112, a valid header must contain a colon separator.
	if len(parts) != 2 {
		return "", "", errors.New("malformed field line")
	}
	name := parts[0]
	// Leading/trailing whitespace in field values is ignored per the HTTP specification.
	value := bytes.TrimSpace(parts[1])
	// Security: Prevent Request Smuggling. Field names must not contain leading
	// spaces. Ambiguous spacing can lead to header injection or smuggling attacks.
	if bytes.HasPrefix(name, []byte(" ")) { // localhost : 49532
		return "", "", errors.New("malformed field name")
	}
	return string(name), string(value), nil
}

// Get retrieves a header value by name (case-insensitive).
func (h *Headers) Get(name string) (string, bool) {
	str, ok := h.headers[strings.ToLower(name)]
	return str, ok
}

// Replace updates a header value or adds it if it doesn't exist.
func (h *Headers) Replace(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = value
}

// Delete removes a header from the collection.
func (h *Headers) Delete(name string) {
	name = strings.ToLower(name)
	delete(h.headers, name)
}

// Set adds a header value. If the header already exists, the value is
// appended with a comma as per HTTP multi-value header conventions.
func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

// ForEach iterates over all stored headers and executes the provided callback.
func (h *Headers) ForEach(cb func(n, v string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

// Parse scans a byte slice for HTTP headers, processing them line-by-line
// until an empty line (CRLF CRLF) is encountered.
func (h *Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}
		// An empty line (index 0 relative to current read) marks the end of the header block.
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		// Validate that the header name is a valid HTTP token.
		if !isToken([]byte(name)) {
			return 0, false, errors.New("malformed header")
		}

		read += idx + len(rn) // n + 2
		h.Set(name, value)
	}
	return read, done, nil

}
