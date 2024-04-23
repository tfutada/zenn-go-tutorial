package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Hour,
		MaxConnsPerHost:       99999,
		DisableKeepAlives:     true,
	}

	myClient := &http.Client{Transport: tr}

	for i := 0; i < 1000; i++ {
		go func(n int) {
			resp, err := myClient.Get("http://127.0.0.1:8080")
			if err != nil {
				fmt.Printf("%d: %s\n", n, err.Error())
				return
			}
			defer resp.Body.Close() // Ensure the response body is closed

			// dump resp
			_, err = io.Copy(os.Stdout, resp.Body)
			if err != nil {
				fmt.Printf("%d: Failed to copy response body to stdout: %s\n", n, err.Error())
				return
			}
		}(i)

		time.Sleep(1 * time.Millisecond)

		if i%100 == 0 {
			fmt.Printf("Sent %d requests\n", i)
			time.Sleep(1 * time.Second)
		}
	}

	time.Sleep(time.Hour)
}
