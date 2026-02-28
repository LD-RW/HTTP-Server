package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LD-RW/HTTPServer/internal/request"
	"github.com/LD-RW/HTTPServer/internal/response"
	"github.com/LD-RW/HTTPServer/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {

		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem is not my problem\n",
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Woopsie, my bad\n",
			}
		} else {
			w.Write([]byte("All good, frfr\n"))

		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error starting s: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
