package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/LD-RW/HTTPServer/internal/headers"
)

// parserState defines the operational phases of the HTTP FSM.
type parserState string

/*
The parser implements a Finite State Machine (FSM) to handle the asynchronous
and streaming nature of TCP traffic. By maintaining state between reads, the
parser can gracefully handle "partial reads" where a single HTTP header or
line is split across multiple network packets.
*/
const (
	StateInit    parserState = "init"    // Initial state: seeking the Request-Line
	StateDone    parserState = "done"    // Terminal state: request successfully parsed
	StateHeaders parserState = "headers" // Intermediate state: consuming header fields
	StateBody    parserState = "body"    // Intermediate state: consuming message body
	stateError   parserState = "error"   // Failure state: protocol violation encountered
)

// Request represents a fully or partially parsed HTTP/1.1 request.
type Request struct {
	RequestLine RequestLine      // Parsed Method, URI, and Version
	Headers     *headers.Headers // Collection of parsed header fields
	state       parserState      // Current internal state of the FSM
	Body        string           // Accumulated message body content
}

// RequestLine represents the start-line of an HTTP request.
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// hasBody determines if the request should contain a message body based on
// the presence of Content-Length or Transfer-Encoding: chunked headers.
func (r Request) hasBody() bool {
	// Standard length-based body
	length := getInt(r.Headers, "content-length", 0)
	if length > 0 {
		return true
	}
	// Dynamic chunked-encoding body
	transferEncoding, exists := r.Headers.Get("transfer-encoding")
	if exists && strings.Contains(strings.ToLower(transferEncoding), "chunked") {
		return true
	}

	return false

}

// getInt safely retrieves a header value as an integer, returning a default
// if the header is missing or malformed.
func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	// String-to-integer conversion with graceful error fallback.
	value, err := strconv.Atoi(valueStr)
	// This handles the case where the header exists but isn't a number (e.g., Content-Length: "hello"
	if err != nil {
		return defaultValue
	}
	return value
}

// String-to-integer conversion with graceful error fallback.
func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

// ValidHTTP ensures the request targets the supported HTTP/1.1 protocol.
func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "HTTP/1.1"

}

var ErrorMalformedRequestLine = fmt.Errorf("malformed HTTP Request Line")
var ErrorRequestInErrorState = fmt.Errorf("request Line is in Error State")
var SEPARATOR = []byte("\r\n")

// parseRequestLine extracts the Method, Target, and Version from the start-line.
// It returns (n=0) if a complete CRLF-terminated line is not yet available in the buffer.
func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	// If no CRLF is found, the line is incomplete; wait for more data.
	if idx == -1 {
		return nil, 0, nil
	}
	/*
		startLine captures just the text (e.g., GET / HTTP/1.1) without the \r\n
	*/
	startLine := b[0:idx]
	read := idx + len(SEPARATOR) // Advance past the CRLF

	// HTTP/1.1 start-line must consist of exactly three space-delimited parts.
	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}
	httpParts := bytes.Split(parts[2], []byte("/"))

	// Validate the protocol format (e.g., HTTP/1.1).
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}
	/*
		Performance Note: Manual byte splitting is preferred over Regular Expressions (Regex)
		to minimize CPU overhead and heap allocations, ensuring high throughput.
	*/
	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	// We return the number of read bytes so we can know where headers exactly begin
	return rl, read, nil
}

// parse executes a single pass of the FSM over the provided data buffer.
// It returns the number of bytes consumed by the parser.
func (r *Request) parse(data []byte) (int, error) {
	/*
		Since data arrives in chunks, the read variable
		tracks how many bytes we have successfully processed in this specific call.
	*/
	read := 0
outer:
	for {
		/*
			currentData is a "view" (slice) that always points to the next byte we haven't looked at yet
			note : This was inspired from the sliding window technique that I have learned from competitive programming
		*/
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case stateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = stateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)

			if err != nil {
				r.state = stateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			/*
				In the real world I don't think we would get an EOF after reading data, therefore we can transition
				nicely to the body then to statedone, I'm doing the transition in here
			*/
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			// We need to check the case of chuncked encoding
			if length == 0 {
				te, isChunked := r.Headers.Get("transfer-encoding")
				if isChunked && strings.Contains(strings.ToLower(te), "chunked") {
					for {
						idx := bytes.Index(currentData, SEPARATOR)
						if idx == -1 {
							break outer
						}
						hexSize := string(currentData[:idx])
						size, err := strconv.ParseInt(hexSize, 16, 64)
						if err != nil {
							r.state = stateError
							return 0, err
						}
						if size == 0 {
							read += idx + len(SEPARATOR)
							r.state = StateDone
							break outer
						}
						totalChunkSize := int(size) + len(SEPARATOR)
						dataStart := idx + len(SEPARATOR)
						// ensure we have the full data chunk and the trailing \r\n
						if len(currentData[dataStart:]) < totalChunkSize {
							break outer
						}
						r.Body = string(currentData[dataStart : dataStart+int(size)])
						used := dataStart + totalChunkSize
						read += used
						currentData = currentData[used:]
					}
				}
			} else {
				/*
					This calculates exactly how many bytes to take from the current network buffer.
					length-len(r.Body): How many more bytes we actually need to finish the body.
					len(currentData): How many bytes are currently sitting in the buffer.
					This is better because Sometimes the network buffer contains
					more than just the body (like the start of a second HTTP request).
				*/
				remaining := min(length-len(r.Body), len(currentData))
				/*
					current data is actually a window from the current read data to the end of the buffer
				*/
				r.Body += string(currentData[:remaining])
				read += remaining
				if len(r.Body) == length {
					r.state = StateDone
				}
			}
		case StateDone:
			break outer

		default:
			panic("It seems I did something wrong :)")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == stateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	/*
		Design Choice: Using a fixed-size buffer is more memory-efficient
		than creating a new slice for every read. bufLen tracks exactly
		how much data is currently sitting in that bucket.
	*/
	request := newRequest()
	buf := make([]byte, 4096)
	bufLen := 0
	/*
		It keeps asking the network card for more bytes until the request
		state machine says it has everything it needs
	*/
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}
		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		/*
			Sliding Window approach
		*/
		copy(buf, buf[readN:bufLen])
		bufLen -= readN

	}
	return request, nil
	/*
		Why this is better than a simple ioutil.ReadAll ?
		-Stop-on-completion: It stops reading the moment the HTTP body is finished, even if the connection is still open
		-Memory Control: It never uses more than 4KB of memory for the raw buffer, regardless of how big the headers are.
		-Fragment Handling: It perfectly manages data that arrives split across multiple packets.
	*/
}
