package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Backend represents a target server configuration
type Backend struct {
	URL *url.URL
}

// NewProxy creates a new reverse proxy for a given backend
func NewProxy(target *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(target)
}

// ProxyHandler handles routing to the appropriate backend
func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Define backend servers (could be dynamic in a real implementation)
	backends := map[string]*Backend{
		"/api": {
			URL: mustParseURL("http://localhost:8081"),
		},
		"/web": {
			URL: mustParseURL("http://localhost:8082"),
		},
	}

	// Simple routing logic based on path prefix
	for prefix, backend := range backends {
		if r.URL.Path[:len(prefix)] == prefix {
			proxy := NewProxy(backend.URL)
			proxy.ServeHTTP(w, r)
			return
		}
	}

	// Fallback if no route matches
	http.Error(w, "Not Found", http.StatusNotFound)
}

// mustParseURL is a helper function to parse URLs with error handling
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}
	return u
}

func main() {
	// Set up the HTTP server
	http.HandleFunc("/", ProxyHandler)

	// Start the server
	log.Println("Starting reverse proxy on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
