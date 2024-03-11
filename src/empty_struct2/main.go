package main

import "fmt"

// Eater is an interface with an Eat method
type Eater interface {
	Eat()
}

// Dog is a type that implements the Eater interface
type Dog struct {
	Name string
}

// Eat Implement the Eat method for Dog
func (d Dog) Eat() {
	fmt.Printf("%s is eating\n", d.Name)
}

// Cat is a type that implements the Eater interface
type Cat struct {
	Name string
}

// Eat Implement the Eat method for Cat
func (c Cat) Eat() {
	fmt.Printf("%s is eating\n", c.Name)
}

// https://medium.com/towardsdev/golang-empty-structs-c9942f3547b3
// This can be particularly useful when you want to group different types based on a shared behaviour without storing additional data.
func main() {
	// Create instances of Dog and Cat
	myDog := Dog{Name: "Buddy"}
	myDog2 := Dog{Name: "Buddy"}
	myCat := Cat{Name: "Whiskers"}

	// Use an empty struct to create a set of eaters
	eaters := make(map[Eater]struct{})
	eaters[myDog] = struct{}{}
	eaters[myDog2] = struct{}{}
	eaters[myCat] = struct{}{}

	// Iterate over the eaters and make them eat
	for eater := range eaters {
		eater.Eat()
	}
}
