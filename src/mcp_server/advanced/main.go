package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Advanced MCP Server Example
//
// This example demonstrates advanced MCP features including:
// - Tools: Functions that perform actions
// - Resources: Data sources that can be read
// - Prompts: Reusable prompt templates
//
// This server showcases how to build a more complex MCP server that can:
// 1. Provide access to data through resources
// 2. Offer reusable prompt templates
// 3. Implement sophisticated tool handlers with validation
//
// Running this server:
//   go run src/mcp_server/advanced/main.go

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	InStock  bool    `json:"in_stock"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

var (
	// In-memory storage for demonstration
	products []Product
	users    []User
	dataDir  string
)

func main() {
	// Initialize data directory path
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	dataDir = filepath.Join(wd, "src", "mcp_server", "advanced", "data")

	// Load initial data
	loadData()

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "advanced-mcp-server",
		Version: "1.0.0",
	}, nil)

	// Register Tools - functions that perform actions
	registerSearchTool(server)
	registerFilterTool(server)
	registerStatsTool(server)

	// Register Resources - data sources that can be read
	registerResources(server)

	// Register Prompts - reusable prompt templates
	registerPrompts(server)

	log.Println("Starting Advanced MCP Server...")
	log.Println("Features enabled:")
	log.Println("  - Tools: search, filter, stats")
	log.Println("  - Resources: users, products, system_info")
	log.Println("  - Prompts: analyze_data, generate_report")
	log.Println("Server ready to accept requests over stdio")

	// Run the server
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// loadData loads sample data from JSON files
func loadData() {
	// Load products
	productsFile := filepath.Join(dataDir, "products.json")
	if data, err := os.ReadFile(productsFile); err == nil {
		json.Unmarshal(data, &products)
		log.Printf("Loaded %d products from %s", len(products), productsFile)
	} else {
		log.Printf("Warning: Could not load products: %v", err)
		// Use default data if file doesn't exist
		products = []Product{
			{ID: 101, Name: "Laptop", Price: 999.99, Category: "Electronics", InStock: true},
			{ID: 102, Name: "Mouse", Price: 29.99, Category: "Accessories", InStock: true},
		}
	}

	// Load users
	usersFile := filepath.Join(dataDir, "users.json")
	if data, err := os.ReadFile(usersFile); err == nil {
		json.Unmarshal(data, &users)
		log.Printf("Loaded %d users from %s", len(users), usersFile)
	} else {
		log.Printf("Warning: Could not load users: %v", err)
		// Use default data if file doesn't exist
		users = []User{
			{ID: 1, Name: "Alice Smith", Email: "alice@example.com", Role: "Engineer"},
			{ID: 2, Name: "Bob Johnson", Email: "bob@example.com", Role: "Designer"},
		}
	}
}

// registerSearchTool adds a tool for searching across data
func registerSearchTool(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "search",
		Description: "Search for items by name or description across products and users",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query string",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Type of items to search",
					"enum":        []string{"all", "products", "users"},
				},
			},
			Required: []string{"query"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, ok := request.Params.Arguments["query"].(string)
		if !ok {
			return nil, fmt.Errorf("query must be a string")
		}
		query = strings.ToLower(query)

		searchType := "all"
		if t, exists := request.Params.Arguments["type"]; exists {
			searchType, _ = t.(string)
		}

		results := make(map[string]interface{})

		// Search products
		if searchType == "all" || searchType == "products" {
			var matchedProducts []Product
			for _, p := range products {
				if strings.Contains(strings.ToLower(p.Name), query) ||
					strings.Contains(strings.ToLower(p.Category), query) {
					matchedProducts = append(matchedProducts, p)
				}
			}
			results["products"] = matchedProducts
			results["product_count"] = len(matchedProducts)
		}

		// Search users
		if searchType == "all" || searchType == "users" {
			var matchedUsers []User
			for _, u := range users {
				if strings.Contains(strings.ToLower(u.Name), query) ||
					strings.Contains(strings.ToLower(u.Role), query) {
					matchedUsers = append(matchedUsers, u)
				}
			}
			results["users"] = matchedUsers
			results["user_count"] = len(matchedUsers)
		}

		// Format response
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal search results: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	mcp.AddTool(server, tool, handler)
}

// registerFilterTool adds a tool for filtering data with conditions
func registerFilterTool(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "filter",
		Description: "Filter products by various criteria (price range, category, availability)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"min_price": map[string]interface{}{
					"type":        "number",
					"description": "Minimum price (optional)",
				},
				"max_price": map[string]interface{}{
					"type":        "number",
					"description": "Maximum price (optional)",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Category to filter by (optional)",
				},
				"in_stock": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by stock availability (optional)",
				},
			},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments

		// Parse filter criteria
		var minPrice, maxPrice *float64
		if v, ok := args["min_price"].(float64); ok {
			minPrice = &v
		}
		if v, ok := args["max_price"].(float64); ok {
			maxPrice = &v
		}

		var category *string
		if v, ok := args["category"].(string); ok {
			category = &v
		}

		var inStock *bool
		if v, ok := args["in_stock"].(bool); ok {
			inStock = &v
		}

		// Apply filters
		var filtered []Product
		for _, p := range products {
			if minPrice != nil && p.Price < *minPrice {
				continue
			}
			if maxPrice != nil && p.Price > *maxPrice {
				continue
			}
			if category != nil && !strings.EqualFold(p.Category, *category) {
				continue
			}
			if inStock != nil && p.InStock != *inStock {
				continue
			}
			filtered = append(filtered, p)
		}

		result := map[string]interface{}{
			"filtered_products": filtered,
			"count":             len(filtered),
			"filters_applied": map[string]interface{}{
				"min_price": minPrice,
				"max_price": maxPrice,
				"category":  category,
				"in_stock":  inStock,
			},
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter results: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	mcp.AddTool(server, tool, handler)
}

// registerStatsTool adds a tool for computing statistics
func registerStatsTool(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "stats",
		Description: "Compute statistics about products and users",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"metric": map[string]interface{}{
					"type":        "string",
					"description": "Type of statistics to compute",
					"enum":        []string{"products", "users", "all"},
				},
			},
			Required: []string{"metric"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metric, ok := request.Params.Arguments["metric"].(string)
		if !ok {
			return nil, fmt.Errorf("metric must be a string")
		}

		stats := make(map[string]interface{})

		if metric == "products" || metric == "all" {
			var totalPrice float64
			var inStockCount int
			categories := make(map[string]int)

			for _, p := range products {
				totalPrice += p.Price
				if p.InStock {
					inStockCount++
				}
				categories[p.Category]++
			}

			avgPrice := 0.0
			if len(products) > 0 {
				avgPrice = totalPrice / float64(len(products))
			}

			stats["products"] = map[string]interface{}{
				"total_count":       len(products),
				"in_stock_count":    inStockCount,
				"out_of_stock":      len(products) - inStockCount,
				"average_price":     avgPrice,
				"categories":        categories,
				"category_count":    len(categories),
			}
		}

		if metric == "users" || metric == "all" {
			roles := make(map[string]int)
			for _, u := range users {
				roles[u.Role]++
			}

			stats["users"] = map[string]interface{}{
				"total_count": len(users),
				"roles":       roles,
				"role_count":  len(roles),
			}
		}

		jsonData, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stats: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	mcp.AddTool(server, tool, handler)
}

// registerResources adds MCP resources (data sources that can be read)
func registerResources(server *mcp.Server) {
	// Resource 1: Users list
	usersResource := &mcp.Resource{
		URI:         "resource://users",
		Name:        "Users List",
		Description: "Complete list of all users in the system",
		MimeType:    "application/json",
	}

	usersHandler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		jsonData, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal users: %w", err)
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      "resource://users",
					MimeType: "application/json",
				},
				Text: string(jsonData),
			},
		}, nil
	}

	mcp.AddResource(server, usersResource, usersHandler)

	// Resource 2: Products list
	productsResource := &mcp.Resource{
		URI:         "resource://products",
		Name:        "Products List",
		Description: "Complete list of all products with pricing and availability",
		MimeType:    "application/json",
	}

	productsHandler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		jsonData, err := json.MarshalIndent(products, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal products: %w", err)
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      "resource://products",
					MimeType: "application/json",
				},
				Text: string(jsonData),
			},
		}, nil
	}

	mcp.AddResource(server, productsResource, productsHandler)

	// Resource 3: System info
	systemResource := &mcp.Resource{
		URI:         "resource://system_info",
		Name:        "System Information",
		Description: "Current system status and metadata",
		MimeType:    "application/json",
	}

	systemHandler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
		info := map[string]interface{}{
			"server_name":    "advanced-mcp-server",
			"version":        "1.0.0",
			"product_count":  len(products),
			"user_count":     len(users),
			"data_directory": dataDir,
		}

		jsonData, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal system info: %w", err)
		}

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      "resource://system_info",
					MimeType: "application/json",
				},
				Text: string(jsonData),
			},
		}, nil
	}

	mcp.AddResource(server, systemResource, systemHandler)
}

// registerPrompts adds MCP prompts (reusable prompt templates)
func registerPrompts(server *mcp.Server) {
	// Prompt 1: Data Analysis
	analyzePrompt := &mcp.Prompt{
		Name:        "analyze_data",
		Description: "Analyze the current product and user data to provide insights",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "focus",
				Description: "What aspect to focus on",
				Required:    false,
			},
		},
	}

	analyzeHandler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		focus := "general"
		if f, ok := request.Params.Arguments["focus"].(string); ok {
			focus = f
		}

		promptText := fmt.Sprintf(`Please analyze the current data in this MCP server and provide insights.

