package main

import (
	"fmt"
	"io"
	"log"
	"strings"
)

func main() {
	pipeReader, pipeWriter := io.Pipe()

	// Start echo in a goroutine to handle writing concurrently
	go func() {
		defer pipeWriter.Close() // Ensure the writer is closed after writing
		if err := echo(pipeWriter, "hello"); err != nil {
			log.Printf("error writing to pipe: %v", err)
			return
		}
	}()

	// Read and transform the data in the main goroutine
	if err := tr(pipeReader, "e", "i"); err != nil {
		log.Printf("error reading from pipe: %v", err)
	}
}

func echo(w io.Writer, s string) error {
	_, err := fmt.Fprint(w, s)
	return err
}

func tr(r io.Reader, old string, new string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	res := strings.Replace(string(data), old, new, -1)
	fmt.Println(res)
	return nil
}
