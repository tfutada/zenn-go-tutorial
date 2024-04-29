package main

import (
	"bufio"
	"os"
)

func main() {
	f := os.Stdout

	w := bufio.NewWriter(f)

	w.WriteString("a")
	w.WriteString("b")
	w.WriteString("c")

	// Perform 1 write system call
	w.Flush()
}
