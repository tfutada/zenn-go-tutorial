package main

import "fmt"

func main() {
	for i := 0; i < 100; i++ {
		slice := make([]struct{}, i, i)
		slice = append(slice, struct{}{})
		fmt.Printf("slice: len=%d, cap=%d\n", len(slice), cap(slice))
	}
}
