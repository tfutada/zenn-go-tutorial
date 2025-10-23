# Using MCP Servers in Claude Desktop

This guide explains how to use the MCP servers (basic, advanced, restaurant) that are installed in your Claude Desktop application.

## Installed Servers

Your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`) includes:

1. **go-tutorial-basic** - Simple calculator, echo, timestamp, weather tools
2. **go-tutorial-advanced** - Product/user management with search, filter, stats tools
3. **go-tutorial-restaurant** - Restaurant booking system with Voice + MCP integration pattern

## How MCP Servers Work

MCP servers expose three types of capabilities:

### 1. Tools (Functions)
Individual operations you can call directly.

**Example:**
```
Use the search_restaurants tool to find Italian restaurants in Downtown for 4 people with vegetarian options
```

### 2. Resources (Data Sources)
Readable data like user preferences, history, or configuration.

**Example:**
```
Read the resource://user_preferences resource to see my dining preferences
```

### 3. Prompts (Workflows)
Pre-defined orchestrated workflows that guide the AI through multi-step processes.

**Example:**
```
Execute the book_restaurant prompt to find and book a restaurant
```

## Restaurant Server: How to Use the book_restaurant Prompt

The restaurant server exposes a `book_restaurant` prompt - a complete orchestrated workflow for finding and booking restaurants.

### ✅ Confirmed: The Prompt IS Available

The Claude Desktop logs confirm that the prompt is being retrieved successfully:
```json
{"prompts":[{
  "name":"book_restaurant",
  "description":"Find and book a restaurant reservation with a guided workflow",
  "arguments": [...]
}]}
```

### How to Invoke It

The challenge is that Claude Desktop's AI needs explicit instructions on HOW to invoke prompts. Try these phrasings:

#### Method 1: Direct Execution Request
```
Execute the book_restaurant prompt from the go-tutorial-restaurant server with these parameters:
- cuisine: Italian
- date: 2025-11-20
- time: 19:00
- party_size: 4
- location: Downtown
- dietary_restrictions: vegetarian, gluten-free
- occasion: birthday

Follow the complete workflow instructions it returns.
```

#### Method 2: MCP Protocol Call
```
Call prompts/get on book_restaurant from go-tutorial-restaurant with arguments:
{
  "cuisine": "Italian",
  "date": "2025-11-20",
  "time": "19:00",
  "party_size": "4",
  "location": "Downtown",
  "dietary_restrictions": "vegetarian, gluten-free",
  "occasion": "birthday"
}

Then execute the workflow step by step.
```

#### Method 3: Natural Request (May Work)
```
I want to book an Italian restaurant for 4 people on November 20th at 7pm in Downtown.
We need vegetarian and gluten-free options, and it's for a birthday celebration.
Please use the book_restaurant prompt/workflow to help me find and book a place.
```

### What the Workflow Does

When the prompt is invoked, it returns detailed workflow instructions that guide Claude through:

**STEP 1: GATHER CONTEXT**
- Read `resource://user_preferences`
- Read `resource://dining_history`
- Call `check_calendar` tool

**STEP 2: SEARCH & EVALUATE**
- Call `search_restaurants` with filters
- Call `get_restaurant_reviews` for top results
- Call `check_availability` for each restaurant
- Call `get_distance_from_location` for travel time

**STEP 3: PRESENT OPTIONS**
- Show 2-3 best matches with all details
- Include ratings, reviews, distances, available times

**STEP 4: WAIT FOR SELECTION**
- Ask which restaurant you prefer
- **DO NOT book yet**

**STEP 5: CONFIRM & BOOK (Only After Approval)**
- Ask explicit confirmation
- Only if you say "yes", "confirm", "book it":
  - Call `create_reservation`
  - Call `add_calendar_event`
  - Call `send_confirmation`

**STEP 6: FINAL SUMMARY**
- Provide confirmation number
- Share contact info
- Explain cancellation policy

### Safety Features

The workflow includes critical safety gates:
- ⚠️ NEVER calls `create_reservation` without explicit user approval
- ⚠️ ALWAYS presents options before booking
- ⚠️ ALWAYS waits for confirmation before state-changing actions

## Using Individual Tools

If the prompt doesn't work, you can still use individual tools:

### Search for Restaurants
```
Call the search_restaurants tool with:
- cuisine: Italian
- location: Downtown
- party_size: 4
- dietary_filters: ["vegetarian", "gluten-free"]
```

### Check Availability
```
Call check_availability for restaurant ID "bella-notte" on 2025-11-20 for 4 people
```

### Get Reviews
```
Call get_restaurant_reviews for restaurant IDs ["bella-notte", "trattoria-rosa"]
```

## Verification

To verify the servers are working, check the logs:

```bash
# View restaurant server log
tail -f ~/Library/Logs/Claude/mcp-server-go-tutorial-restaurant.log

# View all MCP logs
tail -f ~/Library/Logs/Claude/mcp.log
```

You should see messages like:
```
Message from client: {"method":"prompts/list",...}
Message from server: {"prompts":[{"name":"book_restaurant",...}]}
```

## Troubleshooting

### If prompts don't work:
1. The server IS exposing prompts (verified in logs)
2. Try the direct phrasing examples above
3. Fall back to using individual tools

### If tools don't work:
1. Check that Claude Desktop was restarted after config changes
2. Check logs for errors: `~/Library/Logs/Claude/mcp-server-go-tutorial-restaurant.log`
3. Verify binary exists: `ls -l /Users/tafu/futasoft/golang/go-tutorial1/bin/mcp-server-restaurant`

## Test Client

For debugging, you can use the standalone test client:

```bash
cd /Users/tafu/futasoft/golang/go-tutorial1
go run src/mcp_client_restaurant/main.go
```

This will:
- Connect directly to the server
- List all capabilities (tools, resources, prompts)
- Retrieve the `book_restaurant` prompt
- Display the complete workflow instructions
- Demonstrate calling all tools

## Summary

✅ **Server Status**: All 3 servers installed and working
✅ **Prompts Exposed**: `book_restaurant` confirmed in logs
✅ **Tools Available**: 8 restaurant tools + 2 resources
⚠️ **Usage Note**: Claude Desktop AI may need explicit instructions to invoke prompts

Use the phrasing examples above to invoke the `book_restaurant` prompt workflow!
