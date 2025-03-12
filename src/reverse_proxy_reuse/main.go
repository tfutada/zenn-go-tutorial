// https://medium.com/@abhinavv.singh/reverse-proxy-in-go-handling-millions-of-traffic-seamlessly-a76b12d49494
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Backend represents a target server configuration
type Backend struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy // Store the proxy instance
}

// mustParseURL is a helper function to parse URLs with error handling
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}
	return u
}

// ProxyHandler routes requests to pre-initialized proxies
// high order function that returns a http.HandlerFunc that routes requests to pre-initialized proxies
func ProxyHandler(backends map[string]*Backend) http.HandlerFunc {
	// Return the handler function with the given pre-initialized backends
	return func(w http.ResponseWriter, r *http.Request) {
		for prefix, backend := range backends {
			if strings.HasPrefix(r.URL.Path, prefix) { // Safe prefix check
				// also cache the response with the path as the key
				backend.Proxy.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func main() {
	// Initialize backends with proxies once at startup
	backends := map[string]*Backend{
		"/api": {
			URL:   mustParseURL("http://localhost:8081"),
			Proxy: httputil.NewSingleHostReverseProxy(mustParseURL("http://localhost:8081")),
		},
		"/web": {
			URL:   mustParseURL("http://localhost:8082"),
			Proxy: httputil.NewSingleHostReverseProxy(mustParseURL("http://localhost:8082")),
		},
	}

	// Set up the HTTP server with the handler
	http.HandleFunc("/", ProxyHandler(backends))

	// Start the server
	log.Println("Starting reverse proxy on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
