package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("error: ", err)
	}

	// read

	data := make([]byte, 1024)
	count, _ := conn.Read(data)
	fmt.Println(string(data[:count]))

}


