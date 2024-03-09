package main

import (
	"fmt"
	"reflect"
)

func main() {
	var x interface{}
	var y *int = nil
	x = y

	if x != nil {
		fmt.Println("x != nil") // x itself is not nil, but the value it holds is nil
	} else {
		fmt.Println("x == nil")
	}

	fmt.Println(x)

	v := reflect.ValueOf(x).IsNil()
	fmt.Println(v) // true
}
