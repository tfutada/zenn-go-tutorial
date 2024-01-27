package main

import (
    "io"
    "strings"
    "testing"
)

// Your ReadOS function here...

// TestReadOS tests the ReadOS function
func TestReadOS(t *testing.T) {
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
            reader:   strings.NewReader("Hello, World!"),
            n:        1,
            Fsize:    13,
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

