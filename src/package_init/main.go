package main

import (
	"log"
	. "math"
	_ "tutorial1/src/package_init/lab"
)

func main() {
	log.Println("main function")

	// calc an area of a circle with math pi.
	area := Pi * 9
	log.Println("area of a circle:", area)
}
