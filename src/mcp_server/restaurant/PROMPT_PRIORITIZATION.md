# Prompt Prioritization Over Tools

A critical question: **How do we ensure the LLM follows the prompt workflow instead of just randomly calling tools?**

## The Problem

When an MCP server exposes both prompts and tools:

```
Available:
- Prompt: book_restaurant (orchestrated workflow)
- Tools: search_restaurants, check_availability, create_reservation, etc.
```

What prevents the LLM from:
- Immediately calling `search_restaurants` without reading user preferences?
- Calling `create_reservation` before getting user approval?
- Ignoring the workflow entirely and just using tools directly?

## How MCP Systems Handle This

### 1. Prompt Invocation Creates Context

When a user explicitly invokes a prompt, it becomes the **primary instruction**:

```
User: "Use the book_restaurant prompt to find Italian food for 4 people"

System: [Invokes prompts/get with parameters]

LLM receives:
┌─────────────────────────────────────────┐
│ PRIMARY INSTRUCTION (from prompt):      │
│                                         │
│ "Please help me book a restaurant...    │
│                                         │
│ WORKFLOW (follow these steps):         │
│ 1. GATHER CONTEXT - Read resources     │
│ 2. SEARCH & EVALUATE - Call tools      │
│ 3. PRESENT OPTIONS - Show results      │
│ 4. WAIT FOR SELECTION                  │
│ 5. CONFIRM & BOOK - After approval     │
│ ..."                                    │
└─────────────────────────────────────────┘

Tools are available but prompt is the instruction!
```

This is fundamentally different from:
```
User: "Find me Italian restaurants"

LLM thinks: "I'll just call search_restaurants directly"
```

### 2. Prompt Structure Matters

The prompt should use **strong, clear imperatives**:

#### ❌ Weak Prompt (Easy to Ignore)
```
You could search for restaurants. Maybe check some reviews.
Consider asking the user before booking.
```

#### ✅ Strong Prompt (Hard to Ignore)
```
WORKFLOW (follow these steps in order):

1. GATHER CONTEXT (Resources):
   - Read resource://user_preferences FIRST
   - Read resource://dining_history FIRST

2. SEARCH & EVALUATE (Tools):
   - Use search_restaurants tool
   - For each result, use get_restaurant_reviews

3. PRESENT OPTIONS:
   - Show 2-3 best matches

4. WAIT FOR MY SELECTION:
   - DO NOT call create_reservation yet
   - DO NOT make any booking yet

5. CONFIRM & BOOK (ONLY after approval):
   - Wait for me to say "yes", "confirm", "book it"
   - THEN use create_reservation

IMPORTANT RULES:
- Always read resources BEFORE searching
- Present options BEFORE booking
- NEVER call create_reservation without explicit user approval
```

### 3. Tool Descriptions Should Reinforce Workflows

Update tool descriptions to hint at workflow usage:

```go
// ❌ Weak Description
&mcp.Tool{
    Name:        "create_reservation",
    Description: "Create a restaurant reservation",
}

// ✅ Strong Description with Workflow Hint
&mcp.Tool{
    Name:        "create_reservation",
    Description: "Create a restaurant reservation. IMPORTANT: This tool should only be called AFTER presenting options to the user and receiving explicit approval. This is a state-changing action.",
}

// ✅ Another Example
&mcp.Tool{
    Name:        "search_restaurants",
    Description: "Search for restaurants by cuisine and filters. BEST PRACTICE: Read resource://user_preferences first to personalize results.",
}
```

### 4. Modern LLMs Are Good at Following Instructions

Claude Sonnet and other advanced models:
- Understand sequential workflows
- Respect explicit imperatives ("DO NOT call X until Y")
- Follow structured instructions
- Understand context and priorities

When given a clear prompt with workflow steps, Claude will typically follow them.

### 5. Two Modes of Operation

MCP systems naturally have two modes:

#### Mode A: Direct Tool Access (Ad-Hoc)
```
User: "Search for Italian restaurants downtown"

Claude: *Directly calls search_restaurants*
```
- Quick, one-off requests
- No workflow needed
- User knows what they want

