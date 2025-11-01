package main

import (
	"bytes"
	"compress/gzip"
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

	hds := strings.Split(req, "\r\n")
	hdsp := make(map[string]string)

	firstLine := strings.TrimSpace(hds[0])
	body := hds[len(hds)-1]
	m := strings.Split(firstLine, " ")[0]
	p := strings.Split(firstLine, " ")[1]

	for _, hd := range hds[1 : len(hds)-1] {
		parts := strings.Split(hd, ": ")
		key := strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			hdsp[key] = parts[1]
		}
	}

	resp := ""

	if strings.HasPrefix(req, "GET / HTTP/1.1") {
		resp = "HTTP/1.1 200 OK\r\n\r\n"
	} else if m == "GET" && strings.Contains(p, "/echo") {
		extra := ""
		msg := strings.TrimPrefix(p, "/echo/")
		encodings, ok := hdsp["Accept-Encoding"]
		if ok && strings.Contains(encodings, "gzip") {
			var b bytes.Buffer
			w := gzip.NewWriter(&b)
			w.Write([]byte(msg))
			w.Close()
			resp = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n" + extra +
				fmt.Sprintf("Content-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(b.String()), b.String())
		} else {
			resp = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n" +
				fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(msg), msg)
		}
	} else if m == "GET" && strings.Contains(p, "/user-agent") {
		agent := hdsp["User-Agent"]
		resp = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), agent)
	} else if m == "GET" && strings.Contains(req, "/files/") {
		dir := os.Args[2]
		fileName := strings.TrimPrefix(p, "/files/")
		data, err := os.ReadFile(dir + fileName)
		if err != nil {
			resp = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			resp = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
		}
	} else if m == "POST" && strings.Contains(req, "/files/") {
		dir := os.Args[2]
		content := strings.Trim(body, "\x00")
		fileName := strings.TrimPrefix(p, "/files/")
		file, _ := os.Create(dir + fileName)
		file.WriteString(content)
		resp = "HTTP/1.1 201 Created\r\n\r\n"
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
