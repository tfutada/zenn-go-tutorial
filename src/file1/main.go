package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	f, err := os.Open("foo.txt")
	if err != nil {
		fmt.Println(err)
	}

	var totalCount int

	for {
		data := make([]byte, 10)
	
		count, err := f.Read(data)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				break
			}
		}

		totalCount += count
	}

	fmt.Printf("count %v\n", totalCount)
}

// in Rust, you need to provide a buffer to receive file contents, but no need to specify the size.
// let mut contents = String::new();
//    file.read_to_string(&mut contents)?;

