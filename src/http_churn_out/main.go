package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

/*
Example-1 Connection churn, by not reading and closing the request we exhaust ephemeral ports.
Example-2 Increase concurrency so we can only keep two cached connections resulting in connection churn.
*/

// const concurrency = 2
const concurrency = 3 // Example-3 Connection Churn MaxIdleConnsPerHost

// https://dev.to/gkampitakis/http-connection-churn-in-go-34pl
func main() {
	go server()
	// Wait for a second for the server to start.
	time.Sleep(1 * time.Second)
	client()
}

func server() {
	m := http.NewServeMux()
	m.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "health"}`))
	})

	log.Println("listening on port 8080 🚀")
	log.Fatal(http.ListenAndServe("localhost:8080", m))
}

func client() {
	httpClient := customHttpClient()
	requests := atomic.Int64{}
	forever := make(chan struct{})

	// Every 5 seconds print how many requests we have made.
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Printf("%d requests\n", requests.Load())
		}
	}()

	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				time.Sleep(time.Millisecond)

				req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/message", nil)
				res, err := httpClient.Do(req)
				if err != nil {
					log.Println(err)
					continue
				}
				_ = res
				io.Copy(io.Discard, res.Body) // Example-1 Connection churn, ephemeral port exhaustion
				res.Body.Close()              // Example-1 Connection churn, ephemeral port exhaustion

				requests.Add(1)
			}
		}()
	}

	<-forever
}

func customHttpClient() *http.Client {
	cl := http.DefaultClient
	tr := http.DefaultTransport.(*http.Transport)
	cl.Transport = tr

	//tr.MaxIdleConnsPerHost = 2
	tr.MaxIdleConnsPerHost = 3 // Example-2 Connection Churn MaxIdleConnsPerHost

	tr.MaxIdleConns = 100
	cl.Timeout = 10 * time.Second

	return cl
}
