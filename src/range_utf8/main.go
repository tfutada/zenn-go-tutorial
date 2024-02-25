package main

import "fmt"

// range is aware of UTF-8 characters.
func main() {
	for pos, char := range "日本\x80語" { // \x80 is an illegal UTF-8 encoding
		fmt.Printf("character %#U starts at byte position %d\n", char, pos)
	}
}

//character U+65E5 '日' starts at byte position 0
//character U+672C '本' starts at byte position 3
//character U+FFFD '�' starts at byte position 6
//character U+8A9E '語' starts at byte position 7
