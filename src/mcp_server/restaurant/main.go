package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Restaurant Booking MCP Server
//
// This server demonstrates the Voice + MCP integration pattern with:
// - Multi-server simulation (restaurant, calendar, notification, review, maps tools)
// - Resources for user preferences and history
// - Prompts for orchestrated workflows
//
// Designed to work with voice interfaces (like OpenAI Realtime API) that collect
// parameters naturally through conversation, then invoke structured MCP workflows.
//
// Running this server:
//   go run src/mcp_server/restaurant/main.go

// Domain Models

type Restaurant struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Cuisine             string   `json:"cuisine"`
	Location            string   `json:"location"`
	PriceRange          string   `json:"price_range"` // $, $$, $$$, $$$$
	Rating              float64  `json:"rating"`
	HasVegetarian       bool     `json:"has_vegetarian"`
	HasGlutenFree       bool     `json:"has_gluten_free"`
	HasVegan            bool     `json:"has_vegan"`
	Features            []string `json:"features"`
	DistanceKm          float64  `json:"distance_km"`
	BirthdayServices    bool     `json:"birthday_services"`
	PrivateDining       bool     `json:"private_dining"`
	AverageRating       float64  `json:"average_rating"`
	ReviewCount         int      `json:"review_count"`
	PhoneNumber         string   `json:"phone_number"`
	Address             string   `json:"address"`
	CancellationPolicy  string   `json:"cancellation_policy"`
}

type TimeSlot struct {
	Time      string `json:"time"`
	Available bool   `json:"available"`
}

type Reservation struct {
	ID              string    `json:"id"`
	RestaurantID    string    `json:"restaurant_id"`
	RestaurantName  string    `json:"restaurant_name"`
	Date            string    `json:"date"`
	Time            string    `json:"time"`
	PartySize       int       `json:"party_size"`
	SpecialRequests string    `json:"special_requests"`
	ConfirmationNum string    `json:"confirmation_number"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Review struct {
	RestaurantID string  `json:"restaurant_id"`
	Author       string  `json:"author"`
	Rating       float64 `json:"rating"`
	Text         string  `json:"text"`
	Date         string  `json:"date"`
	Helpful      int     `json:"helpful"`
}

type CalendarEvent struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Date        string   `json:"date"`
	Time        string   `json:"time"`
	Location    string   `json:"location"`
	Notes       string   `json:"notes"`
	Reminders   []string `json:"reminders"`
}

type UserPreferences struct {
	FavoriteCuisines    []string `json:"favorite_cuisines"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	PreferredPriceRange string   `json:"preferred_price_range"`
	PreferredLocations  []string `json:"preferred_locations"`
	PastReservations    []string `json:"past_reservations"`
}

// Tool Input/Output Types

// Restaurant Server Tools

type SearchRestaurantsInput struct {
	Cuisine        string   `json:"cuisine"`
	Location       string   `json:"location,omitempty"`
	PartySize      int      `json:"party_size"`
	DietaryFilters []string `json:"dietary_filters,omitempty"`
	MaxDistance    float64  `json:"max_distance,omitempty"`
	MinRating      float64  `json:"min_rating,omitempty"`
	PriceRange     string   `json:"price_range,omitempty"`
}

type SearchRestaurantsOutput struct {
	Restaurants []Restaurant `json:"restaurants"`
	Count       int          `json:"count"`
}

type CheckAvailabilityInput struct {
	RestaurantID string `json:"restaurant_id"`
	Date         string `json:"date"`
	PartySize    int    `json:"party_size"`
}

type CheckAvailabilityOutput struct {
	RestaurantName string     `json:"restaurant_name"`
	Date           string     `json:"date"`
	AvailableSlots []TimeSlot `json:"available_slots"`
}

type CreateReservationInput struct {
	RestaurantID    string `json:"restaurant_id"`
	Date            string `json:"date"`
	Time            string `json:"time"`
	PartySize       int    `json:"party_size"`
	SpecialRequests string `json:"special_requests,omitempty"`
}

type CreateReservationOutput struct {
	Reservation Reservation `json:"reservation"`
	Message     string      `json:"message"`
}

