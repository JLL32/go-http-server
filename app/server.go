package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
)

func main() {
	var filesDir string
	if len(os.Args) == 3 && os.Args[1] == "--directory" {
		filesDir = os.Args[2]
	}

	fmt.Println("Logs from your program will appear here!")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go serve(conn, filesDir)
	}

}

func serve(conn net.Conn, dir string) {
	defer conn.Close()

	req := readRequest(conn)

	switch req.path[0] {
	case "":
		fmt.Fprint(conn, "HTTP/1.1 200 Ok\r\n\r\n")

	case "echo":
		var content string
		if len(req.path) >= 2 {
			content = strings.Join(req.path[1:], "/")
		} else {
			content = ""
		}
		fmt.Fprint(conn, textResponse(content))

	case "user-agent":
		fmt.Fprint(conn, textResponse(req.headers["User-Agent"]))

	case "files":
		if len(req.path) < 2 {
			notFound(conn)
			return
		}

		if req.verb == "POST" {
			name := path.Join(req.path[1:]...)
			path := path.Join(dir, name)
			f, err := os.Create(path)
			if err != nil {
				notFound(conn)
				return
			}

			_, err = f.Write(req.body)
			if err != nil {
				notFound(conn)
				return
			}

			fmt.Fprint(conn, "HTTP/1.1 201 Created\r\n\r\n")
		} else {
			name := path.Join(req.path[1:]...)
			path := path.Join(dir, name)
			buff, err := os.ReadFile(path)
			if err != nil {
				notFound(conn)
				return
			}

			fmt.Fprint(conn, fileResponse(string(buff)))
		}

	default:
		notFound(conn)
	}
}

type request struct {
	verb    string
	path    []string
	version string
	headers map[string]string
	body    []byte
}

func readRequest(r io.Reader) request {
	var lines []string
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if string(line) == "\r\n" {
			break
		}

		lines = append(lines, strings.TrimSuffix(line, "\r\n"))
	}

	if len(lines) == 0 {
		return request{}
	}

	var req request

	head := strings.Split(lines[0], " ")

	req.verb = head[0]

	tokens := strings.Split(head[1], "/")
	for _, v := range tokens {
		if v != "" { // Split() split can result in empty tokens
			req.path = append(req.path, v)
		}
	}
	if len(req.path) == 0 {
		req.path = append(req.path, "")
	}

	req.version = head[2]
	if len(lines) == 1 {
		return req
	}

	req.headers = make(map[string]string)
	for i := 1; i < len(lines); i++ {
		pair := strings.Split(lines[i], ":")
		req.headers[pair[0]] = strings.TrimSpace(pair[1])
	}

	contentLen, err := strconv.Atoi(
		req.headers["Content-Length"])
	if err != nil || contentLen == 0 {
		return req
	}

	for ; contentLen != 0; contentLen-- {
		byte, err := reader.ReadByte()
		if err != nil {
			break
		}

		req.body = append(req.body, byte)
	}

	// fmt.Println(parsedRequest.verb)
	// fmt.Println(parsedRequest.path)
	// fmt.Println(parsedRequest.version)
	// fmt.Println(parsedRequest.headers)
	// fmt.Println(parsedRequest.body)
	return req
}

func textResponse(content string) string {
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
