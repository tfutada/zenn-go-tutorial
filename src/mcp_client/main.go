package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP Client Example
//
// This example demonstrates how to create an MCP client in Go that connects
// to an MCP server and interacts with it by:
// 1. Listing available tools, resources, and prompts
// 2. Calling tools with parameters
// 3. Reading resources
// 4. Getting prompt templates
//
// Before running this client, you need to have an MCP server running.
// You can use either:
//   go run src/mcp_server/basic/main.go
//   go run src/mcp_server/advanced/main.go
//
// Then run this client:
//   go run src/mcp_client/main.go

func main() {
	fmt.Println("=== MCP Client Example ===\n")

	// Choose which server to connect to
	// For this demo, we'll use the advanced server
	serverCommand := "go"
	serverArgs := []string{"run", "src/mcp_server/advanced/main.go"}

	fmt.Printf("Connecting to MCP server: %s %v\n\n", serverCommand, serverArgs)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "example-mcp-client",
		Version: "1.0.0",
	}, nil)

	// Create a transport that launches the server process
	transport := &mcp.CommandTransport{
		Command: exec.Command(serverCommand, serverArgs...),
	}

	// Connect to the server
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer session.Close()

	fmt.Println("✓ Connected to MCP server successfully!\n")

	// Run example interactions
	if err := demonstrateTools(ctx, session); err != nil {
		log.Printf("Error demonstrating tools: %v", err)
	}

	if err := demonstrateResources(ctx, session); err != nil {
		log.Printf("Error demonstrating resources: %v", err)
	}

	if err := demonstratePrompts(ctx, session); err != nil {
		log.Printf("Error demonstrating prompts: %v", err)
	}

	fmt.Println("\n=== Client Session Complete ===")
}

// demonstrateTools shows how to list and call tools
func demonstrateTools(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("--- Tools Demonstration ---\n")

	// List all available tools
	toolsResult, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Printf("Available tools: %d\n", len(toolsResult.Tools))
	for i, tool := range toolsResult.Tools {
		fmt.Printf("%d. %s - %s\n", i+1, tool.Name, tool.Description)
	}
	fmt.Println()

	// Example 1: Call calculator tool
	fmt.Println("Example 1: Calling 'calculator' tool")
	calcResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "calculator",
		Arguments: map[string]interface{}{
			"operation": "add",
			"a":         15.5,
			"b":         24.3,
		},
	})
	if err != nil {
		return fmt.Errorf("calculator tool call failed: %w", err)
	}
	printToolResult("Calculator (15.5 + 24.3)", calcResult)

	// Example 2: Call echo tool
	fmt.Println("Example 2: Calling 'echo' tool")
	echoResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "echo",
		Arguments: map[string]interface{}{
			"text":      "Hello from MCP Client!",
			"transform": "uppercase",
		},
	})
	if err != nil {
		return fmt.Errorf("echo tool call failed: %w", err)
	}
	printToolResult("Echo (uppercase)", echoResult)

	// Example 3: Call timestamp tool
	fmt.Println("Example 3: Calling 'timestamp' tool")
	timeResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "timestamp",
		Arguments: map[string]interface{}{
			"format": "human",
		},
	})
	if err != nil {
		return fmt.Errorf("timestamp tool call failed: %w", err)
	}
	printToolResult("Timestamp (human format)", timeResult)

	// Example 4: Call weather tool
	fmt.Println("Example 4: Calling 'get_weather' tool")
	weatherResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_weather",
		Arguments: map[string]interface{}{
			"city":  "San Francisco",
			"units": "celsius",
		},
	})
	if err != nil {
		return fmt.Errorf("weather tool call failed: %w", err)
	}
	printToolResult("Weather", weatherResult)

	// Example 5: Demonstrate error handling
	fmt.Println("Example 5: Error handling (division by zero)")
	_, err = session.CallTool(ctx, &mcp.CallToolParams{
		Name: "calculator",
		Arguments: map[string]interface{}{
			"operation": "divide",
			"a":         10.0,
			"b":         0.0,
		},
	})
	if err != nil {
		fmt.Printf("  ✓ Error caught as expected: %v\n", err)
	} else {
		fmt.Println("  ✗ Expected an error but got none")
	}

	fmt.Println()
	return nil
}