// Review Server Tools

type GetReviewsInput struct {
	RestaurantIDs []string `json:"restaurant_ids"`
	Limit         int      `json:"limit,omitempty"`
}

type GetReviewsOutput struct {
	Reviews map[string][]Review `json:"reviews"` // restaurant_id -> reviews
}

// Calendar Server Tools

type CheckCalendarInput struct {
	Date string `json:"date"`
	Time string `json:"time,omitempty"`
}

type CheckCalendarOutput struct {
	Date      string   `json:"date"`
	Available bool     `json:"available"`
	Conflicts []string `json:"conflicts,omitempty"`
}

type AddCalendarEventInput struct {
	Title     string   `json:"title"`
	Date      string   `json:"date"`
	Time      string   `json:"time"`
	Location  string   `json:"location"`
	Notes     string   `json:"notes,omitempty"`
	Reminders []string `json:"reminders,omitempty"`
}

type AddCalendarEventOutput struct {
	Event   CalendarEvent `json:"event"`
	Message string        `json:"message"`
}

// Notification Server Tools

type SendConfirmationInput struct {
	Type     string                 `json:"type"` // email, sms
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

type SendConfirmationOutput struct {
	Sent    bool   `json:"sent"`
	Message string `json:"message"`
}

// Maps Server Tools

type GetDistanceInput struct {
	Destinations []string `json:"destinations"` // restaurant IDs
	From         string   `json:"from"`         // user_home, user_work, or address
}

type GetDistanceOutput struct {
	Distances map[string]DistanceInfo `json:"distances"` // restaurant_id -> info
}

type DistanceInfo struct {
	RestaurantID   string  `json:"restaurant_id"`
	RestaurantName string  `json:"restaurant_name"`
	DistanceKm     float64 `json:"distance_km"`
	DrivingMinutes int     `json:"driving_minutes"`
	Address        string  `json:"address"`
}

// Global data stores
var (
	restaurants     []Restaurant
	reservations    []Reservation
	reviews         map[string][]Review
	calendarEvents  []CalendarEvent
	userPreferences UserPreferences
	dataDir         string
)

func main() {
	// Initialize data directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	dataDir = filepath.Join(wd, "src", "mcp_server", "restaurant", "data")

	// Load initial data
	loadData()

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "restaurant-booking-server",
		Version: "1.0.0",
	}, nil)

	// Register Tools
	registerRestaurantTools(server)
	registerReviewTools(server)
	registerCalendarTools(server)
	registerNotificationTools(server)
	registerMapsTools(server)

	// Register Resources
	registerResources(server)

	// Register Prompts
	registerPrompts(server)

	log.Println("Starting Restaurant Booking MCP Server...")
	log.Println("This server demonstrates the Voice + MCP integration pattern")
	log.Println("")
	log.Println("Features enabled:")
	log.Println("  Tools:")
	log.Println("    - Restaurant: search_restaurants, check_availability, create_reservation")
	log.Println("    - Review: get_restaurant_reviews")
	log.Println("    - Calendar: check_calendar, add_calendar_event")
	log.Println("    - Notification: send_confirmation")
	log.Println("    - Maps: get_distance_from_location")
	log.Println("  Resources:")
	log.Println("    - user_preferences, dining_history")
	log.Println("  Prompts:")
	log.Println("    - book_restaurant (orchestrated workflow)")
	log.Println("")
	log.Println("Server ready to accept requests over stdio")

	// Run the server
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func loadData() {
	// Initialize with sample data
	restaurants = []Restaurant{
		{
			ID: "bella-notte", Name: "Bella Notte", Cuisine: "Italian",
			Location: "Downtown", PriceRange: "$$$", Rating: 4.7,
			HasVegetarian: true, HasGlutenFree: true, HasVegan: false,
			Features:      []string{"Outdoor seating", "Wine bar", "Romantic"},
			DistanceKm:    2.3, BirthdayServices: true, PrivateDining: false,
			ReviewCount: 342, PhoneNumber: "(555) 123-4567",
			Address:            "123 Main St, Downtown",
			CancellationPolicy: "Free cancellation up to 24 hours before",
		},
		{
			ID: "trattoria-rosa", Name: "Trattoria Rosa", Cuisine: "Italian",
			Location: "Downtown", PriceRange: "$$", Rating: 4.5,
			HasVegetarian: true, HasGlutenFree: true, HasVegan: true,
			Features:   []string{"Family-friendly", "Quick service"},
			DistanceKm: 1.5, BirthdayServices: true, PrivateDining: false,
			ReviewCount: 218, PhoneNumber: "(555) 234-5678",
			Address:            "456 Oak Ave, Downtown",
			CancellationPolicy: "Free cancellation up to 12 hours before",
		},
		{
			ID: "il-giardino", Name: "Il Giardino", Cuisine: "Italian",
			Location: "Downtown", PriceRange: "$$$$", Rating: 4.8,
			HasVegetarian: true, HasGlutenFree: true, HasVegan: true,
			Features:   []string{"Fine dining", "Chef's table", "Garden seating"},
			DistanceKm: 3.1, BirthdayServices: true, PrivateDining: true,
			ReviewCount: 567, PhoneNumber: "(555) 345-6789",
			Address:            "789 Elm St, Downtown",
			CancellationPolicy: "Free cancellation up to 48 hours before",
		},
		{
			ID: "tokyo-sushi", Name: "Tokyo Sushi", Cuisine: "Japanese",
			Location: "Midtown", PriceRange: "$$$", Rating: 4.6,
			HasVegetarian: true, HasGlutenFree: true, HasVegan: false,
			Features:   []string{"Sushi bar", "Omakase", "Modern"},
			DistanceKm: 2.8, BirthdayServices: false, PrivateDining: false,
			ReviewCount: 423, PhoneNumber: "(555) 456-7890",
			Address:            "321 Pine St, Midtown",
			CancellationPolicy: "Free cancellation up to 24 hours before",
		},
		{
			ID: "le-bistro", Name: "Le Bistro", Cuisine: "French",
			Location: "Downtown", PriceRange: "$$$", Rating: 4.7,
			HasVegetarian: true, HasGlutenFree: false, HasVegan: false,
			Features:   []string{"Authentic French", "Wine selection", "Intimate"},
			DistanceKm: 1.9, BirthdayServices: true, PrivateDining: false,
			ReviewCount: 289, PhoneNumber: "(555) 567-8901",
			Address:            "654 Maple Ave, Downtown",
			CancellationPolicy: "Free cancellation up to 24 hours before",
		},
	}

	reviews = make(map[string][]Review)
	reviews["bella-notte"] = []Review{
		{
			RestaurantID: "bella-notte", Author: "Sarah M.", Rating: 5.0,
			Text:    "Amazing pasta and the birthday dessert was complimentary! Staff was so attentive.",
			Date:    "2025-10-10",
			Helpful: 23,
		},
		{
			RestaurantID: "bella-notte", Author: "John D.", Rating: 4.5,
			Text:    "Great gluten-free options. The risotto was perfect. A bit pricey but worth it.",
			Date:    "2025-10-05",
			Helpful: 18,
		},
	}

	reviews["trattoria-rosa"] = []Review{
		{
			RestaurantID: "trattoria-rosa", Author: "Emily R.", Rating: 4.5,
			Text:    "Excellent vegetarian lasagna and very accommodating staff.",
			Date:    "2025-10-12",
			Helpful: 15,
		},
	}

	reviews["il-giardino"] = []Review{
		{
			RestaurantID: "il-giardino", Author: "Michael B.", Rating: 5.0,
			Text:    "Private dining room was perfect for our celebration. Impeccable service.",
			Date:    "2025-10-08",
			Helpful: 31,
		},
	}

	userPreferences = UserPreferences{
		FavoriteCuisines:    []string{"Italian", "Japanese", "Mediterranean"},
		DietaryRestrictions: []string{},
		PreferredPriceRange: "$$-$$$",
		PreferredLocations:  []string{"Downtown", "Midtown"},
		PastReservations:    []string{"bella-notte", "tokyo-sushi"},
	}

	reservations = []Reservation{}
	calendarEvents = []CalendarEvent{}

	log.Printf("Loaded %d restaurants", len(restaurants))
	log.Printf("Loaded user preferences: %d favorite cuisines", len(userPreferences.FavoriteCuisines))
}

