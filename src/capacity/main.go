package main

import "fmt"

func main() {
	slice := make([]int, 336, 336)
	slice = append(slice, 0)
	fmt.Printf("len=%d, cap=%d", len(slice), cap(slice))
}
