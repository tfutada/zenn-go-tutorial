package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
	fmt.Println("cannot listen", err)	
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("cannot accet", err)	
	}

	// send
	str := "Hello, world"
	data := []byte(str)

	_, err = conn.Write(data)
	

	if err != nil {
		fmt.Println("cannot write", err)	
	}
}


