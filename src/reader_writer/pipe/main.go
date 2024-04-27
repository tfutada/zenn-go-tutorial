package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	pipeReader, pipeWriter := io.Pipe()

	echo(pipeWriter, "hello") // deadlock here as sync writer is waiting for reader to read
	tr(pipeReader, "e", "i")
}

func echo(w io.Writer, s string) {
	fmt.Fprint(w, s)
}

func tr(r io.Reader, old string, new string) {
	data, _ := io.ReadAll(r)
	res := strings.Replace(string(data), old, new, -1)
	fmt.Println(res)
}
