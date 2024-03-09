package main

import (
	_ "embed"
	"fmt"
)

//go:embed hello.txt
var hello string

//go:embed file/world.txt
var world []byte

func main() {
	fmt.Println(hello)
	fmt.Println(world)
}
