package main

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	client := resty.New()
	var user User

	resp, err := client.R().
		SetResult(&user).
		Get("https://jsonplaceholder.typicode.com/users/1")
	if err != nil {
		log.Fatal("Error sending request:", err)
	}

	if resp.IsError() {
		log.Fatalf("Request failed with status: %s", resp.Status())
	}

	fmt.Printf("User: %+v\n", user)
	// Expected Output: User: {ID:1 Name:Leanne Graham}
}
