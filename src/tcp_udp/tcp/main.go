package main

import (
	"fmt"
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
		defer conn.Close()

		println("connected")
		go handleconn(conn)
	}
}

func handleconn(conn net.Conn) {
	time.Sleep(10 * time.Second)

	buf := make([]byte, 1024)

	// this won't terminate. it will keep reading from the connection.
	for {
		n, _ := conn.Read(buf) // consume packets to a buffer. it is possible that multiple 'hello' messages are read at once.
		fmt.Println(string(buf[:n]))
		time.Sleep(3 * time.Second)
	}
}
