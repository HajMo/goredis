package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer listener.Close()

	f, err := os.Create("storage")
	if err != nil {
		fmt.Println("Failed to create storage file")
		os.Exit(1)
	}
	f.Close()

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
		case "set":
			if len(args) != 2 {
				fmt.Println("Error handling set command")
				connection.Write([]byte("-ERR handling command: '" + command + "'\r\n"))
				return
			}
			if _, err := handleSetCommand(args[0], args[1]); err != nil {
				fmt.Println("Error handling set command: ", err.Error())
				connection.Write([]byte("-ERR handling command: '" + command + "'\r\n"))
				os.Exit(1)
			}
			connection.Write([]byte("+OK\r\n"))
		case "get":
			value, _ := handleGetCommand(args[0].String())
			connection.Write([]byte("+" + value + "\r\n"))
		default:
			connection.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}

func handleSetCommand(key Value, value Value) (int, error) {
	file, err := os.OpenFile("storage", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	n, err := file.WriteString(fmt.Sprintf("%s:%s\n", key.String(), value.String()))
	if err != nil {
		panic(err)
	}

	return n, err
}

func handleGetCommand(key string) (string, error) {
	file, err := os.OpenFile("storage", os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	values := []string{}

	for scanner.Scan() {
		scannedKey := strings.Split(scanner.Text(), ":")
		if strings.Contains(scannedKey[0], key) && len(scannedKey[0]) == len(key) {
			values = append(values, scanner.Text())
		}
	}

	if len(values) == 0 {
		return "", nil
	}

	return strings.Split(values[len(values)-1], ":")[1], nil
}
