package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")

	var m map[string]int

	for k, v := range m {
		fmt.Println(k, v)
	}
}