Focus area: %s

Available data:
- %d products across various categories
- %d users with different roles

Use the 'stats' tool to get detailed statistics, then provide:
1. Key observations
2. Trends or patterns
3. Recommendations for improvement

Start by calling the stats tool with metric "all" to get comprehensive data.`,
			focus, len(products), len(users))

		return &mcp.GetPromptResult{
			Messages: []mcp.PromptMessage{
				{
					Role: "user",
					Content: mcp.TextContent{
						Type: "text",
						Text: promptText,
					},
				},
			},
		}, nil
	}

	mcp.AddPrompt(server, analyzePrompt, analyzeHandler)

	// Prompt 2: Generate Report
	reportPrompt := &mcp.Prompt{
		Name:        "generate_report",
		Description: "Generate a comprehensive report on products or users",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "report_type",
				Description: "Type of report to generate (products, users, or inventory)",
				Required:    true,
			},
		},
	}

	reportHandler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		reportType, ok := request.Params.Arguments["report_type"].(string)
		if !ok {
			return nil, fmt.Errorf("report_type is required and must be a string")
		}

		var promptText string
		switch reportType {
		case "products":
			promptText = `Generate a detailed product report by:
1. Using the 'stats' tool to get product statistics
2. Using the 'filter' tool to identify products by price ranges
3. Reading the 'resource://products' resource for complete data
4. Analyzing pricing trends, stock levels, and category distribution
5. Providing recommendations for inventory management`

		case "users":
			promptText = `Generate a detailed user report by:
1. Using the 'stats' tool to get user statistics
2. Reading the 'resource://users' resource for complete data
3. Analyzing role distribution and user engagement
4. Providing recommendations for team structure`

		case "inventory":
			promptText = `Generate a comprehensive inventory report by:
1. Using the 'filter' tool with in_stock=false to find out-of-stock items
2. Using the 'filter' tool with in_stock=true to check available inventory
3. Using the 'stats' tool for category-wise breakdown
4. Recommending which products to restock based on category popularity`

		default:
			return nil, fmt.Errorf("unknown report type: %s (valid: products, users, inventory)", reportType)
		}

		return &mcp.GetPromptResult{
			Messages: []mcp.PromptMessage{
				{
					Role: "user",
					Content: mcp.TextContent{
						Type: "text",
						Text: promptText,
					},
				},
			},
		}, nil
	}

	mcp.AddPrompt(server, reportPrompt, reportHandler)
}
