package main

import "fmt"

func main() {
	var s string
	// writer
	go func() {
		arr := [2]string{"", "hello"}
		for i := 0; ; i++ {
			s = arr[i%2]
		}
	}()
	// reader
	for {
		fmt.Println(s)
	}
}
