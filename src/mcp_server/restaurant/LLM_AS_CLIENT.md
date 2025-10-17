# The LLM as MCP Client

A key insight about the Voice + MCP integration architecture: **The LLM (like Claude Sonnet) IS the MCP client**.

## Architecture Overview

```
┌─────────────────────────────────────┐
│  Human User                         │
│  (speaks naturally)                 │
└──────────────┬──────────────────────┘
               │ voice
               ▼
┌─────────────────────────────────────┐
│  Voice Interface                    │
│  (OpenAI Realtime API)              │
│  - Collects parameters              │
│  - Normalizes to JSON               │
└──────────────┬──────────────────────┘
               │ structured params
               ▼
┌─────────────────────────────────────┐
│  LLM (Claude Sonnet) = MCP CLIENT   │  ← The LLM is the client!
│  - Receives prompt workflow         │
│  - Reads resources for context      │
│  - Calls tools intelligently        │
│  - Makes decisions                  │
│  - Orchestrates multi-step flow     │
└──────────────┬──────────────────────┘
               │ tool calls
               ▼
┌─────────────────────────────────────┐
│  MCP Server(s)                      │
│  - restaurant-booking-server        │
│  - calendar-server                  │
│  - notification-server              │
│  - review-server                    │
│  - maps-server                      │
└─────────────────────────────────────┘
```

## What the LLM Does as MCP Client

When the LLM (Claude) receives the `book_restaurant` prompt, it acts as an intelligent client that:

### 1. Receives Workflow Instructions

The prompt tells Claude exactly what to do:
```
1. GATHER CONTEXT - Read resources
2. SEARCH & EVALUATE - Call search and review tools
3. PRESENT OPTIONS - Show results to user
4. WAIT FOR SELECTION - Don't book yet!
5. CONFIRM & BOOK - After approval, execute booking tools
6. FINAL SUMMARY - Confirm details
```

### 2. Executes Tools Intelligently

Claude decides which tools to call and in what order:

```javascript
// Claude's internal decision-making
1. Read resource://user_preferences
   → Sees user likes Italian, prefers Downtown, past visits

2. Call search_restaurants({
     cuisine: "Italian",
     location: "downtown",
     party_size: 4,
     dietary_filters: ["vegetarian", "gluten-free"]
   })
   → Gets 3 restaurants

3. Call get_restaurant_reviews({
     restaurant_ids: ["bella-notte", "trattoria-rosa", "il-giardino"]
   })
   → Analyzes reviews

4. Call check_availability({
     restaurant_id: "bella-notte",
     date: "2025-11-20",
     party_size: 4
   })
   → Checks time slots

5. Call get_distance_from_location({
     destinations: ["bella-notte", "trattoria-rosa"],
     from: "user_home"
   })
   → Calculates travel times
```

### 3. Makes Intelligent Decisions

- Filters results based on user preferences
- Prioritizes restaurants with better reviews
- Considers availability and distance
- Adapts if preferred time is unavailable

### 4. Maintains Conversation

Claude speaks naturally with the user throughout:
- "I found 3 great options for you..."
- "Which would you prefer?"
- "Should I go ahead and book?"

### 5. Respects Approval Gates

**CRITICAL**: Claude never calls `create_reservation` until user explicitly approves:
```
User: "Bella Notte sounds perfect"
Claude: "Should I book Bella Notte for 4 at 7pm?" ← Asking!

User: "Yes"
Claude: *NOW calls create_reservation* ← Only after approval
```

## Example: Complete Flow

### Voice Collection Phase

```
User: "I need to book an Italian restaurant"

Voice Bot: "Great! When would you like to dine?"
User: "Tomorrow at 7pm"

Voice Bot: "How many people?"
User: "Four of us"

Voice Bot: "Any dietary restrictions?"
User: "One vegetarian, one gluten-free"

Voice Bot: "Is this for a special occasion?"
User: "Yes, a birthday"
```

