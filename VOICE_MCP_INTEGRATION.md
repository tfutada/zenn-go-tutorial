# Voice + MCP Integration Pattern

This document explains how voice interfaces (like OpenAI Realtime API) integrate with the Model Context Protocol (MCP) to create natural, multi-server orchestrated workflows.

## Overview

The integration creates a three-layer architecture where:
1. **Voice Interface** - Natural language parameter collection
2. **MCP Prompts** - Workflow orchestration across multiple servers
3. **MCP Tools** - Specialized service execution

## Architecture

```
┌─────────────────────────────────────────────┐
│   Voice Interface Layer                     │
│   (OpenAI Realtime API / Voice Bot)         │
│   - Natural conversation                    │
│   - Parameter collection & validation       │
│   - Conversational clarification            │
└──────────────────┬──────────────────────────┘
                   │ Structured parameters
                   ▼
┌─────────────────────────────────────────────┐
│   Prompt Orchestration Layer                │
│   (MCP Prompts)                             │
│   - Workflow coordination                   │
│   - Multi-server orchestration              │
│   - Human approval gates                    │
└──────────────────┬──────────────────────────┘
                   │ Tool calls
                   ▼
┌─────────────────────────────────────────────┐
│   Action Execution Layer                    │
│   (MCP Tools from Multiple Servers)         │
│   - Domain-specific operations              │
│   - State changes                           │
│   - External service calls                  │
└─────────────────────────────────────────────┘
```

## Complete Example: Restaurant Booking

### Phase 1: Voice Parameter Collection

The voice bot conducts a natural conversation to gather required information:

```
User: "I need to book a restaurant"

Bot: "I'd be happy to help! What type of cuisine are you in the mood for?"

User: "Italian sounds good"

Bot: "Great choice! When would you like to dine?"

User: "Next Wednesday evening"

Bot: "Got it, November 20th. What time works for you?"

User: "Around seven"

Bot: "Perfect, 7 PM. How many people will be joining you?"

User: "Four of us"

Bot: "Wonderful! Any dietary restrictions I should know about?"

User: "Yeah, one person is vegetarian and another needs gluten-free"

Bot: "Noted. Is this for a special occasion?"

User: "Yes, it's my friend's birthday"

Bot: "How nice! Where would you like to dine? Any specific area?"

User: "Downtown would be best"
```

### Phase 2: Parameter Extraction & Normalization

The voice bot extracts and normalizes parameters:

```json
{
  "cuisine": "Italian",
  "date": "2025-11-20",
  "time": "19:00",
  "party_size": "4",
  "location": "downtown",
  "dietary_restrictions": "one vegetarian, one gluten-free",
  "occasion": "birthday"
}
```

### Phase 3: MCP Prompt Definition

The MCP server exposes a structured prompt:

```json
{
  "name": "book-restaurant",
  "description": "Find and book a restaurant reservation",
  "arguments": [
    {
      "name": "cuisine",
      "description": "Type of cuisine (e.g., Italian, Japanese, Mexican)",
      "required": true
    },
    {
      "name": "date",
      "description": "Reservation date (YYYY-MM-DD)",
      "required": true
    },
    {
      "name": "time",
      "description": "Preferred time (HH:MM)",
      "required": true
    },
    {
      "name": "party_size",
      "description": "Number of people",
      "required": true
    },
    {
      "name": "location",
      "description": "Area or neighborhood",
      "required": false
    },
    {
      "name": "dietary_restrictions",
      "description": "Any dietary needs (e.g., vegan, gluten-free)",
      "required": false
    },
    {
      "name": "occasion",
      "description": "Special occasion (e.g., birthday, anniversary)",
      "required": false
    }
  ]
}
```

### Phase 4: Prompt Invocation & Workflow

When the voice bot calls `prompts/get`, the MCP server returns workflow instructions:

