# MCP Servers in Go

## What is MCP?

The **Model Context Protocol (MCP)** is an open standard by Anthropic that enables AI applications to connect to external data sources and tools — a standardized way for LLMs like Claude to interact with your services.

**Core capabilities:**
- **Tools**: Functions the LLM can call to perform actions
- **Resources**: Readable data sources
- **Prompts**: Reusable workflow templates
- **Transports**: stdio (local), HTTP/SSE (web), custom

## Examples

### Basic Server (`basic/`)
Simple MCP server: calculator, echo, timestamp, weather tools over stdio transport.

### Advanced Server (`advanced/`)
Production-style server: search/filter/statistics tools, resources (users, products, system info), prompts (data analysis, report generation), JSON data loading.

### Restaurant Server (`restaurant/`)
Voice + MCP integration demo: 8 tools across 5 simulated domain servers, resources, orchestrated booking workflow. See `restaurant/CLAUDE.md` for details.

## Quick Start

```bash
# Install SDK
go get github.com/modelcontextprotocol/go-sdk

# Run servers
go run src/mcp_server/basic/main.go
go run src/mcp_server/advanced/main.go
go run src/mcp_server/restaurant/main.go

# Test with client (in separate terminal)
go run src/mcp_client/main.go
```

## Architecture

```
┌─────────────────┐
│   MCP Client    │  (Claude, custom client)
└────────┬────────┘
         │ JSON-RPC over transport
┌────────▼────────┐
│   MCP Server    │  (Your Go application)
├─────────────────┤
│ • Tools         │  Callable functions
│ • Resources     │  Readable data sources
│ • Prompts       │  Template generators
└─────────────────┘
```

### Transport Options
- **Stdio** (recommended for local tools): communication via stdin/stdout, used in these examples
- **HTTP/SSE**: Server-Sent Events for remote/web integrations
- **Custom**: implement your own transport using the jsonrpc package

## Claude Desktop Integration

Config path: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

```json
{
  "mcpServers": {
    "basic-server": {
      "command": "go",
      "args": ["run", "/absolute/path/to/src/mcp_server/basic/main.go"]
    },
    "advanced-server": {
      "command": "go",
      "args": ["run", "/absolute/path/to/src/mcp_server/advanced/main.go"]
    },
    "restaurant-booking": {
      "command": "go",
      "args": ["run", "/absolute/path/to/src/mcp_server/restaurant/main.go"]
    }
  }
}
```

Restart Claude Desktop after config changes. Verify via logs:
```bash
tail -f ~/Library/Logs/Claude/mcp.log
tail -f ~/Library/Logs/Claude/mcp-server-go-tutorial-restaurant.log
```

### Invoking Prompts in Claude Desktop

Prompts can be invoked explicitly:
```
Execute the book_restaurant prompt from go-tutorial-restaurant with:
- cuisine: Italian, date: 2025-11-20, time: 19:00, party_size: 4
Follow the complete workflow instructions it returns.
```

Or naturally: "Use the book_restaurant prompt to find Italian food for 4 people tomorrow."

## Development Guide

### Creating a Tool

```go
func registerMyTool(server *mcp.Server) {
    tool := &mcp.Tool{
        Name:        "my_tool",
        Description: "What this tool does",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "param1": map[string]interface{}{
                    "type":        "string",
                    "description": "Parameter description",
                },
            },
            Required: []string{"param1"},
        },
    }

    handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        param1 := request.Params.Arguments["param1"].(string)
        result := fmt.Sprintf("Processed: %s", param1)
        return mcp.NewToolResultText(result), nil
    }

    mcp.AddTool(server, tool, handler)
}
```

### Creating a Resource

```go
func registerMyResource(server *mcp.Server) {
    resource := &mcp.Resource{
        URI:         "resource://my_data",
        Name:        "My Data",
        Description: "Description of this data source",
        MimeType:    "application/json",
    }

    handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]interface{}, error) {
        data := map[string]interface{}{"key": "value"}
        jsonData, _ := json.MarshalIndent(data, "", "  ")
        return []interface{}{
            mcp.TextResourceContents{
                ResourceContents: mcp.ResourceContents{
                    URI:      "resource://my_data",
                    MimeType: "application/json",
                },
                Text: string(jsonData),
            },
        }, nil
    }

    mcp.AddResource(server, resource, handler)
}
```

### Creating a Prompt

