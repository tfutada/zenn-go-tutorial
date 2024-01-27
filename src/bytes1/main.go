package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	b.Write([]byte("foo"))

	data := make([]byte, 10)
	o, err := b.Read(data)
	if err != nil {
		fmt.Println(err)	
	}

	fmt.Printf("%#v %#v", o, data)
}
