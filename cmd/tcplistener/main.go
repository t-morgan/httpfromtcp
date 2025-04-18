package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection accepted!")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", request.RequestLine.Method)
		fmt.Printf("- Target: %v\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", request.RequestLine.HttpVersion)


		fmt.Println("Connection closed!")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer f.Close()
		defer close(lines)

		buffer := make([]byte, 8)
		lineContents := ""

		for {
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- lineContents + parts[i]
				lineContents = ""
			}
			lineContents += parts[len(parts)-1]
		}
		if lineContents != "" {
			lines <- lineContents
		}
	}()
	return lines
}
