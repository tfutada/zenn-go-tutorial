package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Restaurant MCP Client - Demonstrates Prompt Retrieval and Usage
//
// This client shows how to:
// 1. List available prompts from the MCP server
// 2. Retrieve the book_restaurant prompt with parameters
// 3. Display the workflow instructions
// 4. Demonstrate calling tools
//
// Run this client:
//   go run src/mcp_client_restaurant/main.go

func main() {
	fmt.Println("=== Restaurant MCP Client ===")
	fmt.Println("Demonstrating Prompt Retrieval and Workflow\n")

	// Connect to the restaurant MCP server
	serverCommand := "go"
	serverArgs := []string{"run", "src/mcp_server/restaurant/main.go"}

	fmt.Printf("Connecting to restaurant MCP server: %s %v\n\n", serverCommand, serverArgs)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create the MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "restaurant-mcp-client",
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

	fmt.Println("✓ Connected to restaurant MCP server successfully!\n")

	// Demonstrate all features
	if err := listServerCapabilities(ctx, session); err != nil {
		log.Printf("Error listing capabilities: %v", err)
	}

	if err := demonstratePrompts(ctx, session); err != nil {
		log.Printf("Error demonstrating prompts: %v", err)
	}

	if err := demonstrateResources(ctx, session); err != nil {
		log.Printf("Error demonstrating resources: %v", err)
	}

	if err := demonstrateTools(ctx, session); err != nil {
		log.Printf("Error demonstrating tools: %v", err)
	}

	fmt.Println("\n=== Restaurant Client Session Complete ===")
}

// listServerCapabilities shows what the server provides
func listServerCapabilities(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("=== SERVER CAPABILITIES ===\n")

	// List tools
	toolsResult, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}
	fmt.Printf("✓ Tools: %d\n", len(toolsResult.Tools))
	for i, tool := range toolsResult.Tools {
		fmt.Printf("  %d. %s\n", i+1, tool.Name)
	}

	// List resources
	resourcesResult, err := session.ListResources(ctx, nil)
	if err != nil {
		fmt.Printf("✗ Resources: error listing - %v\n", err)
	} else {
		fmt.Printf("✓ Resources: %d\n", len(resourcesResult.Resources))
		for i, resource := range resourcesResult.Resources {
			fmt.Printf("  %d. %s (%s)\n", i+1, resource.Name, resource.URI)
		}
	}

	// List prompts
	promptsResult, err := session.ListPrompts(ctx, nil)
	if err != nil {
		fmt.Printf("✗ Prompts: error listing - %v\n", err)
	} else {
		fmt.Printf("✓ Prompts: %d\n", len(promptsResult.Prompts))
		for i, prompt := range promptsResult.Prompts {
			fmt.Printf("  %d. %s\n", i+1, prompt.Name)
		}
	}

	fmt.Println()
	return nil
}

// demonstratePrompts shows how to retrieve and use prompts
func demonstratePrompts(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("=== PROMPTS DEMONSTRATION ===\n")

	// List all available prompts
	promptsResult, err := session.ListPrompts(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	fmt.Printf("Available prompts: %d\n\n", len(promptsResult.Prompts))

	for i, prompt := range promptsResult.Prompts {
		fmt.Printf("--- Prompt %d: %s ---\n", i+1, prompt.Name)
		fmt.Printf("Description: %s\n", prompt.Description)

		if len(prompt.Arguments) > 0 {
			fmt.Println("Arguments:")
			for _, arg := range prompt.Arguments {
				required := ""
				if arg.Required {
					required = " (required)"
				}
				fmt.Printf("  - %s: %s%s\n", arg.Name, arg.Description, required)
			}
		}
		fmt.Println()
	}

	// Now retrieve the book_restaurant prompt with sample parameters
	fmt.Println("=== RETRIEVING book_restaurant PROMPT ===\n")

	fmt.Println("Parameters:")
	fmt.Println("  cuisine: Italian")
	fmt.Println("  date: 2025-11-20")
	fmt.Println("  time: 19:00")
	fmt.Println("  party_size: 4")
	fmt.Println("  location: Downtown")
	fmt.Println("  dietary_restrictions: vegetarian, gluten-free")
	fmt.Println("  occasion: birthday")
	fmt.Println()

	getPromptResult, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "book_restaurant",
		Arguments: map[string]string{
			"cuisine":               "Italian",
			"date":                  "2025-11-20",
			"time":                  "19:00",
			"party_size":            "4",
			"location":              "Downtown",
			"dietary_restrictions":  "vegetarian, gluten-free",
			"occasion":              "birthday",
		},
	})

	if err != nil {
		return fmt.Errorf("failed to get book_restaurant prompt: %w", err)
	}

	fmt.Printf("✓ Successfully retrieved book_restaurant prompt!\n")
	fmt.Printf("Messages in prompt: %d\n\n", len(getPromptResult.Messages))

	// Display the workflow instructions
	for i, msg := range getPromptResult.Messages {
		fmt.Printf("--- Message %d [Role: %s] ---\n", i+1, msg.Role)

		if textContent, ok := msg.Content.(*mcp.TextContent); ok {
			// Display the full workflow
			fmt.Println(textContent.Text)
		}
		fmt.Println()
	}

	return nil
}