```json
{
  "description": "Book an Italian restaurant for 4 people on 2025-11-20 at 7:00 PM",
  "messages": [
    {
      "role": "user",
      "content": {
        "type": "text",
        "text": "Please help me book an Italian restaurant for 4 people...\n\nWorkflow:\n\n1. GATHER CONTEXT (Resources):\n   - Read my dining preferences (user-profile-server)\n   - Read my past restaurant reviews (restaurant-server)\n   - Check my calendar (calendar-server) to confirm availability\n\n2. SEARCH & EVALUATE (Tools):\n   - Use search_restaurants (restaurant-server)\n   - Use get_restaurant_reviews (review-server)\n   - Use check_availability (restaurant-server)\n   - Use get_distance_from_location (maps-server)\n\n3. PRESENT OPTIONS:\n   - Show 2-3 best matches with ratings, reviews, availability\n\n4. WAIT FOR SELECTION:\n   - Ask which restaurant to book\n   - DO NOT make any booking yet\n\n5. CONFIRM & BOOK (Only after explicit approval):\n   - Use create_reservation (restaurant-server)\n   - Use add_calendar_event (calendar-server)\n   - Use send_confirmation (notification-server)\n\nImportant: Always wait for confirmation before creating reservations."
      }
    }
  ]
}
```

### Phase 5: Multi-Server Tool Execution

The system coordinates calls across multiple specialized MCP servers:

#### Restaurant Server
```json
{
  "tool": "search_restaurants",
  "arguments": {
    "cuisine": "Italian",
    "location": "downtown",
    "party_size": 4,
    "dietary_filters": ["vegetarian", "gluten-free"]
  }
}
```

#### Review Server
```json
{
  "tool": "get_restaurant_reviews",
  "arguments": {
    "restaurant_ids": ["bella-notte", "trattoria-rosa", "il-giardino"]
  }
}
```

#### Maps Server
```json
{
  "tool": "get_distance_from_location",
  "arguments": {
    "destinations": ["bella-notte", "trattoria-rosa", "il-giardino"],
    "from": "user_home"
  }
}
```

### Phase 6: Voice Presentation of Options

```
Bot: "I found three great options for you:

1. Bella Notte - 4.7 stars, $$$
   Available at 7:00 PM
   Excellent gluten-free pasta and vegetarian options
   They offer complimentary birthday dessert
   12 minutes from downtown square

2. Trattoria Rosa - 4.5 stars, $$
   Available at 6:45 PM or 7:30 PM
   Known for their vegetarian menu
   Birthday decorations available on request
   8 minutes from downtown square

3. Il Giardino - 4.8 stars, $$$$
   Available at 7:15 PM
   Private dining room for 4 available
   Award-winning gluten-free kitchen
   15 minutes from downtown square

Which would you prefer?"

User: "Bella Notte sounds perfect"

Bot: "Excellent choice! Should I go ahead and book Bella Notte
      for 4 people at 7:00 PM on Wednesday, November 20th?
      I'll also add it to your calendar and let them know
      about the birthday celebration and dietary needs."

User: "Yes, please"
```

### Phase 7: Booking Execution

After user approval, execute booking tools:

```json
[
  {
    "server": "restaurant-server",
    "tool": "create_reservation",
    "arguments": {
      "restaurant_id": "bella-notte",
      "date": "2025-11-20",
      "time": "19:00",
      "party_size": 4,
      "special_requests": "Birthday celebration, one vegetarian, one gluten-free"
    }
  },
  {
    "server": "calendar-server",
    "tool": "add_calendar_event",
    "arguments": {
      "title": "Dinner at Bella Notte",
      "date": "2025-11-20",
      "time": "19:00",
      "location": "Bella Notte, Downtown",
      "notes": "Confirmation: BN-482934",
      "reminders": ["1 day before", "2 hours before"]
    }
  },
  {
    "server": "notification-server",
    "tool": "send_confirmation",
    "arguments": {
      "type": "email",
      "template": "restaurant_booking",
      "data": {
        "restaurant": "Bella Notte",
        "confirmation": "BN-482934"
      }
    }
  }
]
```

### Phase 8: Final Confirmation

```
Bot: "All set! Your reservation at Bella Notte is confirmed
      for Wednesday, November 20th at 7:00 PM for 4 people.

      Confirmation number: BN-482934

      I've added it to your calendar with reminders, and
      you'll receive a confirmation email shortly. They're
      aware of the birthday celebration and your dietary needs.

      Would you like me to find parking nearby?"
```

## Multi-Server Architecture

