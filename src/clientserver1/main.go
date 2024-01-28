package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	port := 9999
	idleTime := time.Second

	if err := createServer(9999, idleTime); err != nil {
		log.Fatalf("error creating server: %s", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			IdleConnTimeout:       idleTime + time.Millisecond,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	urlString := fmt.Sprintf("http://localhost:%d/test", port)
	payloadString := "my payload"

	for {
		err := request(client, urlString, payloadString)
		if err != nil {
			log.Printf("request failed: %s", err)
		} else {
			log.Printf("request succeeded")
		}

		time.Sleep(idleTime)
	}
}

func createServer(port int, idleTime time.Duration) error {
	server := &http.Server{
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(`ok!`)); err != nil {
					log.Printf("error writing to client: %s", err)
				}
			},
		),
		IdleTimeout: idleTime,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("error listening: %s", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("error serving: %s", err)
		}
		if err := listener.Close(); err != nil {
			log.Printf("error closing listener: %s", err)
		}
	}()

	return nil
}

func request(client *http.Client, urlString, payloadString string) error {
	payload := bytes.NewBufferString(payloadString)

	req, err := http.NewRequest(http.MethodPost, urlString, payload)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return fmt.Errorf("error reading body: %s", err)
	}

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("error closing body: %s", err)
	}

	if string(body) != "ok!" {
		return fmt.Errorf("unexpected response: %s", body)
	}

	return nil
}
