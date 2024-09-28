package main

import (
	"fmt"
	"sync"
)

type T struct {
	lock sync.Mutex
}

func (t T) Lock() {
	t.lock.Lock()
}

func (t T) Unlock() {
	t.lock.Unlock()
}

func main() {
	var t T
	t.Lock()
	fmt.Println("test")
	t.Unlock()
	fmt.Println("finished")
}
