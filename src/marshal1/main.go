package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// JSON data (as a string for this example)
	jsonStr := `{"name":"Alice","age":25}`

	// Create a Person struct to hold the deserialized data
	var person Person

	// Deserialize JSON to struct with v2 semantics
	err := json.Unmarshal([]byte(jsonStr), &person)
	if err != nil {
		log.Fatal("Error deserializing JSON:", err)
	}

	// Print the deserialized struct
	fmt.Printf("Deserialized: %+v\n", person)
}
