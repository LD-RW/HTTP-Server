package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/LD-RW/HTTPServer/internal/request"
	"github.com/LD-RW/HTTPServer/internal/response"
)

type Server struct {
	// When shutdown, this will be used to stop the runServer function
	closed bool
	// The server uses this to know what to do once a request is successfully parsed
	handler Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

/*
This allows the server package to remain "agnostic".
It doesn't care if you are building an Amazon clone
or a simple calculator; as long as your
function matches this signature, the server can run it.
*/
type Handler func(w *response.Writer, req *request.Request)

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, r)

}

func runServer(s *Server, listener net.Listener) error {

	for {
		conn, err := listener.Accept()
		if s.closed {
			return nil
		}
		if err != nil {
			log.Println("Error accepting connection", err)
			continue
		}
		// So you can serve multiple connections (or clients) in parallel
		go runConnection(s, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		closed:  false,
		handler: handler,
	}
	go runServer(server, listener)
	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
