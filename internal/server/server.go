package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/LD-RW/HTTPServer/internal/response"
)

type Server struct {
	closed bool
}

func runConnection(_s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	headers := response.GetDefaultHeaders(0)
	response.WriteStatusLine(conn, response.StatusOk)
	response.WriteHeaders(conn, headers)
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
		go runConnection(s, conn)
	}
}

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{closed: false}
	go runServer(server, listener)
	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
