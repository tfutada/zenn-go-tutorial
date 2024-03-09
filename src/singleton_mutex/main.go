package main

import (
	"fmt"
	"sync"
	"time"
)

type Singleton struct {
	// Properties of your singleton
	count int
	mux   sync.Mutex
}

var (
	instance *Singleton
	once     sync.Once
)

// SafeIncrement synchronizes access to the count property.
func (s *Singleton) SafeIncrement() int {
	s.mux.Lock()
	defer s.mux.Unlock() // This defer ensures that the mutex is unlocked when the function exits.
	s.count++
	fmt.Println("Safe Count:", s.count)
	return s.count
}

// GetInstance returns the singleton instance.
func GetInstance() *Singleton {
	once.Do(func() {
		instance = &Singleton{
			count: 0,
		}
	})
	return instance
}

func main() {
	// This will print the same instance address every time.
	for i := 0; i < 10; i++ {
		go func() {
			// print the address of the singleton instance
			fmt.Printf("%v\n", GetInstance().SafeIncrement())
		}()
	}
	time.Sleep(1 * time.Second)
	fmt.Printf("done\n")
}
