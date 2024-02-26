package main

import "fmt"

type A struct {
}

func (a *A) b() {
	if a == nil {
		fmt.Println("a is nil")
		return
	}

	fmt.Println(1000)
}

// nil pointer receiver
func main() {
	var a *A
	a.b()
}