### LLM Execution Phase

Claude receives the prompt with parameters:
```json
{
  "cuisine": "Italian",
  "date": "2025-11-20",
  "time": "19:00",
  "party_size": "4",
  "dietary_restrictions": "one vegetarian, one gluten-free",
  "occasion": "birthday"
}
```

Claude then executes:

#### Step 1: Context Gathering
```
Claude → MCP Server: ReadResource(resource://user_preferences)
MCP Server → Claude: {
  "favorite_cuisines": ["Italian", "Japanese"],
  "past_reservations": ["bella-notte", "tokyo-sushi"]
}

Claude thinks: "User has been to Bella Notte before and likes Italian"
```

#### Step 2: Search
```
Claude → MCP Server: search_restaurants({
  cuisine: "Italian",
  location: "downtown",
  party_size: 4,
  dietary_filters: ["vegetarian", "gluten-free"]
})

MCP Server → Claude: {
  restaurants: [
    {id: "bella-notte", rating: 4.7, ...},
    {id: "trattoria-rosa", rating: 4.5, ...},
    {id: "il-giardino", rating: 4.8, ...}
  ]
}
```

#### Step 3: Reviews & Availability
```
Claude → MCP Server: get_restaurant_reviews(["bella-notte", ...])
Claude → MCP Server: check_availability("bella-notte", "2025-11-20", 4)
Claude → MCP Server: get_distance_from_location(["bella-notte", ...])
```

#### Step 4: Present Options (via Voice)
```
Voice Bot (speaking for Claude):
"I found 3 excellent Italian restaurants downtown:

1. Bella Notte - 4.7 stars, $$$
   Available at 7:00 PM
   Excellent gluten-free pasta and vegetarian options
   They offer complimentary birthday dessert
   12 minutes from home

2. Trattoria Rosa - 4.5 stars, $$
   Available at 6:45 PM or 7:30 PM
   Known for their vegetarian menu
   8 minutes from home

3. Il Giardino - 4.8 stars, $$$$
   Available at 7:15 PM
   Private dining room available
   15 minutes from home

Which would you prefer?"
```

#### Step 5: Wait for User Selection
```
User (via voice): "Bella Notte sounds perfect"

Claude thinks: "User selected Bella Notte, but I should confirm
before making the reservation"

Voice Bot: "Excellent choice! Should I go ahead and book
Bella Notte for 4 people at 7:00 PM on November 20th?
I'll also add it to your calendar and let them know about
the birthday celebration and dietary needs."

User: "Yes, please"
```

#### Step 6: Execute Booking (After Approval)
```
Claude → MCP Server: create_reservation({
  restaurant_id: "bella-notte",
  date: "2025-11-20",
  time: "19:00",
  party_size: 4,
  special_requests: "Birthday celebration, one vegetarian, one gluten-free"
})

MCP Server → Claude: {
  confirmation_number: "BN-482934",
  status: "confirmed"
}

Claude → MCP Server: add_calendar_event({
  title: "Dinner at Bella Notte",
  date: "2025-11-20",
  time: "19:00",
  notes: "Confirmation: BN-482934",
  reminders: ["1 day before", "2 hours before"]
})

Claude → MCP Server: send_confirmation({
  type: "email",
  template: "restaurant_booking",
  data: { restaurant: "Bella Notte", confirmation: "BN-482934" }
})
```

#### Step 7: Final Confirmation
```
Voice Bot (speaking for Claude):
"All set! Your reservation at Bella Notte is confirmed
for Wednesday, November 20th at 7:00 PM for 4 people.

Confirmation number: BN-482934

I've added it to your calendar with reminders, and you'll
receive a confirmation email shortly. They're aware of the
birthday celebration and your dietary needs.

Would you like me to find parking nearby?"
```

## Demo Client vs Production

