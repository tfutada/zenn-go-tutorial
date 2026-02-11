# MCP Client Example

Demonstrates how to connect to and interact with MCP servers in Go.

## What This Example Shows

1. **Connection Management** — creating a client, connecting via CommandTransport, session lifecycle
2. **Tool Interaction** — listing tools, calling with parameters, handling responses and errors
3. **Resource Access** — listing and reading resources, pretty-printing JSON
4. **Prompt Usage** — listing prompts, getting templates, handling arguments

## Running

```bash
# From repository root (connects to basic server by default)
go run src/mcp_client/main.go
```

This launches the basic MCP server as a subprocess, connects via stdio, runs all example interactions, and displays results.

### Connecting to Different Servers

Modify `main.go`:
```go
// Change from:
serverArgs := []string{"run", "src/mcp_server/basic/main.go"}
// To:
serverArgs := []string{"run", "src/mcp_server/advanced/main.go"}
```

## Code Structure

```go
// 1. Create client
client := mcp.NewClient(&mcp.Implementation{
    Name:    "example-mcp-client",
    Version: "1.0.0",
}, nil)

// 2. Create transport (launches server as subprocess)
transport := &mcp.CommandTransport{
    Command: exec.Command("go", "run", "server/main.go"),
}

// 3. Connect
session, err := client.Connect(ctx, transport, nil)
defer session.Close()

// 4. Call tools
result, err := session.CallTool(ctx, &mcp.CallToolParams{
    Name: "calculator",
    Arguments: map[string]interface{}{
        "operation": "add",
        "a": 10.0,
        "b": 20.0,
    },
})
```

### Key Functions
- `demonstrateTools()` — list and call tools
- `demonstrateResources()` — list and read resources
- `demonstratePrompts()` — list and get prompts
- `printToolResult()` — pretty-print results

## Transport Types

### CommandTransport (used here)
Launches server as subprocess, communicates via stdio. Good for testing. Local only.

### StdioTransport
For connecting to an already running server via stdin/stdout.

### Custom Transport
Implement `Start`, `Send`, `Receive`, `Close` for HTTP, WebSocket, etc.

## Common Patterns

### Sequential Tool Calls
```go
result1, err := session.CallTool(ctx, params1)
params2 := prepareParams(result1)
result2, err := session.CallTool(ctx, params2)
```

### Parallel Tool Calls
```go
var wg sync.WaitGroup
results := make(chan *mcp.CallToolResult, 2)

wg.Add(2)
go func() {
    defer wg.Done()
    result, _ := session.CallTool(ctx, params1)
    results <- result
}()
go func() {
    defer wg.Done()
    result, _ := session.CallTool(ctx, params2)
    results <- result
}()
wg.Wait()
close(results)
```

### Reading All Resources
```go
resourceList, err := session.ListResources(ctx, nil)
for _, resource := range resourceList.Resources {
    contents, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
        URI: resource.URI,
    })
    processContents(contents)
}
```

### Handling Different Content Types
```go
for _, content := range result.Content {
    switch c := content.(type) {
    case mcp.TextContent:
        fmt.Println("Text:", c.Text)
    case mcp.ImageContent:
        fmt.Println("Image:", c.Data)
    case mcp.ResourceContents:
        fmt.Println("Resource:", c.URI)
    }
}
```

### Context and Timeouts
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
session, err := client.Connect(ctx, transport, nil)
```

## Building Your Own Client

1. Create client with `mcp.NewClient`
2. Choose transport (`CommandTransport` for local, custom for remote)
3. Connect with `client.Connect(ctx, transport, nil)`
4. Discover capabilities: `session.ListTools`, `ListResources`, `ListPrompts`
5. Call tools with `session.CallTool`
6. Process results by type-switching on content

## Troubleshooting

| Problem | Solutions |
|---------|-----------|
| `Failed to connect` | Verify server path, check binary exists, ensure stdio transport |
| `unknown tool` | List tools first, check spelling, verify server registration |
| `context deadline exceeded` | Increase timeout, check server responsiveness |
| `failed to unmarshal` | Verify server returns valid JSON, check output format |

## References

- [MCP Client Documentation](https://docs.claude.com/en/docs/agents-and-tools/mcp)
- [Go SDK Reference](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk)
