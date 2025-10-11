package main

import (
	"fmt"
	"runtime"
	"weak"
)

type User struct {
	Name string
}

type Cache struct {
	data weak.Map[string, *User]
}

func main() {
	cache := Cache{data: weak.MakeMap[string, *User]()}
	u := &User{"Alice"}

	cache.data.Set("alice", u) // weak ref
	longLived := []*User{u}    // strong ref stays forever

	u = nil
	runtime.GC()

	// Even though weak map should free it…
	// GC sees longLived → *User still reachable
	fmt.Println("Still alive:", longLived[0].Name)
}
