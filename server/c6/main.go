package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handlConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	conn.Read(buf)
	req := string(buf)

	headers := strings.Split(req, "\r\n")
	path := strings.TrimSpace(headers[0])
	path = strings.Split(path, " ")[1]

	agent := ""
	for _, msg := range headers {
		if strings.HasPrefix(msg, "User-Agent") {
			agent = strings.Split(msg, ": ")[1]
			break
		}
	}

	resp := ""
	if strings.HasPrefix(req, "GET / HTTP/1.1") {
		resp = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.Contains(req, "/echo/") {
		msg := strings.TrimPrefix(path, "/echo/")
		resp = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)
	} else if strings.Contains(req, "/user-agent") {
		resp = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), agent)
	} else if strings.Contains(req, "/files/") {
		dir := os.Args[2]
		fileName := strings.TrimPrefix(path, "/files/")
		data, err := os.ReadFile(dir + fileName)
		if err != nil {
			resp = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			resp = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
		}
	} else {
		resp = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	conn.Write([]byte(resp))
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		handlConnection(conn)
	}
}
