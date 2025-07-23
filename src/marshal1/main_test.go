package main

import (
	"encoding/json"
	"testing"
)

var jsonStr = []byte(`{"name":"Alice","age":25}`)

func BenchmarkJSONUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var p Person
		if err := json.Unmarshal(jsonStr, &p); err != nil {
			b.Fatalf("failed to unmarshal: %v", err)
		}
	}
}
