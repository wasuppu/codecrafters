package main

import (
	"encoding/binary"
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

		// body
		var body []byte
		apiVersion := binary.BigEndian.Uint16(buf[6:8])
		errorCode := byte(0)
		if apiVersion > 4 {
			errorCode = 35
		}
		body = append(body, 0, errorCode)

		elements := [3]struct {
			apiKey     byte
			minVersion byte
			maxVersion byte
			tagBuffer  byte
		}{
			{1, 0, 17, 0},
			{18, 0, 4, 0},
			{75, 0, 0, 0},
		}

		arrayLengh := byte(len(elements) + 1)
		body = append(body, arrayLengh)

		for _, elememt := range elements {
			body = append(body, 0, elememt.apiKey)
			body = append(body, 0, elememt.minVersion)
			body = append(body, 0, elememt.maxVersion)
			body = append(body, elememt.tagBuffer)
		}

		// throttle
		body = append(body, []byte{0, 0, 0, 0}...)

		// tag buffer
		body = append(body, 0)

		// write response
		messageSize := make([]byte, 4)
		binary.BigEndian.PutUint32(messageSize, uint32(4+len(body)))
		correlationId := buf[8:12]

		conn.Write(messageSize)
		conn.Write(correlationId)
		conn.Write(body)
	}
}
