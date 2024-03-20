package main

import (
	"fmt"
	"net/http"
	"time"
)

func itemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Dynamically access the path variable
	fmt.Fprintf(w, "Retrieving item with ID: %s", id)
}

func delayHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second) // Sleep for 2 seconds
	fmt.Fprint(w, "Delay complete")

}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path") // Accessing the wildcard path variable
	fmt.Fprintf(w, "Accessing file at: %s", path)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/items/{id}", itemHandler)
	mux.HandleFunc("/files/{path...}", filesHandler)
	mux.HandleFunc("/delay", delayHandler)

	http.ListenAndServe(":8080", mux)
}
