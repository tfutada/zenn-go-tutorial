# Restaurant Booking MCP Server

A demonstration of the **Voice + MCP Integration Pattern** - showing how voice interfaces (like OpenAI Realtime API) can collect parameters naturally through conversation and orchestrate complex multi-server workflows using the Model Context Protocol.

## Overview

This server implements a complete restaurant booking system using MCP that showcases:

- **Multi-Server Simulation**: Single server simulating multiple domain servers (restaurant, calendar, notification, review, maps)
- **Resources**: User preferences and dining history for context-aware decisions
- **Prompts**: Orchestrated workflows that guide AI through multi-step processes
- **Tools**: Specialized functions for searching, booking, and managing reservations

## Architecture

```
┌─────────────────────────────────────────────┐
│   Voice Interface Layer                     │
│   (OpenAI Realtime API / Voice Bot)         │
│   - Natural conversation                    │
│   - Parameter collection & validation       │
└──────────────────┬──────────────────────────┘
                   │ Structured parameters
                   ▼
┌─────────────────────────────────────────────┐
│   MCP Prompt Layer                          │
│   (book_restaurant prompt)                  │
│   - Workflow orchestration                  │
│   - Multi-tool coordination                 │
└──────────────────┬──────────────────────────┘
                   │ Tool calls
                   ▼
┌─────────────────────────────────────────────┐
│   MCP Tools Layer                           │
│   - Restaurant tools (search, book)         │
│   - Calendar tools (check, add event)       │
│   - Notification tools (send confirmation)  │
│   - Review tools (get reviews)              │
│   - Maps tools (calculate distance)         │
└─────────────────────────────────────────────┘
```

## Features

### Tools (8 total)

#### Restaurant Server
- **search_restaurants** - Find restaurants by cuisine, location, dietary needs, price range
- **check_availability** - Get available time slots for a specific restaurant
- **create_reservation** - Book a table (requires user approval)

#### Review Server
- **get_restaurant_reviews** - Fetch reviews to help decision-making

#### Calendar Server
- **check_calendar** - Verify user availability
- **add_calendar_event** - Add reservation to calendar with reminders

#### Notification Server
- **send_confirmation** - Send email/SMS confirmations

#### Maps Server
- **get_distance_from_location** - Calculate distance and travel time

### Resources (2 total)

- **resource://user_preferences** - User's favorite cuisines, dietary restrictions, preferred locations
- **resource://dining_history** - Past reservations and restaurant visits

### Prompts (1 total)

- **book_restaurant** - Complete workflow for finding and booking a restaurant with parameters:
  - `cuisine` (required) - Type of cuisine
  - `date` (required) - Reservation date (YYYY-MM-DD)
  - `time` (required) - Preferred time (HH:MM)
  - `party_size` (required) - Number of people
  - `location` (optional) - Area or neighborhood
  - `dietary_restrictions` (optional) - Dietary needs
  - `occasion` (optional) - Special occasion

## Installation

```bash
# Ensure you have Go 1.24+ installed
go version

# Install MCP SDK dependency
go get github.com/modelcontextprotocol/go-sdk
```

## Usage

### Running the Server

```bash
# From the repository root
go run src/mcp_server/restaurant/main.go
```

The server will start and listen for MCP requests over stdio.

### Running the Demo Client

The demo client simulates the voice parameter collection phase:

```bash
# In a separate terminal
go run src/mcp_server/restaurant/client_demo.go
```

This will:
1. Simulate conversational parameter collection (like a voice bot would)
2. Show the normalized JSON parameters
3. Connect to the MCP server
4. Invoke the `book_restaurant` prompt
5. Display the workflow instructions and available tools

### Example Interaction

```
🤖 Bot: What type of cuisine are you in the mood for?
👤 You: Italian

🤖 Bot: Great choice! When would you like to dine?
👤 You: 2025-11-20

🤖 Bot: Perfect! What time works for you on 2025-11-20?
👤 You: 19:00

🤖 Bot: How many people will be joining you?
👤 You: 4

🤖 Bot: Where would you like to dine? Any specific area?
👤 You: Downtown

🤖 Bot: Any dietary restrictions I should know about?
👤 You: one vegetarian, one gluten-free

🤖 Bot: Is this for a special occasion?
👤 You: birthday
```

## The Workflow

When the AI model receives the prompt, it follows this workflow:

### 1. GATHER CONTEXT (Resources)
- Read `resource://user_preferences`
- Read `resource://dining_history`
- Call `check_calendar` tool

### 2. SEARCH & EVALUATE (Tools)
- Call `search_restaurants` with filters
- Call `get_restaurant_reviews` for top results
- Call `check_availability` for each restaurant
- Call `get_distance_from_location` to calculate travel time

### 3. PRESENT OPTIONS
- Show 2-3 best matches with key details
- Include ratings, reviews, distances, available times

### 4. WAIT FOR SELECTION
- User selects preferred restaurant
- **DO NOT book yet**

