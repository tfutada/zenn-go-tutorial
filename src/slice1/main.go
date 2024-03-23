package main

import (
	"bytes"
	"os"
)

func main() {
	_, err := readFileDetails("go.mod")
	if err != nil {
		panic(err)
	}
}

func readFileDetails(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	//return data[5:10], nil
	return bytes.Clone(data[5:10]), nil
}