// demonstrateResources shows how to list and read resources
func demonstrateResources(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("--- Resources Demonstration ---\n")

	// List all available resources
	resourcesResult, err := session.ListResources(ctx, nil)
	if err != nil {
		// Resources might not be available in basic server
		fmt.Println("Note: This server doesn't provide resources")
		fmt.Println("Try connecting to the advanced server to see resources\n")
		return nil
	}

	fmt.Printf("Available resources: %d\n", len(resourcesResult.Resources))
	for i, resource := range resourcesResult.Resources {
		fmt.Printf("%d. %s (%s) - %s\n", i+1, resource.Name, resource.URI, resource.Description)
	}
	fmt.Println()

	// Try to read the first resource if available
	if len(resourcesResult.Resources) > 0 {
		firstResource := resourcesResult.Resources[0]
		fmt.Printf("Reading resource: %s\n", firstResource.URI)

		readResult, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: firstResource.URI,
		})
		if err != nil {
			return fmt.Errorf("failed to read resource: %w", err)
		}

		fmt.Printf("Resource contents: %d items\n", len(readResult.Contents))
		// Note: Resource content handling depends on the server implementation
	}

	fmt.Println()
	return nil
}

// demonstratePrompts shows how to list and get prompts
func demonstratePrompts(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("--- Prompts Demonstration ---\n")

	// List all available prompts
	promptsResult, err := session.ListPrompts(ctx, nil)
	if err != nil {
		// Prompts might not be available in basic server
		fmt.Println("Note: This server doesn't provide prompts")
		fmt.Println("Try connecting to the advanced server to see prompts\n")
		return nil
	}

	fmt.Printf("Available prompts: %d\n", len(promptsResult.Prompts))
	for i, prompt := range promptsResult.Prompts {
		fmt.Printf("%d. %s - %s\n", i+1, prompt.Name, prompt.Description)
		if len(prompt.Arguments) > 0 {
			fmt.Println("   Arguments:")
			for _, arg := range prompt.Arguments {
				required := ""
				if arg.Required {
					required = " (required)"
				}
				fmt.Printf("   - %s: %s%s\n", arg.Name, arg.Description, required)
			}
		}
	}
	fmt.Println()

	// Try to get the first prompt if available
	if len(promptsResult.Prompts) > 0 {
		firstPrompt := promptsResult.Prompts[0]
		fmt.Printf("Getting prompt: %s\n", firstPrompt.Name)

		getPromptParams := &mcp.GetPromptParams{
			Name:      firstPrompt.Name,
			Arguments: make(map[string]string),
		}

		// Add any required arguments
		for _, arg := range firstPrompt.Arguments {
			if arg.Required {
				// Provide a default value for demo
				getPromptParams.Arguments[arg.Name] = "example"
			}
		}

		promptResult, err := session.GetPrompt(ctx, getPromptParams)
		if err != nil {
			return fmt.Errorf("failed to get prompt: %w", err)
		}

		fmt.Printf("Prompt template (%d messages):\n", len(promptResult.Messages))
		for i, msg := range promptResult.Messages {
			fmt.Printf("  Message %d [%s]:\n", i+1, msg.Role)
			if textContent, ok := msg.Content.(*mcp.TextContent); ok {
				// Truncate long prompts for display
				text := textContent.Text
				if len(text) > 200 {
					text = text[:200] + "..."
				}
				fmt.Printf("    %s\n", text)
			}
		}
	}

	fmt.Println()
	return nil
}

// printToolResult is a helper function to print tool results nicely
func printToolResult(label string, result *mcp.CallToolResult) {
	fmt.Printf("  %s:\n", label)

	if len(result.Content) == 0 {
		fmt.Println("    (no content)")
		return
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			// Try to pretty print JSON
			var jsonData interface{}
			if json.Unmarshal([]byte(textContent.Text), &jsonData) == nil {
				prettyJSON, _ := json.MarshalIndent(jsonData, "    ", "  ")
				fmt.Printf("    %s\n", string(prettyJSON))
			} else {
				// Not JSON, print as-is
				fmt.Printf("    %s\n", textContent.Text)
			}
		}
	}
	fmt.Println()
}
