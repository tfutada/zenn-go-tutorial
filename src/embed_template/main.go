package main

import (
	"embed"
	"os"
	"text/template"
)

// Use the //go:embed directive to include the template file in the binary.
// This assumes you have a 'templates' directory with a 'hello.tmpl' file.

//go:embed templates/*
var templateFS embed.FS

type User struct {
	Name string
}

func main() {
	// Access the file from the embedded file system.
	tmplData, err := templateFS.ReadFile("templates/hello.tmpl")
	if err != nil {
		panic(err)
	}

	// Create a new template and parse the text from the embedded file.
	tmpl, err := template.New("hello").Parse(string(tmplData))
	if err != nil {
		panic(err)
	}

	// Execute the template with some data.
	user := User{Name: "Tech Guru"}
	err = tmpl.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}
}
