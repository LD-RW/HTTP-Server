package main

import (
	"fmt"
	"log"
	"net"

	"github.com/LD-RW/HTTPServer/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %v\n", r.RequestLine.Method)
		fmt.Printf("- Target: %v\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)

		})
	}

}
