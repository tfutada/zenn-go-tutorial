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
type ErrorMessage struct {
	// Message is the error message returned by the API
	// NB. jsonplaceholder.typicode.com does not return error messages
	Message string `json:"message"`
}

func main() {

	client := req.NewClient()
	var posts []Post
	var errMsg ErrorMessage

	start := time.Now()
	resp, err := client.
		SetTimeout(10 * time.Second).
		R().
		EnableDumpWithoutHeader().
		SetSuccessResult(&posts).
		SetErrorResult(&errMsg).
		Get("https://jsonplaceholder.typicode.com/posts")

	duration := time.Since(start)
	fmt.Printf("Fetch duration: %v\n", duration)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	if resp.IsErrorState() {
		fmt.Println("Request failed with status:", resp.Status)
		fmt.Println(errMsg.Message) // Record error message returned.
		return
	}

	fmt.Printf("Total posts: %d\n", len(posts))
	for i := 0; i < 5; i++ { // show first 5
		fmt.Printf("Post %d: %+v\n", i+1, posts[i])
	}
}
