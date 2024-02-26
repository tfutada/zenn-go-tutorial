package main

import "fmt"

type Person struct {
	name string
	age  int
}

func main() {
	person := newPerson("John", 25)
	fmt.Println(person)
}

func newPerson(name string, age int) *Person {
	// it is ok to return an address of a local variable,
	// which survives after the function returns.
	return &Person{name, age}
}