### Servers Involved

1. **User Profile Server**
   - Resources: `user_preferences`, `dining_history`
   - Provides context about user's tastes and habits

2. **Restaurant Server**
   - Tools: `search_restaurants`, `check_availability`, `create_reservation`
   - Resources: `restaurant_database`, `user_past_reservations`

3. **Review Server**
   - Tools: `get_restaurant_reviews`, `get_rating_details`
   - Resources: `aggregated_reviews`, `critic_ratings`

4. **Calendar Server**
   - Tools: `check_availability`, `add_calendar_event`, `set_reminder`
   - Resources: `user_calendar`, `upcoming_events`

5. **Maps Server**
   - Tools: `get_distance_from_location`, `get_directions`, `search_parking`
   - Resources: `user_home_location`, `user_work_location`

6. **Notification Server**
   - Tools: `send_confirmation`, `send_reminder`
   - Resources: `user_contact_preferences`

## Key Benefits

### 1. Natural Interaction
- Users speak conversationally, not fill forms
- Voice bot handles validation and clarification
- Seamless error correction ("Oh, I meant Thursday")

### 2. Guided Parameter Collection
- Voice bot knows what's required from prompt schema
- Can ask follow-up questions intelligently
- Validates in real-time ("Did you mean November 20th?")

### 3. Structured Handoff
- Clean parameter mapping to MCP prompts
- Type conversion and normalization handled
- No ambiguity in the execution phase

### 4. Complex Orchestration
- MCP handles multi-server coordination
- Resources provide context before actions
- Tools execute specialized operations

### 5. Voice Feedback Loop
- Results presented conversationally
- Status updates during long operations
- Natural approval gates ("Should I book this?")

### 6. Human-in-the-Loop
- Explicit approval before irreversible actions
- Options presented before decisions
- User maintains control throughout

## Implementation Pattern

### High-Level Integration Flow

```javascript
// Pseudo-code showing the integration

async function voiceToMCPWorkflow() {
  // 1. Voice conversation phase
  const params = await voiceBot.collectParameters({
    promptSchema: mcpServer.getPrompt("book-restaurant"),
    realtimeAPI: openai.realtime,
    conversationalStyle: "friendly"
  });

  // 2. Invoke MCP prompt with collected parameters
  const workflow = await mcpClient.invokePrompt({
    name: "book-restaurant",
    arguments: params
  });

  // 3. Execute workflow with voice updates
  for (const step of workflow.steps) {
    // Provide status updates via voice
    voiceBot.speak(step.statusMessage);

    // Execute the step (may call multiple servers/tools)
    const result = await step.execute();

    // Check if user approval is required
    if (step.requiresApproval) {
      const approved = await voiceBot.askForConfirmation(
        result.summary
      );

      if (!approved) {
        voiceBot.speak("No problem, canceling the booking.");
        break;
      }
    }
  }

  // 4. Voice confirmation
  voiceBot.speak(workflow.confirmationMessage);
}
```

### Voice Bot Parameter Collection

```javascript
class VoiceParameterCollector {
  async collect(promptSchema) {
    const params = {};

    for (const arg of promptSchema.arguments) {
      if (arg.required || shouldAsk(arg)) {
        const value = await this.askForParameter(arg);
        params[arg.name] = this.normalize(value, arg);
      }
    }

    // Confirm all parameters
    const confirmed = await this.confirmParameters(params);
    return confirmed ? params : this.collect(promptSchema);
  }

  async askForParameter(arg) {
    const question = this.generateQuestion(arg);
    const response = await this.voiceAPI.listen(question);

    // Handle clarifications
    if (!this.isValid(response, arg)) {
      return this.askForClarification(arg, response);
    }

    return response;
  }

  normalize(value, argSchema) {
    // Convert natural language to structured format
    // "next Wednesday" -> "2025-11-20"
    // "around seven" -> "19:00"
    // "four people" -> 4
    return normalizer.convert(value, argSchema.type);
  }
}
```

### MCP Prompt Handler

```go
// Go implementation sketch

type RestaurantPrompt struct {
    Name        string
    Description string
    Arguments   []PromptArgument
}

func (p *RestaurantPrompt) GetPrompt(args map[string]interface{}) (*PromptResult, error) {
    // Build workflow instructions
    workflow := fmt.Sprintf(`
