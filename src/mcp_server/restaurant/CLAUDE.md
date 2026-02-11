# Restaurant Booking MCP Server

A demonstration of the **Voice + MCP Integration Pattern** — a restaurant booking system using MCP that simulates multiple domain servers. See `src/mcp_server/CLAUDE.md` for general MCP architecture and the Voice + MCP pattern overview.

## Features

### Tools (8 total)

| Server Domain | Tool | Description |
|---------------|------|-------------|
| Restaurant | `search_restaurants` | Find by cuisine, location, dietary needs, price |
| Restaurant | `check_availability` | Get available time slots |
| Restaurant | `create_reservation` | Book a table (requires user approval) |
| Review | `get_restaurant_reviews` | Fetch reviews for decision-making |
| Calendar | `check_calendar` | Verify user availability |
| Calendar | `add_calendar_event` | Add reservation with reminders |
| Notification | `send_confirmation` | Send email/SMS confirmations |
| Maps | `get_distance_from_location` | Calculate distance and travel time |

### Resources (2 total)
- `resource://user_preferences` — favorite cuisines, dietary restrictions, preferred locations
- `resource://dining_history` — past reservations and restaurant visits

### Prompts (1 total)
- **`book_restaurant`** — orchestrated workflow with parameters: `cuisine` (req), `date` (req), `time` (req), `party_size` (req), `location`, `dietary_restrictions`, `occasion`

## Running

```bash
# Server (listens on stdio)
go run src/mcp_server/restaurant/main.go

# Demo client (simulates voice parameter collection)
go run src/mcp_server/restaurant/client_demo.go
```

The demo client simulates conversational parameter collection, invokes the `book_restaurant` prompt, and displays the workflow instructions.

## Workflow

When the `book_restaurant` prompt is invoked, it returns workflow instructions:

1. **GATHER CONTEXT** — Read `resource://user_preferences` and `resource://dining_history`, call `check_calendar`
2. **SEARCH & EVALUATE** — `search_restaurants` → `get_restaurant_reviews` → `check_availability` → `get_distance_from_location`
3. **PRESENT OPTIONS** — Show 2-3 best matches with ratings, reviews, distances, available times
4. **WAIT FOR SELECTION** — DO NOT book yet
5. **CONFIRM & BOOK** (after explicit approval) — `create_reservation` → `add_calendar_event` → `send_confirmation`
6. **FINAL SUMMARY** — Confirmation number, contact info, cancellation policy

### Safety Gates
- NEVER calls `create_reservation` without explicit user approval
- ALWAYS presents options before booking
- ALWAYS waits for confirmation before state-changing actions

## Sample Data

| Restaurant | Cuisine | Price | Rating | Location | Distance |
|-----------|---------|-------|--------|----------|----------|
| Bella Notte | Italian | $$$ | 4.7 | Downtown | 2.3 km |
| Trattoria Rosa | Italian | $$ | 4.5 | Downtown | 1.5 km |
| Il Giardino | Italian | $$$$ | 4.8 | Downtown | 3.1 km |
| Tokyo Sushi | Japanese | $$$ | 4.6 | Midtown | 2.8 km |
| Le Bistro | French | $$$ | 4.7 | Downtown | 1.9 km |

## Claude Desktop Integration

```json
{
  "mcpServers": {
    "restaurant-booking": {
      "command": "go",
      "args": ["run", "/absolute/path/to/src/mcp_server/restaurant/main.go"]
    }
  }
}
```

Config path — macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

## Testing Individual Tools

```go
// Search
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
            "date":          "2025-11-20",
            "party_size":    4,
        },
    },
})
```

## Extension Ideas

- Split into actual separate MCP servers (restaurant, calendar, notification)
- Connect to real databases (PostgreSQL, Redis)
- Integrate external APIs (OpenTable, Google Calendar, Twilio)
- Add authentication, waitlist management, table preferences

## Key Design Patterns

1. **Context-First**: Read resources (preferences, history) before searching
2. **Human-in-the-Loop**: Present options, get approval, then book
3. **Multi-Server Coordination**: Single workflow across domain servers
4. **Graceful Degradation**: Suggest alternatives when preferred option unavailable