func registerRestaurantTools(server *mcp.Server) {
	// Tool: search_restaurants
	searchTool := &mcp.Tool{
		Name:        "search_restaurants",
		Description: "Search for restaurants based on cuisine, location, dietary needs, and other filters. Returns a list of matching restaurants with ratings and availability. WORKFLOW NOTE: Best used as part of the book_restaurant prompt workflow. Consider reading resource://user_preferences first for personalized results.",
	}

	searchHandler := func(ctx context.Context, request *mcp.CallToolRequest, input SearchRestaurantsInput) (*mcp.CallToolResult, SearchRestaurantsOutput, error) {
		var matched []Restaurant

		for _, r := range restaurants {
			// Filter by cuisine
			if !strings.EqualFold(r.Cuisine, input.Cuisine) {
				continue
			}

			// Filter by location
			if input.Location != "" && !strings.EqualFold(r.Location, input.Location) {
				continue
			}

			// Filter by dietary requirements
			if len(input.DietaryFilters) > 0 {
				hasAllDietary := true
				for _, dietary := range input.DietaryFilters {
					switch strings.ToLower(dietary) {
					case "vegetarian":
						if !r.HasVegetarian {
							hasAllDietary = false
						}
					case "gluten-free", "gluten_free":
						if !r.HasGlutenFree {
							hasAllDietary = false
						}
					case "vegan":
						if !r.HasVegan {
							hasAllDietary = false
						}
					}
				}
				if !hasAllDietary {
					continue
				}
			}

			// Filter by rating
			if input.MinRating > 0 && r.Rating < input.MinRating {
				continue
			}

			// Filter by distance
			if input.MaxDistance > 0 && r.DistanceKm > input.MaxDistance {
				continue
			}

			// Filter by price range
			if input.PriceRange != "" && r.PriceRange != input.PriceRange {
				continue
			}

			matched = append(matched, r)
		}

		return nil, SearchRestaurantsOutput{
			Restaurants: matched,
			Count:       len(matched),
		}, nil
	}

	mcp.AddTool(server, searchTool, searchHandler)

	// Tool: check_availability
	availTool := &mcp.Tool{
		Name:        "check_availability",
		Description: "Check available time slots for a specific restaurant on a given date. Returns list of available times.",
	}

	availHandler := func(ctx context.Context, request *mcp.CallToolRequest, input CheckAvailabilityInput) (*mcp.CallToolResult, CheckAvailabilityOutput, error) {
		// Find restaurant
		var restaurant *Restaurant
		for i := range restaurants {
			if restaurants[i].ID == input.RestaurantID {
				restaurant = &restaurants[i]
				break
			}
		}

		if restaurant == nil {
			return nil, CheckAvailabilityOutput{}, fmt.Errorf("restaurant not found: %s", input.RestaurantID)
		}

		// Generate mock availability (in real world, would query reservation system)
		times := []string{"17:00", "17:30", "18:00", "18:30", "19:00", "19:30", "20:00", "20:30", "21:00"}
		slots := make([]TimeSlot, len(times))

		for i, t := range times {
			// Randomly mark some as unavailable for demo
			available := rand.Float64() > 0.3
			slots[i] = TimeSlot{Time: t, Available: available}
		}

		return nil, CheckAvailabilityOutput{
			RestaurantName: restaurant.Name,
			Date:           input.Date,
			AvailableSlots: slots,
		}, nil
	}

	mcp.AddTool(server, availTool, availHandler)

	// Tool: create_reservation
	reserveTool := &mcp.Tool{
		Name:        "create_reservation",
		Description: "Create a restaurant reservation. ⚠️ CRITICAL: This is a state-changing action. Only call this tool AFTER: (1) searching for options, (2) presenting choices to user, (3) receiving explicit user approval (e.g., user says 'yes', 'confirm', 'book it'). Never call this tool directly without user confirmation. This creates a real reservation.",
	}

	reserveHandler := func(ctx context.Context, request *mcp.CallToolRequest, input CreateReservationInput) (*mcp.CallToolResult, CreateReservationOutput, error) {
		// Find restaurant
		var restaurant *Restaurant
		for i := range restaurants {
			if restaurants[i].ID == input.RestaurantID {
				restaurant = &restaurants[i]
				break
			}
		}

		if restaurant == nil {
			return nil, CreateReservationOutput{}, fmt.Errorf("restaurant not found: %s", input.RestaurantID)
		}

		// Create reservation
		reservation := Reservation{
			ID:              fmt.Sprintf("RES-%d", time.Now().Unix()),
			RestaurantID:    input.RestaurantID,
			RestaurantName:  restaurant.Name,
			Date:            input.Date,
			Time:            input.Time,
			PartySize:       input.PartySize,
			SpecialRequests: input.SpecialRequests,
			ConfirmationNum: fmt.Sprintf("%s-%d", strings.ToUpper(input.RestaurantID[:2]), rand.Intn(900000)+100000),
			Status:          "confirmed",
			CreatedAt:       time.Now(),
		}

		reservations = append(reservations, reservation)

		return nil, CreateReservationOutput{
			Reservation: reservation,
			Message:     fmt.Sprintf("Reservation confirmed at %s for %d people on %s at %s", restaurant.Name, input.PartySize, input.Date, input.Time),
		}, nil
	}

	mcp.AddTool(server, reserveTool, reserveHandler)
}

