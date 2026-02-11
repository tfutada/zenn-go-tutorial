package main

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
)

// Small payload - simple struct
type Small struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// Medium payload - nested struct with slices
type Medium struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Age      int      `json:"age"`
	Active   bool     `json:"active"`
	Tags     []string `json:"tags"`
	Address  Address  `json:"address"`
	Metadata map[string]string `json:"metadata"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	Zip     string `json:"zip"`
}

// Large payload - slice of medium structs
type Large struct {
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Results []Medium `json:"results"`
}

var (
	smallObj = Small{
		ID:     1,
		Name:   "Alice",
		Active: true,
	}

	mediumObj = Medium{
		ID:     42,
		Name:   "Bob Smith",
		Email:  "bob@example.com",
		Age:    30,
		Active: true,
		Tags:   []string{"go", "rust", "python", "typescript"},
		Address: Address{
			Street:  "123 Main St",
			City:    "Tokyo",
			Country: "Japan",
			Zip:     "100-0001",
		},
		Metadata: map[string]string{
			"role":       "engineer",
			"department": "backend",
			"level":      "senior",
		},
	}

	largeObj = func() Large {
		results := make([]Medium, 100)
		for i := range results {
			results[i] = mediumObj
			results[i].ID = i
		}
		return Large{Total: 100, Page: 1, Results: results}
	}()

	// Pre-marshaled JSON for unmarshal benchmarks
	smallJSON, _  = json.Marshal(smallObj)
	mediumJSON, _ = json.Marshal(mediumObj)
	largeJSON, _  = json.Marshal(largeObj)
)

// Sinks to prevent compiler optimization
var (
	bytesSink []byte
	smallSink Small
	medSink   Medium
	largeSink Large
)

// --- Marshal benchmarks ---

func BenchmarkMarshal_Small_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytesSink, _ = json.Marshal(smallObj)
	}
}

func BenchmarkMarshal_Small_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytesSink, _ = sonic.Marshal(smallObj)
	}
}

func BenchmarkMarshal_Medium_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytesSink, _ = json.Marshal(mediumObj)
	}
}

func BenchmarkMarshal_Medium_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytesSink, _ = sonic.Marshal(mediumObj)
	}
}

func BenchmarkMarshal_Large_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytesSink, _ = json.Marshal(largeObj)
	}
}

func BenchmarkMarshal_Large_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytesSink, _ = sonic.Marshal(largeObj)
	}
}

// --- Unmarshal benchmarks ---

func BenchmarkUnmarshal_Small_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Small
		_ = json.Unmarshal(smallJSON, &v)
		smallSink = v
	}
}

func BenchmarkUnmarshal_Small_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Small
		_ = sonic.Unmarshal(smallJSON, &v)
		smallSink = v
	}
}

func BenchmarkUnmarshal_Medium_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Medium
		_ = json.Unmarshal(mediumJSON, &v)
		medSink = v
	}
}

func BenchmarkUnmarshal_Medium_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Medium
		_ = sonic.Unmarshal(mediumJSON, &v)
		medSink = v
	}
}

func BenchmarkUnmarshal_Large_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Large
		_ = json.Unmarshal(largeJSON, &v)
		largeSink = v
	}
}

func BenchmarkUnmarshal_Large_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Large
		_ = sonic.Unmarshal(largeJSON, &v)
		largeSink = v
	}
}

// --- Schemaless (interface{}) unmarshal benchmarks ---

var ifaceSink interface{}

func BenchmarkUnmarshalSchemaless_Medium_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v interface{}
		_ = json.Unmarshal(mediumJSON, &v)
		ifaceSink = v
	}
}

func BenchmarkUnmarshalSchemaless_Medium_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v interface{}
		_ = sonic.Unmarshal(mediumJSON, &v)
		ifaceSink = v
	}
}

// --- Get (sonic-specific: extract field without full unmarshal) ---

func BenchmarkGet_Sonic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node, _ := sonic.Get(mediumJSON, "address", "city")
		bytesSink, _ = node.MarshalJSON()
	}
}

func BenchmarkGet_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v Medium
		_ = json.Unmarshal(mediumJSON, &v)
		bytesSink, _ = json.Marshal(v.Address.City)
	}
}