#### Mode B: Prompt-Driven (Orchestrated)
```
User: "Use book_restaurant prompt for Italian food tomorrow"

Claude: *Follows entire workflow*
  1. Reads resources
  2. Searches intelligently
  3. Presents options
  4. Waits for approval
  5. Books after confirmation
```
- Complex, multi-step workflows
- Context-aware
- Safety gates

## Improving Our Implementation

Let's enhance the restaurant server to better enforce workflow:

### 1. Update Tool Descriptions

```go
// In main.go - registerRestaurantTools()

searchTool := &mcp.Tool{
    Name:        "search_restaurants",
    Description: "Search for restaurants based on cuisine, location, and dietary needs. WORKFLOW NOTE: Best used as part of the book_restaurant prompt workflow. Consider reading resource://user_preferences first for personalized results.",
}

reserveTool := &mcp.Tool{
    Name:        "create_reservation",
    Description: "Create a restaurant reservation. ⚠️ CRITICAL: This is a state-changing action. Only call this tool AFTER: (1) searching for options, (2) presenting choices to user, (3) receiving explicit user approval. Never call this tool directly without user confirmation.",
}
```

### 2. Strengthen Prompt Instructions

```go
// In main.go - registerPrompts()

workflowText := fmt.Sprintf(`Please help me book a %s restaurant...

⚠️ CRITICAL WORKFLOW - FOLLOW STRICTLY:

STEP 1 - GATHER CONTEXT (DO THIS FIRST):
   YOU MUST read these resources BEFORE any tool calls:
   - Read resource://user_preferences
   - Read resource://dining_history

   DO NOT proceed to step 2 until you have read both resources.

STEP 2 - SEARCH & EVALUATE:
   Now use these tools in sequence:
   - Call search_restaurants with parameters
   - Call get_restaurant_reviews for top 3-5 results
   - Call check_availability for each
   - Call get_distance_from_location

STEP 3 - PRESENT OPTIONS:
   Show me 2-3 best matches with details.
   DO NOT PROCEED TO STEP 4 YET.

STEP 4 - WAIT FOR MY SELECTION:
   Ask which restaurant I prefer.
   ⚠️ DO NOT call create_reservation
   ⚠️ DO NOT call any booking tools
   ⚠️ WAIT for my response

STEP 5 - CONFIRM & BOOK (ONLY AFTER MY APPROVAL):
   After I select a restaurant, confirm details once more.
   Ask: "Should I go ahead and book [restaurant]?"

   ONLY if I say "yes", "confirm", "book it", or similar:
   - Call create_reservation
   - Call add_calendar_event
   - Call send_confirmation

   If I say "no", "cancel", "wait":
   - DO NOT call any booking tools
   - Ask what I'd like to do instead

⚠️ SAFETY RULES (NEVER VIOLATE):
1. NEVER call create_reservation without explicit approval
2. ALWAYS read resources before searching
3. ALWAYS present options before booking
4. ALWAYS wait for user confirmation

If you are uncertain whether I've approved, ASK AGAIN.
It is better to double-check than to make an unwanted reservation.
`, cuisine, partySize, date, time, ...)
```

### 3. Add Tool Metadata for Workflow Hints

While MCP doesn't have built-in workflow enforcement, we can add conventions:

```go
type ToolWorkflowMetadata struct {
    RequiresApproval bool
    RecommendedStep  int
    PrerequisiteTools []string
}

// This could be documented in tool descriptions
searchTool := &mcp.Tool{
    Name: "search_restaurants",
    Description: `Search for restaurants by cuisine and filters.

Workflow Information:
- Recommended Step: 2 (after reading resources)
- Prerequisites: Read resource://user_preferences
- Approval Required: No
- Use Case: Part of book_restaurant workflow`,
}

reserveTool := &mcp.Tool{
    Name: "create_reservation",
    Description: `Create a restaurant reservation.