func registerReviewTools(server *mcp.Server) {
	reviewTool := &mcp.Tool{
		Name:        "get_restaurant_reviews",
		Description: "Get reviews for one or more restaurants. Helps users make informed decisions.",
	}

	reviewHandler := func(ctx context.Context, request *mcp.CallToolRequest, input GetReviewsInput) (*mcp.CallToolResult, GetReviewsOutput, error) {
		limit := input.Limit
		if limit == 0 {
			limit = 3
		}

		result := make(map[string][]Review)

		for _, rid := range input.RestaurantIDs {
			if reviewList, ok := reviews[rid]; ok {
				// Limit number of reviews
				if len(reviewList) > limit {
					result[rid] = reviewList[:limit]
				} else {
					result[rid] = reviewList
				}
			}
		}

		return nil, GetReviewsOutput{Reviews: result}, nil
	}

	mcp.AddTool(server, reviewTool, reviewHandler)
}

func registerCalendarTools(server *mcp.Server) {
	// Tool: check_calendar
	checkTool := &mcp.Tool{
		Name:        "check_calendar",
		Description: "Check user's calendar for availability on a specific date and time.",
	}

	checkHandler := func(ctx context.Context, request *mcp.CallToolRequest, input CheckCalendarInput) (*mcp.CallToolResult, CheckCalendarOutput, error) {
		// Check for conflicts
		var conflicts []string
		for _, event := range calendarEvents {
			if event.Date == input.Date {
				if input.Time == "" || event.Time == input.Time {
					conflicts = append(conflicts, event.Title)
				}
			}
		}

		return nil, CheckCalendarOutput{
			Date:      input.Date,
			Available: len(conflicts) == 0,
			Conflicts: conflicts,
		}, nil
	}

	mcp.AddTool(server, checkTool, checkHandler)

	// Tool: add_calendar_event
	addTool := &mcp.Tool{
		Name:        "add_calendar_event",
		Description: "Add an event to user's calendar with optional reminders.",
	}

	addHandler := func(ctx context.Context, request *mcp.CallToolRequest, input AddCalendarEventInput) (*mcp.CallToolResult, AddCalendarEventOutput, error) {
		event := CalendarEvent{
			ID:        fmt.Sprintf("CAL-%d", time.Now().Unix()),
			Title:     input.Title,
			Date:      input.Date,
			Time:      input.Time,
			Location:  input.Location,
			Notes:     input.Notes,
			Reminders: input.Reminders,
		}

		calendarEvents = append(calendarEvents, event)

		return nil, AddCalendarEventOutput{
			Event:   event,
			Message: fmt.Sprintf("Added '%s' to calendar for %s at %s", input.Title, input.Date, input.Time),
		}, nil
	}

	mcp.AddTool(server, addTool, addHandler)
}

