package main

import "tutorial1/src/interface_check"

type Dog struct{}

func main() {
	dog := Dog{}
	doSomething(dog)
}

func doSomething(animal interface_check.Animal) {
	animal.Speak()
}
