package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:9092")
	if err != nil {
		fmt.Println("Failed to bind to port 9092")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection: ", err)
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Failed to read from connection:", err)
			return
		}

		messageSize := make([]byte, 4)
		conn.Write(messageSize)

		correlationId := byte(7) // hard-coded correlation_id 7
		conn.Write([]byte{0, 0, 0, correlationId})
	}
}