func registerNotificationTools(server *mcp.Server) {
	notifyTool := &mcp.Tool{
		Name:        "send_confirmation",
		Description: "Send confirmation notification (email/SMS) to user.",
	}

	notifyHandler := func(ctx context.Context, request *mcp.CallToolRequest, input SendConfirmationInput) (*mcp.CallToolResult, SendConfirmationOutput, error) {
		// Mock sending notification
		log.Printf("Sending %s notification using template '%s'", input.Type, input.Template)
		log.Printf("Notification data: %+v", input.Data)

		return nil, SendConfirmationOutput{
			Sent:    true,
			Message: fmt.Sprintf("Confirmation sent via %s", input.Type),
		}, nil
	}

	mcp.AddTool(server, notifyTool, notifyHandler)
}

func registerMapsTools(server *mcp.Server) {
	mapsTool := &mcp.Tool{
		Name:        "get_distance_from_location",
		Description: "Calculate distance and driving time from user's location to restaurants.",
	}

	mapsHandler := func(ctx context.Context, request *mcp.CallToolRequest, input GetDistanceInput) (*mcp.CallToolResult, GetDistanceOutput, error) {
		result := make(map[string]DistanceInfo)

		for _, rid := range input.Destinations {
			for _, r := range restaurants {
				if r.ID == rid {
					// Calculate mock driving time (5 km/h avg in city + some variance)
					drivingMin := int(r.DistanceKm * 6)

					result[rid] = DistanceInfo{
						RestaurantID:   r.ID,
						RestaurantName: r.Name,
						DistanceKm:     r.DistanceKm,
						DrivingMinutes: drivingMin,
						Address:        r.Address,
					}
					break
				}
			}
		}

		return nil, GetDistanceOutput{Distances: result}, nil
	}

	mcp.AddTool(server, mapsTool, mapsHandler)
}

