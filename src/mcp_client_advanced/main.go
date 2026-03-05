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

// Advanced MCP Client Example
//
// This client demonstrates how to interact with the advanced MCP server,
// showcasing all its features:
// 1. Tools: search, filter, stats
// 2. Resources: products, users, system_info
// 3. Prompts: analyze_data, generate_report
//
// The client automatically launches the advanced server and connects to it.
//
// Run this client:
//   go run src/mcp_client_advanced/main.go

func main() {
	fmt.Println("=== Advanced MCP Client ===\n")

	// Connect to the advanced MCP server
	serverCommand := "go"
	serverArgs := []string{"run", "src/mcp_server/advanced/main.go"}

	fmt.Printf("Connecting to advanced MCP server: %s %v\n\n", serverCommand, serverArgs)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "advanced-mcp-client",
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

	fmt.Println("✓ Connected to advanced MCP server successfully!\n")

	// Demonstrate all features
	if err := demonstrateTools(ctx, session); err != nil {
		log.Printf("Error demonstrating tools: %v", err)
	}

	if err := demonstrateResources(ctx, session); err != nil {
		log.Printf("Error demonstrating resources: %v", err)
	}

	if err := demonstratePrompts(ctx, session); err != nil {
		log.Printf("Error demonstrating prompts: %v", err)
	}

	fmt.Println("\n=== Advanced Client Session Complete ===")
}

// demonstrateTools shows how to use the advanced server's tools
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

	// Example 1: Search for products
	fmt.Println("Example 1: Search for products containing 'laptop'")
	searchResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "search",
		Arguments: map[string]interface{}{
			"query": "laptop",
			"type":  "products",
		},
	})
	if err != nil {
		return fmt.Errorf("search tool call failed: %w", err)
	}
	printToolResult("Search (laptop in products)", searchResult)

	// Example 2: Search for users by role
	fmt.Println("Example 2: Search for users with role 'engineer'")
	userSearchResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "search",
		Arguments: map[string]interface{}{
			"query": "engineer",
			"type":  "users",
		},
	})
	if err != nil {
		return fmt.Errorf("user search tool call failed: %w", err)
	}
	printToolResult("Search (engineer in users)", userSearchResult)

	// Example 3: Filter products by price range
	fmt.Println("Example 3: Filter products with price between $20 and $100")
	minPrice := 20.0
	maxPrice := 100.0
	filterResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "filter",
		Arguments: map[string]interface{}{
			"min_price": minPrice,
			"max_price": maxPrice,
		},
	})
	if err != nil {
		return fmt.Errorf("filter tool call failed: %w", err)
	}
	printToolResult("Filter (price $20-$100)", filterResult)

	// Example 4: Filter products by category
	fmt.Println("Example 4: Filter products in 'Electronics' category")
	filterCategoryResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "filter",
		Arguments: map[string]interface{}{
			"category": "Electronics",
		},
	})
	if err != nil {
		return fmt.Errorf("filter category tool call failed: %w", err)
	}
	printToolResult("Filter (Electronics)", filterCategoryResult)

	// Example 5: Filter in-stock products
	fmt.Println("Example 5: Filter only in-stock products")
	inStock := true
	filterStockResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "filter",
		Arguments: map[string]interface{}{
			"in_stock": inStock,
		},
	})
	if err != nil {
		return fmt.Errorf("filter stock tool call failed: %w", err)
	}
	printToolResult("Filter (in stock)", filterStockResult)

	// Example 6: Get product statistics
	fmt.Println("Example 6: Get product statistics")
	productStatsResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "stats",
		Arguments: map[string]interface{}{
			"metric": "products",
		},
	})
	if err != nil {
		return fmt.Errorf("product stats tool call failed: %w", err)
	}
	printToolResult("Stats (products)", productStatsResult)

	// Example 7: Get user statistics
	fmt.Println("Example 7: Get user statistics")
	userStatsResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "stats",
		Arguments: map[string]interface{}{
			"metric": "users",
		},
	})
	if err != nil {
		return fmt.Errorf("user stats tool call failed: %w", err)
	}
	printToolResult("Stats (users)", userStatsResult)

	// Example 8: Get all statistics
	fmt.Println("Example 8: Get comprehensive statistics")
	allStatsResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "stats",
		Arguments: map[string]interface{}{
			"metric": "all",
		},
	})
	if err != nil {
		return fmt.Errorf("all stats tool call failed: %w", err)
	}
	printToolResult("Stats (all)", allStatsResult)

	fmt.Println()
	return nil
}

