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
			continue
		}

		go func() {
			defer conn.Close()

			var buff []string
			sc := bufio.NewScanner(conn)
			for sc.Scan() {
				line := sc.Text()
				if line == "" {
					break
				}
				buff = append(buff, line)
			}

			req := parseRequest(buff)

			if len(req.path) == 0 {
				fmt.Fprint(conn, "HTTP/1.1 200 Ok\r\n\r\n")
				return
			}

			switch req.path[0] {
			case "echo":
				var content string
				if len(req.path) >= 2 {
					content = strings.Join(req.path[1:], "/")
				} else {
					content = ""
				}
				fmt.Fprint(conn, contentResponse(content))

			case "user-agent":
				fmt.Fprint(conn, contentResponse(req.headers["User-Agent"]))

			case "files":
				if len(req.path) < 2 {
					notFound(conn)
					return
				}

				name := strings.Join(req.path[1:], "/")
				buff, err := os.ReadFile(name)
				if err != nil {
					notFound(conn)
					return
				}
				fmt.Fprint(conn, fileResponse(string(buff)))

			default:
				notFound(conn)
			}
		}()
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
	tokens := strings.Split(head[1], "/")
	for _, v := range tokens {
		if v != "" {
			parsedRequest.path = append(parsedRequest.path, v)
		}
	}
	parsedRequest.version = head[2]
	if len(req) == 1 || len(req) == 2 {
		return parsedRequest
	}

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

func fileResponse(fileContent string) string {
	res := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: application/octet-stream\r\n" +
		fmt.Sprintf("Content-Length: %v\r\n", len(fileContent)) +
		"\r\n" +
		fileContent

	return res
}

func notFound(w io.Writer) {
	fmt.Fprint(w, "HTTP/1.1 404 Not Found\r\n\r\n")
}