func registerResources(server *mcp.Server) {
	// Resource: User Preferences
	prefResource := &mcp.Resource{
		URI:         "resource://user_preferences",
		Name:        "User Dining Preferences",
		Description: "User's favorite cuisines, dietary restrictions, and location preferences",
		MIMEType:    "application/json",
	}

	prefHandler := func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		jsonData, err := json.MarshalIndent(userPreferences, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal preferences: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "resource://user_preferences",
					MIMEType: "application/json",
					Text:     string(jsonData),
				},
			},
		}, nil
	}

	server.AddResource(prefResource, prefHandler)

	// Resource: Dining History
	historyResource := &mcp.Resource{
		URI:         "resource://dining_history",
		Name:        "Dining History",
		Description: "User's past reservations and restaurant visits",
		MIMEType:    "application/json",
	}

	historyHandler := func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		jsonData, err := json.MarshalIndent(reservations, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal history: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "resource://dining_history",
					MIMEType: "application/json",
					Text:     string(jsonData),
				},
			},
		}, nil
	}

	server.AddResource(historyResource, historyHandler)
}

func registerPrompts(server *mcp.Server) {
	// Main prompt: book_restaurant
	bookPrompt := &mcp.Prompt{
		Name:        "book_restaurant",
		Description: "Find and book a restaurant reservation with a guided workflow",
		Arguments: []*mcp.PromptArgument{
			{Name: "cuisine", Description: "Type of cuisine (e.g., Italian, Japanese, Mexican)", Required: true},
			{Name: "date", Description: "Reservation date (YYYY-MM-DD)", Required: true},
			{Name: "time", Description: "Preferred time (HH:MM)", Required: true},
			{Name: "party_size", Description: "Number of people", Required: true},
			{Name: "location", Description: "Area or neighborhood", Required: false},
			{Name: "dietary_restrictions", Description: "Any dietary needs (e.g., vegan, gluten-free)", Required: false},
			{Name: "occasion", Description: "Special occasion (e.g., birthday, anniversary)", Required: false},
		},
	}

	bookHandler := func(ctx context.Context, request *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		args := request.Params.Arguments

		cuisine := args["cuisine"]
		date := args["date"]
		time := args["time"]
		partySize := args["party_size"]
		location := args["location"]
		dietary := args["dietary_restrictions"]
		occasion := args["occasion"]

		workflowText := fmt.Sprintf(`Please help me book a %s restaurant for %s people on %s at %s.

Location preference: %s
Dietary requirements: %s
Special occasion: %s

⚠️ CRITICAL WORKFLOW - FOLLOW STRICTLY IN THIS EXACT ORDER:

STEP 1 - GATHER CONTEXT (DO THIS FIRST, BEFORE ANY TOOL CALLS):
   YOU MUST read these resources BEFORE proceeding to step 2:
   - Read resource://user_preferences (to understand my dining preferences)
   - Read resource://dining_history (to see my past reservations)
   - Use check_calendar tool (to verify I'm available on %s at %s)

   ⚠️ DO NOT proceed to step 2 until you have read both resources and checked calendar.

STEP 2 - SEARCH & EVALUATE:
   Now use these tools in sequence:
   - Call search_restaurants with:
     * cuisine: %s
     * location: %s
     * party_size: %s
     * dietary_filters: [%s]

   - For the top 3-5 restaurants found, call get_restaurant_reviews

   - Call check_availability for each restaurant on %s

   - Call get_distance_from_location to calculate travel time from user_home

STEP 3 - PRESENT OPTIONS:
   Show me 2-3 best matches with:
     * Name, rating, price range
     * Distance and estimated travel time
     * Available time slots (if %s not available, suggest alternatives)
     * Highlights from reviews
     * Confirmation they meet dietary needs: %s
     * Special services if this is a %s occasion

   Format as a clear numbered list with key details.

   ⚠️ DO NOT PROCEED TO STEP 4 YET.

STEP 4 - WAIT FOR MY SELECTION:
   Ask me which restaurant I prefer.

   ⚠️ CRITICAL: DO NOT call create_reservation
   ⚠️ CRITICAL: DO NOT call any booking tools
   ⚠️ CRITICAL: DO NOT make any reservations
   ⚠️ WAIT for my response selecting a restaurant

STEP 5 - CONFIRM & BOOK (ONLY AFTER MY EXPLICIT APPROVAL):
   After I select a restaurant, confirm the details once more:
   Ask: "Should I go ahead and book [restaurant name] for [party_size] people on [date] at [time]?"

   ⚠️ ONLY if I say "yes", "confirm", "book it", "go ahead", or similar explicit approval:
   - Call create_reservation with:
     * restaurant_id: [selected restaurant]
     * date: %s
     * time: %s
     * party_size: %s
     * special_requests: Include occasion (%s) and dietary needs (%s)

   - Call add_calendar_event:
     * title: "Dinner at [restaurant name]"
     * Include confirmation number in notes
     * reminders: ["1 day before", "2 hours before"]

   - Call send_confirmation:
     * type: "email"
     * template: "restaurant_booking"
     * Include all reservation details

   ⚠️ If I say "no", "cancel", "wait", "let me think":
   - DO NOT call any booking tools
   - Ask what I'd like to do instead

STEP 6 - FINAL SUMMARY:
   - Provide confirmation number
   - Restaurant contact info
   - Cancellation policy
   - Offer to help with anything else (parking, pre-ordering dessert, etc.)

⚠️ SAFETY RULES (NEVER VIOLATE THESE):
1. NEVER call create_reservation without explicit user approval
2. ALWAYS read resources BEFORE calling search_restaurants
3. ALWAYS present options BEFORE booking
4. ALWAYS wait for user confirmation before state-changing actions
5. If uncertain whether user approved, ASK AGAIN - better to double-check than make unwanted reservation

If you are uncertain at any point whether I've given approval, ASK AGAIN.
Be conversational and helpful throughout, but ALWAYS respect the approval gates.`,
			cuisine, partySize, date, time,
			location, dietary, occasion,
			date, time,
			cuisine, location, partySize, dietary,
			date,
			time, dietary, occasion,
			date, time, partySize, occasion, dietary)

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: workflowText,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(bookPrompt, bookHandler)
}