### 5. CONFIRM & BOOK (After Approval)
- Call `create_reservation`
- Call `add_calendar_event`
- Call `send_confirmation`

### 6. FINAL SUMMARY
- Provide confirmation number
- Share contact info and cancellation policy

## Sample Data

The server includes sample data for 5 restaurants:

1. **Bella Notte** - Italian, $$$, Rating 4.7
   - Downtown, 2.3 km away
   - Vegetarian and gluten-free options
   - Birthday services available

2. **Trattoria Rosa** - Italian, $$, Rating 4.5
   - Downtown, 1.5 km away
   - Vegetarian, gluten-free, and vegan options
   - Family-friendly

3. **Il Giardino** - Italian, $$$$, Rating 4.8
   - Downtown, 3.1 km away
   - Fine dining with private dining room
   - Extensive dietary accommodations

4. **Tokyo Sushi** - Japanese, $$$, Rating 4.6
   - Midtown, 2.8 km away
   - Sushi bar and omakase

5. **Le Bistro** - French, $$$, Rating 4.7
   - Downtown, 1.9 km away
   - Authentic French cuisine

## Integration with Claude Desktop

You can integrate this server with Claude Desktop by adding to your configuration:

### macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
### Windows: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "restaurant-booking": {
      "command": "go",
      "args": ["run", "/path/to/go-tutorial1/src/mcp_server/restaurant/main.go"]
    }
  }
}
```

Then restart Claude Desktop and you'll be able to:
- Use the `book_restaurant` prompt
- Call individual tools like `search_restaurants`
- Read resources like user preferences

## Real-World Voice Integration

In a production system with OpenAI Realtime API:

```javascript
// 1. Voice conversation collects parameters
const params = await voiceBot.collectParameters({
  promptSchema: mcpClient.getPrompt("book_restaurant"),
  realtimeAPI: openai.realtime
});

// 2. Invoke MCP prompt
const workflow = await mcpClient.invokePrompt({
  name: "book_restaurant",
  arguments: params
});

// 3. AI model executes workflow
// - Reads resources for context
// - Calls tools to search and evaluate
// - Presents options via voice
// - Waits for user approval
// - Executes booking tools
// - Confirms via voice

// 4. User receives spoken confirmation
```

## Key Design Patterns

### 1. Context-First Approach
Always read resources (preferences, history) before searching to provide personalized results.

### 2. Human-in-the-Loop
Present options before booking. **Never** call `create_reservation` without explicit user approval.

### 3. Multi-Server Coordination
Single workflow coordinates multiple domain servers (restaurant, calendar, notification, etc.).

### 4. Graceful Error Handling
If preferred time unavailable, suggest alternatives. Handle "no" or "cancel" gracefully.

### 5. Natural Language Interface
Voice interface handles natural conversation, MCP handles structured orchestration.

## Testing Individual Tools

You can test individual tools using the MCP client:

```go
// Search for Italian restaurants
result, err := client.CallTool(ctx, &mcp.CallToolRequest{
    Params: mcp.CallToolRequestParams{
        Name: "search_restaurants",
        Arguments: map[string]interface{}{
            "cuisine": "Italian",
            "location": "Downtown",
            "party_size": 4,
            "dietary_filters": []string{"vegetarian", "gluten-free"},
        },
    },
})

// Check availability
result, err := client.CallTool(ctx, &mcp.CallToolRequest{
    Params: mcp.CallToolRequestParams{
        Name: "check_availability",
        Arguments: map[string]interface{}{
            "restaurant_id": "bella-notte",
            "date": "2025-11-20",
            "party_size": 4,
        },
    },
})
```

## Extension Ideas

This server can be extended to demonstrate:

1. **Actual Multi-Server Architecture**
   - Split into separate restaurant-server, calendar-server, notification-server
   - Have one client coordinate across all three

2. **Real Database Integration**
   - Connect to PostgreSQL or MongoDB for restaurants
   - Use Redis for caching availability

3. **External API Integration**
   - OpenTable or Yelp API for real restaurant data
   - Google Calendar API for actual calendar management
   - Twilio for SMS confirmations

4. **Authentication & Authorization**
   - OAuth for user identity
   - Permission scoping for different users

5. **Advanced Features**
   - Waitlist management
   - Table preferences (window seat, booth)
   - Pre-ordering options
   - Group booking coordination

## Related Documentation

- [VOICE_MCP_INTEGRATION.md](../../../VOICE_MCP_INTEGRATION.md) - Complete guide to Voice + MCP pattern
- [MCP Documentation](https://modelcontextprotocol.io/) - Official MCP documentation
- [OpenAI Realtime API](https://platform.openai.com/docs/guides/realtime) - Voice interface integration

## License

Part of the go-tutorial1 repository. See repository root for license information.

## Contributing

This is a demonstration project. Feel free to:
- Extend with additional tools
- Add more sophisticated search algorithms
- Integrate with real APIs
- Create multi-server variants

See the main repository README for contribution guidelines.
