package main

import (
	"fmt"
	"github.com/imroc/req/v3"
	"time"
)

type Post struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func main() {
	client := req.NewClient()
	var posts []Post

	start := time.Now()
	resp, err := client.R().
		SetSuccessResult(&posts).
		Get("https://jsonplaceholder.typicode.com/posts")
	duration := time.Since(start)
	fmt.Printf("Fetch duration: %v\n", duration)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	if resp.IsErrorState() {
		fmt.Println("Request failed with status:", resp.Status)
		return
	}

	fmt.Printf("Total posts: %d\n", len(posts))
	for i := 0; i < 5; i++ { // show first 5
		fmt.Printf("Post %d: %+v\n", i+1, posts[i])
	}
}
