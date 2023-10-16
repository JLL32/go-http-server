package main

import (
	"bufio"
	"fmt"
	"regexp"
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
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	//
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(conn)
	if ok := scanner.Scan(); ok {
		head := strings.Split(scanner.Text(), " ")

		if head[1] == "/" {
			fmt.Fprint(conn, "HTTP/1.1 200 Ok\r\n\r\n")
		} else if ok, _ := regexp.Match("/echo*", []byte(head[1])); ok {
			path := strings.Split(head[1][1:], "/")
			if len(path) == 2 {
				fmt.Fprint(conn, contentResponse(path[1]))
			} else {
				fmt.Fprint(conn, contentResponse(""))
			}
		} else {
			fmt.Fprint(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		}
	}
}

func contentResponse(content string) string {
	res := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		fmt.Sprintf("Content-Length: %v\r\n", len(content)) +
		"\r\n" +
		content

	return res
}
