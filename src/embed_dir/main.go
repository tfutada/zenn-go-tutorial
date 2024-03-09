package main

import (
	"embed"
	"fmt"
)

//go:embed file/*
var files embed.FS

func main() {
	hello, _ := files.ReadFile("file/hello.txt")
	world, _ := files.ReadFile("file/world.txt")
	fmt.Println(string(hello), string(world))
}
