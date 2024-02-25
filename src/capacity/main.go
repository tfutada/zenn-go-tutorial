package main

import "fmt"

func main() {
	x2Counter := 0
	for i := 0; i < 1000; i++ {
		slice := make([]int, i, i)
		slice = append(slice, 0)
		if cap(slice) == i*2 {
			x2Counter++
			println(i)
		}
	}
	fmt.Println(x2Counter)
}
