package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		time.Sleep(time.Second * 5)
		fmt.Println("Hello")
	}()

	for i := 0; i < 10; i++ {
		fmt.Println(i)
		time.Sleep(time.Millisecond * 100)
	}

	wg.Wait()
}
