package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Voice + MCP Client Demo
//
// This client simulates the voice interface layer that:
// 1. Conducts a conversational dialogue to collect parameters
// 2. Normalizes the collected data into structured format
// 3. Invokes the MCP prompt with the parameters
// 4. Displays the workflow instructions
//
// In a real implementation, this would integrate with OpenAI Realtime API
// or another voice interface, and the workflow would be executed by an AI model.
//
// Running this demo:
//   # Terminal 1 - Start the server
//   go run src/mcp_server/restaurant/main.go
//
//   # Terminal 2 - Run the client
//   go run src/mcp_server/restaurant/client_demo.go

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  Restaurant Booking - Voice + MCP Integration Demo          ║")
	fmt.Println("║  This simulates voice parameter collection → MCP workflow    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Phase 1: Voice Parameter Collection (Simulated)
	fmt.Println("═══ PHASE 1: Voice Parameter Collection ═══")
	fmt.Println("(Simulating conversational voice interface)")
	fmt.Println()

	params := collectParametersViaVoice()

	// Phase 2: Display collected parameters
	fmt.Println("\n═══ PHASE 2: Parameters Extracted & Normalized ═══")
	displayParameters(params)

	// Phase 3: Invoke MCP Prompt
	fmt.Println("\n═══ PHASE 3: Invoking MCP Prompt ═══")
	fmt.Println("Connecting to restaurant-booking-server...")
	fmt.Println()

	// Ask if user wants to proceed with actual MCP call
	fmt.Print("Would you like to invoke the MCP prompt with these parameters? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" {
		invokeMCPPrompt(params)
	} else {
		fmt.Println("\nDemo completed. Parameters collected successfully!")
		fmt.Println("\nIn a real implementation:")
		fmt.Println("1. Voice interface collects parameters naturally")
		fmt.Println("2. AI model receives the prompt workflow")
		fmt.Println("3. AI orchestrates multi-server tool calls")
		fmt.Println("4. Results are presented back via voice")
	}
}

type BookingParams struct {
	Cuisine             string
	Date                string
	Time                string
	PartySize           string
	Location            string
	DietaryRestrictions string
	Occasion            string
}

func collectParametersViaVoice() BookingParams {
	reader := bufio.NewReader(os.Stdin)
	params := BookingParams{}

	// Simulate conversational parameter collection
	fmt.Println("🤖 Bot: Hi! I'd be happy to help you book a restaurant!")
	fmt.Println()

	// Cuisine
	fmt.Println("🤖 Bot: What type of cuisine are you in the mood for?")
	fmt.Print("👤 You: ")
	cuisine, _ := reader.ReadString('\n')
	params.Cuisine = strings.TrimSpace(cuisine)
	fmt.Println()

	// Date
	fmt.Println("🤖 Bot: Great choice! When would you like to dine?")
	fmt.Println("     (Please enter date as YYYY-MM-DD, e.g., 2025-11-20)")
	fmt.Print("👤 You: ")
	date, _ := reader.ReadString('\n')
	params.Date = strings.TrimSpace(date)
	fmt.Println()

	// Time
	fmt.Printf("🤖 Bot: Perfect! What time works for you on %s?\n", params.Date)
	fmt.Println("     (Please enter time as HH:MM, e.g., 19:00)")
	fmt.Print("👤 You: ")
	time, _ := reader.ReadString('\n')
	params.Time = strings.TrimSpace(time)
	fmt.Println()

	// Party size
	fmt.Println("🤖 Bot: How many people will be joining you?")
	fmt.Print("👤 You: ")
	partySize, _ := reader.ReadString('\n')
	params.PartySize = strings.TrimSpace(partySize)
	fmt.Println()

	// Location
	fmt.Println("🤖 Bot: Where would you like to dine? Any specific area?")
	fmt.Println("     (e.g., Downtown, Midtown, or just press Enter to skip)")
	fmt.Print("👤 You: ")
	location, _ := reader.ReadString('\n')
	params.Location = strings.TrimSpace(location)
	fmt.Println()

	// Dietary restrictions
	fmt.Println("🤖 Bot: Any dietary restrictions I should know about?")
	fmt.Println("     (e.g., vegetarian, gluten-free, vegan, or press Enter to skip)")
	fmt.Print("👤 You: ")
	dietary, _ := reader.ReadString('\n')
	params.DietaryRestrictions = strings.TrimSpace(dietary)
	fmt.Println()

	// Occasion
	fmt.Println("🤖 Bot: Is this for a special occasion?")
	fmt.Println("     (e.g., birthday, anniversary, or press Enter to skip)")
	fmt.Print("👤 You: ")
	occasion, _ := reader.ReadString('\n')
	params.Occasion = strings.TrimSpace(occasion)
	fmt.Println()

	fmt.Println("🤖 Bot: Excellent! Let me find the best restaurants for you...")
	return params
}

func displayParameters(params BookingParams) {
	fmt.Println("✓ Collected parameters:")
	fmt.Printf("  • Cuisine: %s\n", params.Cuisine)
	fmt.Printf("  • Date: %s\n", params.Date)
	fmt.Printf("  • Time: %s\n", params.Time)
	fmt.Printf("  • Party Size: %s\n", params.PartySize)
	if params.Location != "" {
		fmt.Printf("  • Location: %s\n", params.Location)
	}
	if params.DietaryRestrictions != "" {
		fmt.Printf("  • Dietary Restrictions: %s\n", params.DietaryRestrictions)
	}
	if params.Occasion != "" {
		fmt.Printf("  • Occasion: %s\n", params.Occasion)
	}
	fmt.Println()

	fmt.Println("📊 Normalized JSON:")
	jsonData, _ := json.MarshalIndent(map[string]string{
		"cuisine":              params.Cuisine,
		"date":                 params.Date,
		"time":                 params.Time,
		"party_size":           params.PartySize,
		"location":             params.Location,
		"dietary_restrictions": params.DietaryRestrictions,
		"occasion":             params.Occasion,
	}, "  ", "  ")
	fmt.Println(string(jsonData))
}

func invokeMCPPrompt(params BookingParams) {
	fmt.Println("\n📡 Connecting to MCP server...")

	// Create the MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "restaurant-booking-client",
		Version: "1.0.0",
	}, nil)

	// Create a transport that launches the server process
	transport := &mcp.CommandTransport{
		Command: exec.Command("go", "run", "src/mcp_server/restaurant/main.go"),
	}

	// Connect to the server
	ctx := context.Background()
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer session.Close()

	fmt.Println("✓ Connected to restaurant-booking-server")
	fmt.Println()

	// List available prompts
	fmt.Println("🔍 Discovering available prompts...")
	promptsResp, err := session.ListPrompts(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list prompts: %v", err)
	}

	fmt.Printf("✓ Found %d prompt(s):\n", len(promptsResp.Prompts))
	for _, p := range promptsResp.Prompts {
		fmt.Printf("  - %s: %s\n", p.Name, p.Description)
	}
	fmt.Println()

	// Get the book_restaurant prompt with our parameters
	fmt.Println("📝 Invoking 'book_restaurant' prompt with collected parameters...")

	arguments := map[string]string{
		"cuisine":    params.Cuisine,
		"date":       params.Date,
		"time":       params.Time,
		"party_size": params.PartySize,
	}

	if params.Location != "" {
		arguments["location"] = params.Location
	}
	if params.DietaryRestrictions != "" {
		arguments["dietary_restrictions"] = params.DietaryRestrictions
	}
	if params.Occasion != "" {
		arguments["occasion"] = params.Occasion
	}

	promptResp, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "book_restaurant",
		Arguments: arguments,
	})
	if err != nil {
		log.Fatalf("Failed to get prompt: %v", err)
	}

	fmt.Println("✓ Received workflow instructions from MCP server")
	fmt.Println()

	// Display the prompt messages
	fmt.Println("═══ WORKFLOW INSTRUCTIONS (from MCP Prompt) ═══")
	fmt.Println()
	for i, msg := range promptResp.Messages {
		fmt.Printf("Message %d (Role: %s):\n", i+1, msg.Role)
		fmt.Println("─────────────────────────────────────────────────────────────")
		if textContent, ok := msg.Content.(*mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		}
		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Println()
	}

	// List available tools
	fmt.Println("═══ AVAILABLE TOOLS (from MCP Server) ═══")
	toolsResp, err := session.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Printf("\nThe AI model can now use these %d tools:\n\n", len(toolsResp.Tools))
	for _, tool := range toolsResp.Tools {
		fmt.Printf("  🔧 %s\n", tool.Name)
		fmt.Printf("     %s\n\n", tool.Description)
	}

	// List available resources
	fmt.Println("═══ AVAILABLE RESOURCES (from MCP Server) ═══")
	resourcesResp, err := session.ListResources(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list resources: %v", err)
	}

	fmt.Printf("\nThe AI model can read these %d resources:\n\n", len(resourcesResp.Resources))
	for _, res := range resourcesResp.Resources {
		fmt.Printf("  📄 %s\n", res.URI)
		fmt.Printf("     %s\n\n", res.Description)
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("✅ DEMO COMPLETE!")
	fmt.Println()
	fmt.Println("What happens next in a real Voice + MCP system:")
	fmt.Println()
	fmt.Println("1. The AI model receives the workflow instructions above")
	fmt.Println("2. It reads resources to gather context about user preferences")
	fmt.Println("3. It calls search_restaurants tool to find options")
	fmt.Println("4. It calls get_restaurant_reviews to evaluate options")
	fmt.Println("5. It presents 2-3 best matches via VOICE")
	fmt.Println("6. User selects one via VOICE")
	fmt.Println("7. After user approval, it calls create_reservation")
	fmt.Println("8. It calls add_calendar_event to add to calendar")
	fmt.Println("9. It calls send_confirmation to email confirmation")
	fmt.Println("10. Final summary is read via VOICE")
	fmt.Println()
	fmt.Println("This demonstrates the three-layer architecture:")
	fmt.Println("  Voice Interface → MCP Prompts → Multi-Server Tools")
	fmt.Println()
}
