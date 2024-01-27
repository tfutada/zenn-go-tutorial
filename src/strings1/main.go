package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	fmt.Println("Hello, world!")

	r := strings.NewReader("abc") // returns a new Reader interface

	data := make([]byte, 10)
	count, err := r.Read(data)
	if err != nil {
		fmt.Errorf("err: %v", err)
	}

	fmt.Printf("%v, %#v \n", count, data)

	// rewind
	r.Seek(0, io.SeekStart)

	// read again
	data2 := make([]byte, 10)
	count2, err := r.Read(data2)
	if err != nil {
		fmt.Errorf("err: %v", err)
	}

	fmt.Printf("%v, %#v", count2, data2)
}

