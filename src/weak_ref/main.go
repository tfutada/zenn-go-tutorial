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
	data map[string]weak.Pointer[User]
}

func main() {
	cache := Cache{data: make(map[string]weak.Pointer[User])}
	u := &User{"Alice"}

	cache.data["alice"] = weak.Make(u) // weak ref
	longLived := []*User{u}            // strong ref stays forever

	fmt.Println("Before GC - Strong ref exists:", longLived[0].Name)

	u = nil
	runtime.GC()

	// Even though weak pointer should allow GC…
	// GC sees longLived → *User still reachable
	fmt.Println("After GC - Strong ref still exists:", longLived[0].Name)
	if ptr := cache.data["alice"].Value(); ptr != nil {
		fmt.Println("Weak pointer still valid:", ptr.Name)
	} else {
		fmt.Println("Weak pointer: User has been garbage collected")
	}

	// Demonstrate actual GC: remove strong reference
	fmt.Println("\nRemoving strong reference...")
	longLived = nil
	runtime.GC()
	runtime.GC() // Run GC twice to ensure collection

	if ptr := cache.data["alice"].Value(); ptr != nil {
		fmt.Println("Still alive:", ptr.Name)
	} else {
		fmt.Println("User has been garbage collected")
	}
}