Please help me book a %s restaurant for %v people on %s at %s.

Special requirements: %s
Occasion: %s

Workflow:
1. GATHER CONTEXT (Resources):
   - Read user dining preferences
   - Read past restaurant reviews
   - Check calendar availability

2. SEARCH & EVALUATE (Tools):
   - search_restaurants (restaurant-server)
   - get_restaurant_reviews (review-server)
   - check_availability (restaurant-server)
   - get_distance_from_location (maps-server)

3. PRESENT OPTIONS:
   - Show 2-3 best matches

4. WAIT FOR SELECTION:
   - Ask which restaurant to book
   - DO NOT make any booking yet

5. CONFIRM & BOOK (Only after approval):
   - create_reservation (restaurant-server)
   - add_calendar_event (calendar-server)
   - send_confirmation (notification-server)

Important: Always wait for confirmation before creating reservations.
`, args["cuisine"], args["party_size"], args["date"], args["time"],
       args["dietary_restrictions"], args["occasion"])

    return &PromptResult{
        Description: fmt.Sprintf("Book %s restaurant for %v people",
                                args["cuisine"], args["party_size"]),
        Messages: []Message{
            {
                Role: "user",
                Content: Content{
                    Type: "text",
                    Text: workflow,
                },
            },
        },
    }, nil
}
```

## Additional Use Cases

### Travel Planning

Voice collects: destination, dates, budget, preferences
→ MCP orchestrates: flights, hotels, weather, activities, bookings

### Healthcare Appointment

Voice collects: symptoms, preferred time, doctor specialty
→ MCP orchestrates: availability check, insurance verification, appointment booking, reminders

### Shopping Assistant

Voice collects: product type, budget, delivery preferences
→ MCP orchestrates: product search, price comparison, inventory check, order placement

### Meeting Scheduler

Voice collects: attendees, topic, duration, date preferences
→ MCP orchestrates: calendar availability, room booking, invitation sending, preparation reminders

## Best Practices

### 1. Parameter Validation
- Validate in real-time during voice collection
- Confirm ambiguous inputs ("Did you mean November 20th?")
- Provide defaults when appropriate

### 2. Approval Gates
- Always confirm before irreversible actions
- Summarize what will happen before execution
- Make cancellation easy and clear

### 3. Status Updates
- Provide progress updates during long operations
- Use natural language ("Searching restaurants...")
- Handle timeouts gracefully

### 4. Error Handling
- Provide clear error messages via voice
- Offer alternatives when primary option fails
- Allow easy retry or modification

### 5. Context Preservation
- Remember conversation history
- Use resources to understand user preferences
- Learn from past interactions

## Security Considerations

### 1. Authentication
- Verify user identity before accessing personal data
- Require re-authentication for sensitive operations
- Implement voice biometrics where appropriate

### 2. Authorization
- Validate user permissions for each action
- Implement approval workflows for high-risk operations
- Audit all booking and payment actions

### 3. Data Privacy
- Minimize voice recording retention
- Encrypt sensitive parameters in transit
- Comply with GDPR/privacy regulations

### 4. Rate Limiting
- Prevent abuse through excessive API calls
- Implement per-user quotas
- Monitor for suspicious patterns

## Conclusion

The Voice + MCP integration pattern creates a powerful paradigm where:
- Natural language interfaces handle user interaction
- MCP prompts provide structured orchestration
- Specialized servers execute domain-specific operations

This separation of concerns enables:
- Better user experiences through natural conversation
- Maintainable multi-server architectures
- Reusable workflow templates
- Safe human-in-the-loop approvals

The pattern scales from simple single-server operations to complex multi-server workflows while maintaining a consistent, voice-driven user experience.

## References

- [Model Context Protocol Documentation](https://modelcontextprotocol.io/)
- [MCP Server Concepts](https://modelcontextprotocol.io/docs/learn/server-concepts)
- [OpenAI Realtime API](https://platform.openai.com/docs/guides/realtime)
- [MCP Prompts Specification](https://spec.modelcontextprotocol.io/specification/draft/server/prompts/)
