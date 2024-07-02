package main

import (
	"bytes"
	"io"
	"os"
)

// UTF-8 BOM
var bom = []byte{0xef, 0xbb, 0xbf}

func main() {
	io.Copy(os.Stdout, io.MultiReader(bytes.NewReader(bom), os.Stdin))

	buf := new(bytes.Buffer)

	input := buf.Bytes()
	if !bytes.HasPrefix(input, bom) {
		// Prepend BOM if not already present
		input = append(bom, input...)
	}
}

// check using xxd command
