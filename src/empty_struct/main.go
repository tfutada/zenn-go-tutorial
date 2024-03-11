package main

import (
	"fmt"
)

// Set is a set-like data structure using an empty struct
type Set map[string]struct{}

func main() {
	// Create a set
	mySet := make(Set)

	// Add elements to the set
	mySet["apple"] = struct{}{}
	mySet["banana"] = struct{}{}
	mySet["banana"] = struct{}{}
	mySet["orange"] = struct{}{}

	// Check if an element is in the set
	if _, exists := mySet["banana"]; exists {
		fmt.Println("Banana is in the set!")
	}

	// Print all elements in the set
	for key := range mySet {
		fmt.Println(key)
	}
}
