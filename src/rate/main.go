package main

import (
	"fmt"

	"golang.org/x/time/rate"
)

func main() {
	s := rate.Sometimes{Every: 2}

	for i := range 10 {
		s.Do(func() { fmt.Println(i) })
	}
}
