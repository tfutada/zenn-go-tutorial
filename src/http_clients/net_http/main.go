package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// This example demonstrates how to use the net/http package to make a GET request
// https://dev.to/shrsv/http-requests-in-go-only-the-most-useful-libraries-8nb?utm_source=dormosheio&utm_campaign=dormosheio
func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://jsonplaceholder.typicode.com/users/1", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	fmt.Printf("User: %+v\n", user)
	// Output: User: {ID:1 Name:Leanne Graham}
}
