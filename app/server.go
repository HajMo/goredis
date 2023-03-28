package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	for {
		value, err := Decode(bufio.NewReader(connection))
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("Error decoding RESP: ", err.Error())
			return // Ignore clients that we fail to read from
		}

		command := value.Array()[0].String()
		args := value.Array()[1:]

		switch command {
		case "ping":
			connection.Write([]byte("+PONG\r\n"))
		case "echo":
			connection.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[0].String()), args[0].String())))
		default:
			connection.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}
