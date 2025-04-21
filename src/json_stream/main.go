// File: stream_array.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	// 1. Open the file (acts like createReadStream in Node)
	f, err := os.Open("/Users/tafu/LARGE_FILE/large.json")
	if err != nil {
		log.Fatalf("open large.json: %v", err)
	}
	defer f.Close()

	// 2. Create a streaming decoder
	dec := json.NewDecoder(f)

	// 3. The first token **must** be the beginning of the array: '['
	tok, err := dec.Token()
	if err != nil {
		log.Fatalf("read token: %v", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		log.Fatalf("expected '[' but got %v", tok)
	}

	// 4. Iterate over each element
	index := 0
	for dec.More() {
		var value any // decode into an unconstrained interface{}
		if err := dec.Decode(&value); err != nil {
			log.Fatalf("decode element %d: %v", index, err)
		}
		fmt.Printf("[%d] %v\n", index, value) // echo like console.log
		index++
	}

	// 5. Consume the closing ']' so the stream ends cleanly
	if _, err := dec.Token(); err != nil {
		log.Fatalf("closing token: %v", err)
	}

	fmt.Println("✅ Done")
}