### Current `client_demo.go` (Educational)
- **Purpose**: Demonstrate the voice parameter collection phase
- **What it does**:
  - Simulates conversational parameter collection
  - Displays normalized JSON
  - Invokes MCP prompt
  - Shows workflow instructions to help understand what the LLM would receive
- **What it doesn't do**:
  - Doesn't actually execute the workflow
  - Doesn't make intelligent decisions
  - Doesn't call tools based on results

### Production (Claude Desktop + MCP)
- **Purpose**: Actually execute the workflow
- **What it does**:
  - Claude IS the MCP client
  - Claude executes workflow instructions
  - Claude makes intelligent tool calling decisions
  - Claude maintains natural conversation
  - Claude respects approval gates

## Setting Up Claude Desktop as MCP Client

### 1. Configure MCP Server

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

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

### 2. Restart Claude Desktop

The MCP server will be loaded automatically.

### 3. Use the Prompt

In Claude Desktop, you can now say:
```
"Use the book_restaurant prompt to find me an Italian restaurant
for 4 people tomorrow at 7pm downtown"
```

Claude will:
1. Invoke the prompt with those parameters
2. Execute the entire workflow
3. Present options
4. Wait for your approval
5. Make the booking
6. Confirm details

## Key Insights

### 1. The LLM is the Intelligent Layer

The LLM (Claude) provides:
- **Decision making**: Which tools to call, when, and with what parameters
- **Context awareness**: Understanding user preferences and past behavior
- **Error handling**: Adapting when preferred options unavailable
- **Natural conversation**: Speaking with the user throughout
- **Workflow orchestration**: Following complex multi-step processes

### 2. MCP Servers are Specialized

Each MCP server focuses on domain-specific operations:
- **restaurant-server**: Search, availability, reservations
- **calendar-server**: Scheduling and reminders
- **notification-server**: Confirmations
- **review-server**: Ratings and reviews
- **maps-server**: Distances and directions

### 3. Prompts Provide Structure

MCP prompts give Claude:
- **Clear instructions**: Step-by-step workflow
- **Parameter schemas**: What information is needed
- **Safety guidelines**: When to wait for approval
- **Context hints**: Which resources to read

### 4. Voice Provides Natural Interface

The voice interface (OpenAI Realtime API):
- Converts speech to structured parameters
- Enables natural conversation
- Provides parameter validation
- Speaks Claude's responses

## Benefits of This Architecture

### 1. Separation of Concerns
- **Voice**: Parameter collection
- **LLM**: Intelligence and orchestration
- **MCP Servers**: Domain operations

### 2. Maintainability
- Update voice interface without changing servers
- Add new tools without changing prompts
- Modify workflows without changing voice layer

### 3. Scalability
- Add more servers for new domains
- Multiple LLMs can use same servers
- Share servers across applications

### 4. Safety
- Approval gates in prompts
- LLM respects instructions
- Audit trail of tool calls

### 5. Flexibility
- Same servers work with or without voice
- Can use Claude Desktop, API, or other LLMs
- Easy to test individual components

## Comparison with Traditional Approaches

### Traditional Approach
```
Voice → Rules Engine → Database → Response
```
- Fixed workflows
- No context awareness
- Brittle error handling
- Limited adaptability

### Voice + MCP + LLM Approach
```
Voice → LLM (MCP Client) → MCP Servers → Response
```
- Flexible workflows
- Context-aware decisions
- Intelligent error handling
- Highly adaptable

## Conclusion

The LLM as MCP client is the key innovation that makes Voice + MCP integration powerful:

- **Voice interfaces** handle natural language input/output
- **LLMs** provide intelligence and orchestration
- **MCP servers** provide specialized tools and data
- **Prompts** provide structure and safety

Together, they create a system where users can speak naturally, and complex multi-step workflows are executed intelligently with appropriate human approval gates.

This is the future of conversational AI - not rigid command-response systems, but intelligent agents that can orchestrate complex workflows across multiple services while maintaining natural conversation.
