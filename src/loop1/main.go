package main

import (
	"fmt"
	"time"
)

func main() {
	data := []string{"one", "two", "three"}

	for _, v := range data {
		// fork
		go func(v string) {
			fmt.Println(v)
		}(v)
	}

	time.Sleep(3 * time.Second)
	//goroutines print: three, three, three
}
