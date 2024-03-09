package main

import (
	"fmt"
	"sync"
	"time"
)

type Singleton struct {
	// Properties of your singleton
}

var (
	instance *Singleton
	once     sync.Once
)

// GetInstance returns the singleton instance.
func GetInstance() *Singleton {
	once.Do(func() {
		instance = &Singleton{}
	})
	return instance
}

func main() {
	// This will print the same instance address every time.
	for i := 0; i < 10; i++ {
		go func() {
			// print the address of the singleton instance
			fmt.Printf("%p\n", GetInstance())
		}()
	}
	time.Sleep(1 * time.Second)
	fmt.Printf("done\n")
}
