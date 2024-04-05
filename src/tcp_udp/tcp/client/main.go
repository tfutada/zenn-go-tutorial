package main

import (
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8088")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// メッセージを送信する
	for {
		msg := "hello"
		conn.Write([]byte(msg))

		// sleep
		time.Sleep(1 * time.Second)
	}
}