// demonstrateResources shows how to read resources
func demonstrateResources(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("=== RESOURCES DEMONSTRATION ===\n")

	// List all available resources
	resourcesResult, err := session.ListResources(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	fmt.Printf("Available resources: %d\n\n", len(resourcesResult.Resources))

	// Read user preferences
	fmt.Println("Reading resource://user_preferences...")
	prefResult, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "resource://user_preferences",
	})
	if err != nil {
		log.Printf("Failed to read preferences: %v", err)
	} else if len(prefResult.Contents) > 0 {
		content := prefResult.Contents[0]
		var jsonData interface{}
		if json.Unmarshal([]byte(content.Text), &jsonData) == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "  ", "  ")
			fmt.Printf("  %s\n", string(prettyJSON))
		}
	}
	fmt.Println()

	// Read dining history
	fmt.Println("Reading resource://dining_history...")
	historyResult, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "resource://dining_history",
	})
	if err != nil {
		log.Printf("Failed to read history: %v", err)
	} else if len(historyResult.Contents) > 0 {
		content := historyResult.Contents[0]
		fmt.Printf("  %s\n", content.Text)
	}
	fmt.Println()

	return nil
}

// demonstrateTools shows how to call various restaurant tools
func demonstrateTools(ctx context.Context, session *mcp.ClientSession) error {
	fmt.Println("=== TOOLS DEMONSTRATION ===\n")

	// Example 1: Search for Italian restaurants
	fmt.Println("Example 1: Searching for Italian restaurants in Downtown")
	fmt.Println("Parameters: cuisine=Italian, location=Downtown, party_size=4, dietary_filters=[vegetarian, gluten-free]")

	searchResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "search_restaurants",
		Arguments: map[string]interface{}{
			"cuisine":         "Italian",
			"location":        "Downtown",
			"party_size":      4,
			"dietary_filters": []string{"vegetarian", "gluten-free"},
		},
	})
	if err != nil {
		return fmt.Errorf("search_restaurants failed: %w", err)
	}
	printToolResult("Search Results", searchResult)

	// Example 2: Get reviews for restaurants found
	fmt.Println("Example 2: Getting reviews for top restaurants")

	reviewResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_restaurant_reviews",
		Arguments: map[string]interface{}{
			"restaurant_ids": []string{"bella-notte", "trattoria-rosa"},
			"limit":          2,
		},
	})
	if err != nil {
		return fmt.Errorf("get_restaurant_reviews failed: %w", err)
	}
	printToolResult("Reviews", reviewResult)

	// Example 3: Check availability
	fmt.Println("Example 3: Checking availability at Bella Notte on 2025-11-20")

	availResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_availability",
		Arguments: map[string]interface{}{
			"restaurant_id": "bella-notte",
			"date":          "2025-11-20",
			"party_size":    4,
		},
	})
	if err != nil {
		return fmt.Errorf("check_availability failed: %w", err)
	}
	printToolResult("Availability", availResult)

	// Example 4: Check calendar
	fmt.Println("Example 4: Checking user's calendar for 2025-11-20")

	calendarResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_calendar",
		Arguments: map[string]interface{}{
			"date": "2025-11-20",
			"time": "19:00",
		},
	})
	if err != nil {
		return fmt.Errorf("check_calendar failed: %w", err)
	}
	printToolResult("Calendar Check", calendarResult)

	// Example 5: Get distance information
	fmt.Println("Example 5: Getting distance from user's location")

	distanceResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_distance_from_location",
		Arguments: map[string]interface{}{
			"destinations": []string{"bella-notte", "trattoria-rosa"},
			"from":         "user_home",
		},
	})
	if err != nil {
		return fmt.Errorf("get_distance_from_location failed: %w", err)
	}
	printToolResult("Distance Information", distanceResult)

	fmt.Println()
	fmt.Println("⚠️  NOTE: We're NOT calling create_reservation in this demo")
	fmt.Println("    The workflow requires explicit user approval before booking!")
	fmt.Println()

	return nil
}

// printToolResult is a helper function to print tool results nicely
func printToolResult(label string, result *mcp.CallToolResult) {
	fmt.Printf("\n--- %s ---\n", label)

	if len(result.Content) == 0 {
		fmt.Println("  (no content)")
		return
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			// Try to pretty print JSON
			var jsonData interface{}
			if json.Unmarshal([]byte(textContent.Text), &jsonData) == nil {
				prettyJSON, _ := json.MarshalIndent(jsonData, "  ", "  ")
				fmt.Printf("%s\n", string(prettyJSON))
			} else {
				// Not JSON, print as-is with indentation
				lines := strings.Split(textContent.Text, "\n")
				for _, line := range lines {
					fmt.Printf("  %s\n", line)
				}
			}
		}
	}
	fmt.Println()
}