// demonstrateResources shows how to read resources from the advanced server
func demonstrateResources(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("--- Resources Demonstration ---\n")

	// List all available resources
	resourcesResult, err := session.ListResources(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	fmt.Printf("Available resources: %d\n", len(resourcesResult.Resources))
	for i, resource := range resourcesResult.Resources {
		fmt.Printf("%d. %s (%s) - %s\n", i+1, resource.Name, resource.URI, resource.Description)
	}
	fmt.Println()

	// Read each resource
	for _, resource := range resourcesResult.Resources {
		fmt.Printf("Reading resource: %s\n", resource.URI)

		readResult, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: resource.URI,
		})
		if err != nil {
			log.Printf("  Failed to read resource %s: %v", resource.URI, err)
			continue
		}

		if len(readResult.Contents) > 0 {
			content := readResult.Contents[0]
			// Try to pretty print JSON
			var jsonData interface{}
			if json.Unmarshal([]byte(content.Text), &jsonData) == nil {
				prettyJSON, _ := json.MarshalIndent(jsonData, "  ", "  ")
				fmt.Printf("  Content (pretty JSON):\n  %s\n", string(prettyJSON))
			} else {
				fmt.Printf("  Content: %s\n", content.Text)
			}
		}
		fmt.Println()
	}

	return nil
}

// demonstratePrompts shows how to use prompts from the advanced server
func demonstratePrompts(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("--- Prompts Demonstration ---\n")

	// List all available prompts
	promptsResult, err := session.ListPrompts(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
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

	// Example 1: Get analyze_data prompt
	fmt.Println("Example 1: Get 'analyze_data' prompt with focus on 'inventory'")
	analyzePrompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "analyze_data",
		Arguments: map[string]string{
			"focus": "inventory",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get analyze_data prompt: %w", err)
	}
	printPromptResult("Analyze Data Prompt (focus: inventory)", analyzePrompt)

	// Example 2: Get generate_report prompt for products
	fmt.Println("Example 2: Get 'generate_report' prompt for products")
	reportPrompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "generate_report",
		Arguments: map[string]string{
			"report_type": "products",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get generate_report prompt: %w", err)
	}
	printPromptResult("Generate Report Prompt (type: products)", reportPrompt)

	// Example 3: Get generate_report prompt for users
	fmt.Println("Example 3: Get 'generate_report' prompt for users")
	userReportPrompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "generate_report",
		Arguments: map[string]string{
			"report_type": "users",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get user report prompt: %w", err)
	}
	printPromptResult("Generate Report Prompt (type: users)", userReportPrompt)

	// Example 4: Get generate_report prompt for inventory
	fmt.Println("Example 4: Get 'generate_report' prompt for inventory")
	inventoryReportPrompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "generate_report",
		Arguments: map[string]string{
			"report_type": "inventory",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get inventory report prompt: %w", err)
	}
	printPromptResult("Generate Report Prompt (type: inventory)", inventoryReportPrompt)

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

// printPromptResult is a helper function to print prompt results nicely
func printPromptResult(label string, result *mcp.GetPromptResult) {
	fmt.Printf("  %s:\n", label)
	fmt.Printf("  Messages: %d\n", len(result.Messages))

	for i, msg := range result.Messages {
		fmt.Printf("  Message %d [%s]:\n", i+1, msg.Role)
		if textContent, ok := msg.Content.(*mcp.TextContent); ok {
			// Indent the prompt text
			lines := splitLines(textContent.Text)
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}
	fmt.Println()
}

// splitLines splits text into lines for better formatting
func splitLines(text string) []string {
	var lines []string
	current := ""
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}