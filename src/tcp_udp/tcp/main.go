package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// https://qiita.com/shirakiyo/items/eaffa7353a1a5114a09d
func main() {
	listener, err := net.Listen("tcp", "localhost:8088")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		println("connected")
		go handleconn(conn)
	}
}

func handleconn(conn net.Conn) {
	defer conn.Close() // Ensure the connection is closed when this function exits

	time.Sleep(10 * time.Second)

	buf := make([]byte, 1024)

	// this won't terminate. it will keep reading from the connection.
	for {
		n, err := conn.Read(buf) // consume packets to a buffer. it is possible that multiple 'hello' messages are read at once.
		if err != nil {
			if err == io.EOF {
				// The client closed the connection gracefully
				fmt.Println("Client has closed the connection")
			} else {
				// An actual error occurred
				fmt.Println("Error reading:", err)
			}
			break // Exit the loop either way
		}

		fmt.Println(string(buf[:n]))
		time.Sleep(3 * time.Second)
	}
}
