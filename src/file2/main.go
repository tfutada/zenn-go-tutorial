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

	var data []byte

	for {
		temp := make([]byte, 10)

		count, err := f.Read(temp)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				break
			}
		}

		data = append(data, temp[:count]...)
	}

	fmt.Printf("count %v\n", len(data))
}

// in Rust, you need to provide a buffer to receive file contents, but no need to specify the size.
// let mut contents = String::new();
//    file.read_to_string(&mut contents)?;
