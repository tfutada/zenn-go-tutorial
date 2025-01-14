package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

var memstat runtime.MemStats

func mem() {
	runtime.GC()
	runtime.ReadMemStats(&memstat)
	const KiB = 1024
	fmt.Println("! The program is now using: ", memstat.Alloc/KiB, "KiB")
}

func main() {
	data, err := os.ReadFile("src/interning/large_book.txt")
	if err != nil {
		log.Fatalln("Could not read file:", err)
	}

	book := string(data)
	mem()
	// Print the book contents
	//fmt.Println("book:", book)

	var Bwords []string
	// 'i' will move through the string as we locate each word
	for i := 0; i < len(book); {
		// 1. Skip any leading spaces
		for i < len(book) && book[i] == ' ' {
			i++
		}

		// If we've reached the end of the string after skipping spaces, stop
		if i >= len(book) {
			break
		}

		// 2. Identify the start of the word
		start := i

		// 3. Move 'i' forward until a space or end-of-string
		for i < len(book) && book[i] != ' ' {
			i++
		}

		// Extract the word
		word := book[start:i]

		// 4. Check if this word starts with B or b
		if len(word) > 0 && (word[0] == 'B' || word[0] == 'b') {
			Bwords = append(Bwords, word)
		}

		// The loop continues until i reaches len(book)
	}

	fmt.Printf("Found %d words starting with 'B' or 'b'\n", len(Bwords))
	fmt.Printf("The first 10 words are: %#v\n", Bwords[:10])
	mem()
}
