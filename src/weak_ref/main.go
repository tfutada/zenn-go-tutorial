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

	// Add finalizer to make GC behavior visible
	runtime.SetFinalizer(u, func(*User) {
		fmt.Println("User finalizer called - object being collected")
	})

	cache.data["alice"] = weak.Make(u) // weak ref
	longLived := []*User{u}            // strong ref prevents GC

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

	if ptr := cache.data["alice"].Value(); ptr != nil {
		fmt.Println("Still alive:", ptr.Name)
	} else {
		fmt.Println("User has been garbage collected")
	}
}
