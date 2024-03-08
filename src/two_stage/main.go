package main

import "fmt"

func setupTeardown() func() {
	fmt.Println("Run initialization")
	return func() {
		fmt.Println("Run cleanup")
	}
}

func main() {
	defer setupTeardown()() // <-------- double ()
	fmt.Println("Main function called")
}
