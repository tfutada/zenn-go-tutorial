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
	fmt.Println("The program is now using", memstat.Alloc/KiB, "KiB")
}

func main() {
	data, err := os.ReadFile("src/interning/large_book.txt")
	if err != nil {
		log.Fatalln("Could not read file:", err)
	}
	book := string(data)
	mem()
	// Print the book contents
	fmt.Println("book:", book)

	mem()

}
