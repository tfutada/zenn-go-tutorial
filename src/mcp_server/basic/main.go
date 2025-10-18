package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Basic MCP Server Example
//
// This example demonstrates how to create a simple MCP (Model Context Protocol) server
// in Go using the official SDK. The server exposes several tools that can be called
// by MCP clients like Claude.
//
// What is MCP?
// MCP is an open standard that enables AI applications to connect to external data
// sources and tools. Think of it as a "USB-C port for AI" - a standardized way for
// LLMs to interact with your services.
//
// Running this server:
//   go run src/mcp_server/basic/main.go
//
// The server communicates over stdio (standard input/output), which is the simplest
// transport mechanism for local tools.

// Input/Output types for tools - the SDK uses these to automatically generate JSON schemas

type CalculatorInput struct {
	Operation string   `json:"operation"`
	A         float64  `json:"a"`
	B         *float64 `json:"b,omitempty"`
}

type CalculatorOutput struct {
	Result float64 `json:"result"`
}

type EchoInput struct {
	Text      string `json:"text"`
	Transform string `json:"transform,omitempty"`
}

type EchoOutput struct {
	Result string `json:"result"`
}

type TimestampInput struct {
	Format string `json:"format"`
}

type TimestampOutput struct {
	Timestamp string `json:"timestamp"`
}

type WeatherInput struct {
	City  string `json:"city"`
	Units string `json:"units,omitempty"`
}

type WeatherOutput struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Units       string  `json:"units"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
	Timestamp   string  `json:"timestamp"`
}

func main() {
	// Create an MCP server with implementation details
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "basic-mcp-server", // Server identifier
		Version: "1.0.0",             // Server version
	}, nil)

	// Register tools - these are functions that the LLM can call
	mcp.AddTool(server, &mcp.Tool{
		Name:        "calculator",
		Description: "Performs basic arithmetic operations (add, subtract, multiply, divide, power, sqrt)",
	}, calculatorHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echoes back text with optional transformations (uppercase, lowercase, reverse, word_count)",
	}, echoHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "timestamp",
		Description: "Returns current timestamp in various formats",
	}, timestampHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_weather",
		Description: "Gets mock weather data for a given city (demonstration only)",
	}, weatherHandler)

	log.Println("Starting Basic MCP Server...")
	log.Println("Registered tools: calculator, echo, timestamp, get_weather")
	log.Println("Server ready to accept requests over stdio")

	// Run the server over stdio transport
	// This means the server reads requests from stdin and writes responses to stdout
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// calculatorHandler performs basic arithmetic operations
func calculatorHandler(ctx context.Context, request *mcp.CallToolRequest, input CalculatorInput) (*mcp.CallToolResult, CalculatorOutput, error) {
	var result float64
	b := 0.0
	if input.B != nil {
		b = *input.B
	}

	switch input.Operation {
	case "add":
		result = input.A + b
	case "subtract":
		result = input.A - b
	case "multiply":
		result = input.A * b
	case "divide":
		if b == 0 {
			return nil, CalculatorOutput{}, fmt.Errorf("division by zero")
		}
		result = input.A / b
	case "power":
		result = math.Pow(input.A, b)
	case "sqrt":
		if input.A < 0 {
			return nil, CalculatorOutput{}, fmt.Errorf("cannot take square root of negative number")
		}
		result = math.Sqrt(input.A)
	default:
		return nil, CalculatorOutput{}, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	output := CalculatorOutput{Result: result}

	// Return nil for CallToolResult - SDK will automatically create it from output
	return nil, output, nil
}

// echoHandler manipulates and echoes back text
func echoHandler(ctx context.Context, request *mcp.CallToolRequest, input EchoInput) (*mcp.CallToolResult, EchoOutput, error) {
	transform := input.Transform
	if transform == "" {
		transform = "none"
	}

	var result string
	switch transform {
	case "uppercase":
		result = strings.ToUpper(input.Text)
	case "lowercase":
		result = strings.ToLower(input.Text)
	case "reverse":
		runes := []rune(input.Text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	case "word_count":
		words := strings.Fields(input.Text)
		result = fmt.Sprintf("Word count: %d", len(words))
	default:
		result = input.Text
	}

	output := EchoOutput{Result: result}

	// Return nil for CallToolResult - SDK will automatically create it from output
	return nil, output, nil
}

// timestampHandler returns current time in various formats
func timestampHandler(ctx context.Context, request *mcp.CallToolRequest, input TimestampInput) (*mcp.CallToolResult, TimestampOutput, error) {
	now := time.Now()
	var result string

	switch input.Format {
	case "unix":
		result = fmt.Sprintf("Unix timestamp: %d", now.Unix())
	case "iso8601":
		result = fmt.Sprintf("ISO 8601: %s", now.Format("2006-01-02T15:04:05Z07:00"))
	case "rfc3339":
		result = fmt.Sprintf("RFC 3339: %s", now.Format(time.RFC3339))
	case "human":
		result = fmt.Sprintf("Human readable: %s", now.Format("Monday, January 2, 2006 at 3:04:05 PM MST"))
	default:
		return nil, TimestampOutput{}, fmt.Errorf("unknown format: %s", input.Format)
	}

	output := TimestampOutput{Timestamp: result}

	// Return nil for CallToolResult - SDK will automatically create it from output
	return nil, output, nil
}

// weatherHandler returns mock weather data
func weatherHandler(ctx context.Context, request *mcp.CallToolRequest, input WeatherInput) (*mcp.CallToolResult, WeatherOutput, error) {
	units := input.Units
	if units == "" {
		units = "celsius"
	}

	temp := 22.5
	if units == "fahrenheit" {
		temp = temp*9/5 + 32
	}

	output := WeatherOutput{
		City:        input.City,
		Temperature: temp,
		Units:       units,
		Condition:   "Partly Cloudy",
		Humidity:    65,
		WindSpeed:   12.3,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Return nil for CallToolResult - SDK will automatically create it from output
	return nil, output, nil
}