```go
func registerMyPrompt(server *mcp.Server) {
    prompt := &mcp.Prompt{
        Name:        "my_prompt",
        Description: "What this prompt does",
        Arguments: []mcp.PromptArgument{
            {Name: "topic", Description: "The topic", Required: true},
        },
    }

    handler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
        topic := request.Params.Arguments["topic"].(string)
        return &mcp.GetPromptResult{
            Messages: []mcp.PromptMessage{
                {
                    Role:    "user",
                    Content: mcp.TextContent{Type: "text", Text: fmt.Sprintf("Analyze: %s", topic)},
                },
            },
        }, nil
    }

    mcp.AddPrompt(server, prompt, handler)
}
```

## Voice + MCP Integration Pattern

A three-layer architecture for voice-driven workflows:

```
┌─────────────────────────────────────┐
│  Voice Interface (Realtime API)     │  Natural conversation,
│  Parameter collection & validation  │  structured output
└──────────────────┬──────────────────┘
                   │ JSON parameters
┌──────────────────▼──────────────────┐
│  LLM (Claude) = MCP CLIENT         │  Receives prompt workflow,
│  Reads resources, calls tools,      │  makes decisions,
│  orchestrates multi-step flow       │  respects approval gates
└──────────────────┬──────────────────┘
                   │ tool calls
┌──────────────────▼──────────────────┐
│  MCP Server(s)                      │  Domain-specific tools
│  restaurant / calendar / maps / ... │  and resources
└─────────────────────────────────────┘
```

**Key insight**: The LLM (Claude) IS the MCP client. It receives workflow instructions from prompts and intelligently decides which tools to call, in what order, while maintaining natural conversation and respecting approval gates.

### How It Works

1. **Voice collects parameters** naturally ("Italian restaurant, 4 people, tomorrow at 7pm")
2. **Parameters normalized** to structured JSON
3. **MCP prompt invoked** with parameters — returns step-by-step workflow instructions
4. **LLM executes workflow**: reads resources for context → searches → presents options → waits for approval → books
5. **Voice presents results** conversationally

### Use Cases Beyond Restaurants

- **Travel**: destination, dates, budget → flights, hotels, activities orchestration
- **Healthcare**: symptoms, time → availability, insurance, appointment booking
- **Shopping**: product type, budget → search, price comparison, order placement
- **Meetings**: attendees, topic → calendar availability, room booking, invitations

## Prompt-Driven Workflows vs Direct Tool Access

MCP systems naturally support two modes:

### Direct Tool Access (Ad-Hoc)
```
User: "Search for Italian restaurants downtown"
→ LLM calls search_restaurants directly
```
Quick one-off requests. No workflow needed.

### Prompt-Driven (Orchestrated)
```
User: "Use book_restaurant prompt for Italian food tomorrow"
→ LLM follows entire workflow: resources → search → present → wait → book
```
Complex multi-step workflows with safety gates.

### Making Prompts Effective

**Use strong imperatives** in prompt text:
```
STEP 1 - GATHER CONTEXT (DO THIS FIRST):
  Read resource://user_preferences BEFORE any tool calls.

STEP 4 - WAIT FOR SELECTION:
  ⚠️ DO NOT call create_reservation
  ⚠️ WAIT for user response

SAFETY RULES:
- NEVER call create_reservation without explicit approval
- ALWAYS present options before booking
```

**Add workflow hints to tool descriptions:**
```go
&mcp.Tool{
    Name:        "create_reservation",
    Description: "Create a restaurant reservation. ⚠️ Only call AFTER presenting options and receiving explicit user approval.",
}
```

## Best Practices

### Tool Design
- Keep tools focused and single-purpose
- Provide clear names and descriptions with workflow hints
- Validate all inputs; implement comprehensive error handling
- Mark state-changing tools clearly in descriptions

### Performance
- Use goroutines for parallel operations; implement timeouts
- Cache frequently accessed data; use efficient JSON parsing

### Security
- Validate all inputs; never expose credentials in responses
- Sanitize file paths; implement rate limiting for public servers

## Troubleshooting

- **Server won't start**: verify SDK installed (`go get github.com/modelcontextprotocol/go-sdk`), Go 1.21+, no port conflicts
- **Tools not working**: verify registration, check argument schema match, test with example client first
- **Connection issues**: ensure server is running, transport type matches, file paths are absolute
- **Prompts not invoked**: try explicit phrasing ("Execute the X prompt from Y server"), fall back to individual tool calls

## References

- [MCP Documentation](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://github.com/modelcontextprotocol)
- [OpenAI Realtime API](https://platform.openai.com/docs/guides/realtime)
