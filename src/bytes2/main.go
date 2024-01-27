package main

import (
	"bytes"
	"fmt"
)

func main() {
	var b bytes.Buffer
	b.Write([]byte("foo"))
	fmt.Printf("%#v \n", b) // bytes.Buffer{buf:[]uint8{0x66, 0x6f, 0x6f}, off:0, lastRead:0}

	data := make([]byte, 10)
	o, err := b.Read(data)
	if err != nil {
		fmt.Println(err)	
	}

	fmt.Printf("%#v %#v \n", o, data)

	fmt.Printf("%#v \n", b) // bytes.Buffer{buf:[]uint8{0x66, 0x6f, 0x6f}, off:3, lastRead:-1}
}
