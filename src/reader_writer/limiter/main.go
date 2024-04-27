package main

import (
	"io"
	"log"
	"os"
	"strings"
)

// https://medium.com/@andreiboar/fundamentals-of-i-o-in-go-part-2-e7bb68cd5608
func main() {
	r := strings.NewReader("Hello!")

	lr := io.LimitReader(r, 4)

	// Output: Hell
	if _, err := io.Copy(os.Stdout, lr); err != nil {
		log.Fatal(err)
	}
}
