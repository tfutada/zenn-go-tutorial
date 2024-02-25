package main

import (
	"fmt"
	"net/http"
)

func itemHandler(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id") // Dynamically access the path variable
	fmt.Fprintf(w, "Retrieving item with ID: %s", id)
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path") // Accessing the wildcard path variable
	fmt.Fprintf(w, "Accessing file at: %s", path)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/items/{id}", itemHandler)
	mux.HandleFunc("/files/{path...}", filesHandler)

	http.ListenAndServe(":8080", mux)
}
