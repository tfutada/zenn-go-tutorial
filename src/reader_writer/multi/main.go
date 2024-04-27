package main

import (
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	r1 := strings.NewReader("first reader\n")
	r2 := strings.NewReader("second reader\n")
	r3 := strings.NewReader("third reader\n")

	// add to a slice
	readers := []io.Reader{r1, r2, r3}

	// merge all 3 readers into one
	r := io.MultiReader(readers...)

	if _, err := io.Copy(os.Stdout, r); err != nil {
		log.Fatal(err)
	}

}
