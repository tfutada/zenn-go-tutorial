package main

import (
	"errors"
	"fmt"
)

var (
	err1 = errors.New("error 1st")
	err2 = errors.New("error 2nd")
)

func main() {
	err := err1
	err = errors.Join(err, err2)

	fmt.Println(errors.Is(err, err1)) // true
	fmt.Println(errors.Is(err, err2)) // true
}
