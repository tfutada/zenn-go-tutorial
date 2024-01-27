package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testing"
)

// Your ReadOS function here...

// TestReadOS tests the ReadOS function
func TestReadOS(t *testing.T) {
	fileName := "big-1mb.log"

	f, err := os.Open(fileName)
	defer f.Close()

	if err != nil {
		fmt.Println(err)
	}

	size, err := getFileSize(fileName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// use bufio
	reader := bufio.NewReader(f)

	tests := []struct {
		name     string
		reader   io.Reader
		n        int
		Fsize    int
		expected int
		wantErr  bool
	}{
		{
			name:     "NormalRead",
			reader:   reader,
			n:        1,
			Fsize:    size,
			expected: 13,
			wantErr:  false,
		},
		// Add more test cases here...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReadOS(tt.reader, tt.n, tt.Fsize)
		})
	}
}

func getFileSize(fileName string) (int, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return int(fileInfo.Size()), nil
}
