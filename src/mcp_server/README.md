# Model Context Protocol (MCP) Servers in Go

This directory contains comprehensive examples of building MCP (Model Context Protocol) servers in Go.

## What is MCP?

The **Model Context Protocol (MCP)** is an open standard introduced by Anthropic that enables AI applications to connect to external data sources and tools. Think of it as a "USB-C port for AI" - a standardized way for LLMs like Claude to interact with your services and data.

### Key Features

- **Tools**: Functions that the LLM can call to perform actions
- **Resources**: Data sources that can be read
- **Prompts**: Reusable prompt templates
- **Multiple Transports**: stdio, HTTP/SSE, custom transports

## Examples Overview

### 1. Basic MCP Server (`basic/`)

A simple MCP server demonstrating core concepts:
- Multiple tools (calculator, echo, timestamp, weather)
- Stdio transport (simplest option for local tools)
- Input validation and error handling
- Clear code structure with comments

**Use Case**: Learning MCP basics, building simple command-line tools

### 2. Advanced MCP Server (`advanced/`)

A sophisticated MCP server showcasing advanced features:
- Multiple Tools (search, filter, statistics)
- Multiple Resources (users, products, system info)
- Multiple Prompts (data analysis, report generation)
- Data loading from JSON files
- Complex query and filtering capabilities

**Use Case**: Production-ready servers, data integration, complex workflows

## Quick Start

### Prerequisites

```bash
# Install the official MCP Go SDK
go get github.com/modelcontextprotocol/go-sdk
```

### Running the Basic Server

```bash
# From repository root
go run src/mcp_server/basic/main.go
```

### Running the Advanced Server

```bash
# From repository root
go run src/mcp_server/advanced/main.go
```

### Testing with the Client

```bash
# In one terminal, start a server
go run src/mcp_server/basic/main.go

# In another terminal, run the client
go run src/mcp_client/main.go
```

## Architecture

### MCP Components

```
┌─────────────────┐
│   MCP Client    │  (e.g., Claude, custom client)
└────────┬────────┘
         │ JSON-RPC over transport
         │
┌────────▼────────┐
│   MCP Server    │  (Your Go application)
├─────────────────┤
│ • Tools         │  Callable functions
│ • Resources     │  Readable data sources
│ • Prompts       │  Template generators
└─────────────────┘
```

### Transport Options

1. **Stdio** (Recommended for local tools)
   - Communication via stdin/stdout
   - Perfect for CLI tools and desktop integrations
   - Used in these examples

2. **HTTP/SSE** (For web services)
   - Server-Sent Events for streaming
   - Better for remote access and web integrations

3. **Custom**
   - Implement your own transport using jsonrpc package

## Integration with Claude

### Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "basic-server": {
      "command": "go",
      "args": ["run", "/path/to/src/mcp_server/basic/main.go"]
    },
    "advanced-server": {
      "command": "go",
      "args": ["run", "/path/to/src/mcp_server/advanced/main.go"]
    }
  }
}
```

### Claude Code

Claude Code automatically discovers and uses MCP servers configured in your environment.

## Development Guide

### Creating a New Tool

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

        // Your logic here
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
        data := map[string]interface{}{
            "key": "value",
        }

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
            {
                Name:        "topic",
                Description: "The topic to generate prompt for",
                Required:    true,
            },
        },
    }

    handler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
        topic := request.Params.Arguments["topic"].(string)

        promptText := fmt.Sprintf("Please analyze the following topic: %s", topic)

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

    mcp.AddPrompt(server, prompt, handler)
}
```

## Best Practices

### Tool Design
- Keep tools focused and single-purpose
- Provide clear, descriptive names and descriptions
- Use strong typing for arguments
- Implement comprehensive error handling
- Validate all inputs

### Performance
- Leverage Go's concurrency (goroutines) for parallel operations
- Implement timeouts for long-running operations
- Use efficient JSON parsing
- Cache frequently accessed data

### Security
- Validate all inputs thoroughly
- Never expose sensitive credentials in responses
- Use appropriate authentication mechanisms
- Sanitize file paths and user inputs
- Implement rate limiting for public-facing servers

### Error Handling
- Return meaningful error messages
- Use proper error wrapping with `fmt.Errorf`
- Log errors for debugging
- Don't leak internal details to clients

## Common Use Cases

### Local Tools
- File system operations
- Database queries
- API integrations
- Data transformation
- System utilities

### Enterprise Integration
- Company knowledge bases
- Internal API access
- Compliance checking
- Code analysis tools
- Custom workflows

### Data Services
- Real-time data feeds
- Analytics and reporting
- Data aggregation
- Search and filtering

## Troubleshooting

### Server Won't Start
- Check that the Go SDK is installed: `go get github.com/modelcontextprotocol/go-sdk`
- Verify your Go version is 1.21 or higher
- Check for port conflicts if using HTTP transport

### Tools Not Working
- Verify tool registration in the server code
- Check that arguments match the InputSchema
- Review server logs for error messages
- Test with the example client first

### Connection Issues
- Ensure the server is running when the client connects
- Check that the transport type matches (stdio, HTTP, etc.)
- Verify file paths in configuration are absolute

## Additional Resources

- [Official MCP Documentation](https://docs.claude.com/en/docs/agents-and-tools/mcp)
- [MCP Go SDK GitHub](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://github.com/modelcontextprotocol)
- [Example MCP Servers](https://github.com/modelcontextprotocol)

## Contributing

Feel free to extend these examples with:
- Additional tools for different use cases
- Alternative transport implementations
- Integration with popular Go libraries
- Performance optimizations
- More comprehensive error handling

## License

These examples are provided as educational material for learning MCP server development in Go.
