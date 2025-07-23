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

// deserializeJSON is a function that takes a JSON byte slice and unmarshals it into a Person struct.
func main() {
	// JSON data (as a string for this example)
	jsonStr := `{"name":"Alice","age":25}`
	jsonBytes := []byte(jsonStr)

	// Create a Person struct to hold the deserialized data
	var tPerson Person

	// Deserialize JSON to struct with v2 semantics
	err := json.Unmarshal(jsonBytes, &tPerson)
	if err != nil {
		log.Fatal("Error deserializing JSON:", err)
	}

	// Print the deserialized struct
	fmt.Printf("Deserialized: %+v\n", tPerson)
}
