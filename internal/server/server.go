package server

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	closed bool
}

func runConnection(s *Server, conn io.ReadWriteCloser) {

	out := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nContent-Length: 13\n\nHello World!")
	conn.Write(out)
	conn.Close()
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
