// udp_peer.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	localAddr := ":9001"
	//remoteAddr := "127.0.0.1:9002"

	conn, err := net.ListenPacket("udp", localAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var dest *net.UDPAddr

	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, _ := conn.ReadFrom(buf)
			dest = addr.(*net.UDPAddr)

			fmt.Printf("\n[From %s] %s\n", addr, string(buf[:n]))
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("You: ")
		text, _ := reader.ReadString('\n')
		conn.WriteTo([]byte(text), dest)
	}
}