⚠️ Workflow Information:
- Recommended Step: 5 (final action)
- Prerequisites: search_restaurants, present options, user approval
- Approval Required: YES - EXPLICIT USER CONFIRMATION REQUIRED
- Use Case: ONLY call after user says "yes", "confirm", or "book it"
- Warning: This is a state-changing action that creates a real reservation`,
}
```

## Testing Workflow Adherence

You can test if the LLM follows the workflow:

### Test 1: Does it Read Resources First?
```
User: "Use book_restaurant for Italian tomorrow"

Expected:
✅ Claude reads resource://user_preferences first
✅ Claude reads resource://dining_history
✅ Then Claude calls search_restaurants

Not Expected:
❌ Claude immediately calls search_restaurants
```

### Test 2: Does it Wait for Approval?
```
User: "Use book_restaurant for Italian tomorrow"
Claude: *presents 3 options*
User: "Bella Notte looks good"

Expected:
✅ Claude asks: "Should I book Bella Notte?"
✅ Claude waits for response

Not Expected:
❌ Claude immediately calls create_reservation
```

### Test 3: Does it Respect "No"?
```
User: "Use book_restaurant for Italian tomorrow"
Claude: *presents options*
User: "Actually, I changed my mind"

Expected:
✅ Claude stops workflow
✅ Claude doesn't call create_reservation
✅ Claude asks what user wants instead

Not Expected:
❌ Claude proceeds with booking anyway
```

## When Tools Are Called Directly (Bypass Workflow)

There are legitimate reasons to call tools directly:

### Scenario 1: Quick Lookup
```
User: "What Italian restaurants are downtown?"

Claude: *Calls search_restaurants directly*

This is fine - user just wants information, not full workflow
```

### Scenario 2: Check Single Thing
```
User: "Is Bella Notte available tomorrow at 7?"

Claude: *Calls check_availability directly*

This is fine - simple yes/no question
```

### Scenario 3: Tool Invocation is Better
For these cases, direct tool use is actually preferable to the full workflow!

## Client-Side Enforcement (Advanced)

In some architectures, the client application can enforce workflows:

```typescript
// Pseudo-code for advanced enforcement
class WorkflowEnforcedMCPClient {
  async invokePrompt(name, params) {
    const workflow = await this.getPromptWorkflow(name);

    // Only expose tools relevant to current step
    this.restrictToolsTo(workflow.currentStepTools);

    // Execute prompt
    const result = await this.mcpSession.getPrompt(name, params);

    // Monitor tool calls
    this.onToolCall((toolName) => {
      if (!workflow.allowedTools.includes(toolName)) {
        throw new Error(`Tool ${toolName} not allowed in current workflow step`);
      }
    });
  }
}
```

However, **this is not standard MCP** - it's custom enforcement logic.

## Best Practices Summary

### 1. Write Strong Prompts
- Use imperative language ("MUST", "DO NOT", "NEVER")
- Number steps clearly
- Repeat critical rules
- Use visual markers (⚠️, ❌, ✅)

### 2. Descriptive Tool Metadata
- Add workflow hints to tool descriptions
- Mark state-changing tools clearly
- Mention prerequisites
- Note approval requirements

### 3. Trust Modern LLMs
- Claude Sonnet is very good at following instructions
- Clear prompts are usually sufficient
- LLMs understand context and safety

### 4. Test Thoroughly
- Test that workflow is followed
- Test that approval gates work
- Test error cases and "no" responses

### 5. User Control
- Users choose: direct tool use vs prompt workflow
- Both modes are valuable
- Prompts provide structure; tools provide flexibility

## Conclusion

**Prompts don't technically "override" tools** - instead:

1. **Explicit prompt invocation** makes the prompt the primary instruction
2. **Strong workflow language** guides the LLM's behavior
3. **Tool descriptions** reinforce proper usage
4. **Modern LLMs** are good at following structured instructions
5. **Both modes** (direct tools vs prompted workflows) have value

The key is writing clear, imperative prompts with explicit sequencing and safety rules. When users invoke prompts explicitly, LLMs like Claude treat them as the primary instruction and follow the workflow.

**The architecture trusts the LLM to be intelligent** - which modern models like Claude Sonnet are!
