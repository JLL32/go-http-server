package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	//

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		var buff []string
		sc := bufio.NewScanner(conn)
		for sc.Scan() {
			line := sc.Text()
			// length := len(line)
			if line == "" {
				break
			}
			buff = append(buff, line)
		}

		// req := parseRequest(strings.Split(string(buff), "\r\n"))
		req := parseRequest(buff)

		// log.Println("hello")

		if len(req.path) == 1 {
			fmt.Fprint(conn, "HTTP/1.1 200 Ok\r\n\r\n")
		} else if req.path[1] == "echo" {
			if len(req.path) >= 3 {
				fmt.Fprint(conn, contentResponse(strings.Join(req.path[2:], "/")))
			} else {
				fmt.Fprint(conn, contentResponse(""))
			}
		} else if req.path[1] == "user-agent" {
			fmt.Fprint(conn, contentResponse(req.headers["User-Agent"]))
		} else {
			fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		}

		// log.Println("hello")

		conn.Close()
	}

}

type request struct {
	verb    string
	path    []string
	version string
	headers map[string]string
	body    bytes.Buffer
}

func parseRequest(req []string) request {
	if len(req) == 0 {
		return request{}
	}

	var parsedRequest request
	head := strings.Split(req[0], " ")

	parsedRequest.verb = head[0]
	parsedRequest.path = strings.Split(head[1], "/")
	parsedRequest.version = head[2]
	if len(req) == 1 || len(req) == 2 {
		return parsedRequest
	}

	// log.Println("hello")
	i := 2
	parsedRequest.headers = make(map[string]string)
	for ; i < len(req) && req[i] != ""; i++ {
		pair := strings.Split(req[i], ":")
		parsedRequest.headers[pair[0]] = strings.TrimSpace(pair[1])
	}

	for _, line := range req[i:] {
		io.WriteString(&parsedRequest.body, line)
	}

	// fmt.Println(parsedRequest.verb)
	// fmt.Println(parsedRequest.path)
	// fmt.Println(parsedRequest.version)
	// fmt.Println(parsedRequest.headers)
	// fmt.Println(parsedRequest.body)
	return parsedRequest
}

func contentResponse(content string) string {
	res := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		fmt.Sprintf("Content-Length: %v\r\n", len(content)) +
		"\r\n" +
		content

	return res
}
