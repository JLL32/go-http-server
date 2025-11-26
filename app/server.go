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
	// defer conn.Close()

	req := readRequest(conn)

	respHeaders := make(map[string]string)

	if v, ok := req.headers["Accept-Encoding"]; ok {
		encodings := strings.Split(v, ",")
		for _, enco := range encodings {
			if strings.TrimSpace(enco) == "gzip" {
				respHeaders["Content-Encoding"] = "gzip"
				break
			}
		}
	}

	switch req.path[0] {
	case "":
		resp := NewResponse(200)
		fmt.Fprintf(conn, resp.String())

	case "echo":
		var content string
		if len(req.path) >= 2 {
			content = strings.Join(req.path[1:], "/")
		} else {
			content = ""
		}
		fmt.Fprint(conn, textResponse(respHeaders, content))

	case "user-agent":
		fmt.Fprint(conn, textResponse(respHeaders, req.headers["User-Agent"]))

	case "files":
		if len(req.path) < 2 {
			notFound(respHeaders, conn)
			return
		}

		if req.verb == "POST" {
			name := path.Join(req.path[1:]...)
			path := path.Join(dir, name)
			f, err := os.Create(path)
			if err != nil {
				notFound(respHeaders, conn)
				return
			}

			_, err = f.Write(req.body)
			if err != nil {
				notFound(respHeaders, conn)
				return
			}

			resp := NewResponse(201)
			fmt.Fprintf(conn, resp.String())
		} else {
			name := path.Join(req.path[1:]...)
			path := path.Join(dir, name)
			buff, err := os.ReadFile(path)
			if err != nil {
				notFound(respHeaders, conn)
				return
			}

			fmt.Fprint(conn, fileResponse(respHeaders, string(buff)))
		}

	default:
		notFound(respHeaders, conn)
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

type Response struct {
	Status  int
	Headers map[string]string
	Body    string
}

func NewResponse(status int) Response {
	return Response{
		Status:  status,
		Headers: make(map[string]string),
	}
}

func (resp *Response) String() string {
	var (
		respString string = fmt.Sprintf("HTTP/1.1 %v %v\r\n", resp.Status, statusText(resp.Status))
		headers    []string
	)

	for k, v := range resp.Headers {
		headers = append(headers, k+": "+v)
	}

	if len(headers) > 0 {
		respString += strings.Join(headers, "\r\n")
		respString += "\r\n"
	}
	respString += "\r\n"

	if len(resp.Body) > 0 {
		respString += resp.Body
	}

	return respString
}

func statusText(status int) string {
	switch status {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 404:
		return "Not Found"
	}
	return ""
}

func textResponse(defaultHeaders map[string]string, content string) string {
	resp := NewResponse(200)

	for k, v := range defaultHeaders {
		resp.Headers[k] = v
	}
	resp.Headers["Content-Type"] = "text/plain"
	resp.Headers["Content-Length"] = strconv.Itoa(len(content))

	resp.Body = content

	return resp.String()
}

func fileResponse(defaultHeaders map[string]string, fileContent string) string {
	resp := NewResponse(200)

	for k, v := range defaultHeaders {
		resp.Headers[k] = v
	}
	resp.Headers["Content-Type"] = "application/octet-stream"
	resp.Headers["Content-Length"] = strconv.Itoa(len(fileContent))

	resp.Body = fileContent

	return resp.String()
}

func notFound(defaultHeaders map[string]string, w io.Writer) {
	resp := NewResponse(404)

	for k, v := range defaultHeaders {
		resp.Headers[k] = v
	}

	w.Write([]byte(resp.String()))
}
