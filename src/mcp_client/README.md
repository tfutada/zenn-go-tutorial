# MCP Client Example

This directory contains an example MCP (Model Context Protocol) client written in Go that demonstrates how to connect to and interact with MCP servers.

## What This Example Shows

The client demonstrates:

1. **Connection Management**
   - Creating an MCP client
   - Connecting to a server using CommandTransport (launches server process)
   - Managing session lifecycle

2. **Tool Interaction**
   - Listing available tools
   - Calling tools with parameters
   - Handling tool responses
   - Error handling for failed tool calls

3. **Resource Access**
   - Listing available resources
   - Reading resource contents
   - Pretty-printing JSON data

4. **Prompt Usage**
   - Listing available prompts
   - Getting prompt templates
   - Handling prompt arguments

## Running the Example

### Prerequisites

Make sure you have the MCP Go SDK installed:

```bash
go get github.com/modelcontextprotocol/go-sdk
```

### Basic Usage

The client is configured to connect to the basic MCP server by default:

```bash
# From repository root
go run src/mcp_client/main.go
```

This will:
1. Launch the basic MCP server as a subprocess
2. Connect to it via stdio transport
3. Run through all the example interactions
4. Display the results

### Connecting to Different Servers

To connect to the advanced server instead, modify the `main.go` file:

```go
// Change from:
serverArgs := []string{"run", "src/mcp_server/basic/main.go"}

// To:
serverArgs := []string{"run", "src/mcp_server/advanced/main.go"}
```

## Example Output

```
=== MCP Client Example ===

Connecting to MCP server: go [run src/mcp_server/basic/main.go]

✓ Connected to MCP server successfully!

--- Tools Demonstration ---

Available tools: 4
1. calculator - Performs basic arithmetic operations (add, subtract, multiply, divide, power, sqrt)
2. echo - Echoes back text with optional transformations (uppercase, lowercase, reverse, word_count)
3. timestamp - Returns current timestamp in various formats
4. get_weather - Gets mock weather data for a given city (demonstration only)

Example 1: Calling 'calculator' tool
  Calculator (15.5 + 24.3):
    Result: 39.8000

Example 2: Calling 'echo' tool
  Echo (uppercase):
    HELLO FROM MCP CLIENT!

...
```

## Code Structure

### Main Components

```go
// 1. Create client
client := mcp.NewClient(&mcp.Implementation{
    Name:    "example-mcp-client",
    Version: "1.0.0",
}, nil)

// 2. Create transport
transport := &mcp.CommandTransport{
    Command: exec.Command("go", "run", "server/main.go"),
}

// 3. Connect
session, err := client.Connect(ctx, transport, nil)
defer session.Close()

// 4. Interact with server
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

- `demonstrateTools()` - Shows how to list and call tools
- `demonstrateResources()` - Shows how to list and read resources
- `demonstratePrompts()` - Shows how to list and get prompts
- `printToolResult()` - Helper for pretty-printing results

## Transport Types

### CommandTransport (Used in This Example)

Launches the server as a subprocess and communicates via stdio:

```go
transport := &mcp.CommandTransport{
    Command: exec.Command("go", "run", "server/main.go"),
}
```

**Pros:**
- Automatic server lifecycle management
- Perfect for testing
- Works with any stdio-based server

**Cons:**
- Only for local servers
- Creates a new process each time

### StdioTransport

For connecting to an already running server via stdin/stdout:

```go
transport := &mcp.StdioTransport{}
```

### Custom Transport

You can implement your own transport for HTTP, WebSocket, etc.:

```go
type MyTransport struct {
    // Your fields
}

func (t *MyTransport) Start(ctx context.Context) error {
    // Connect to server
}

func (t *MyTransport) Send(msg interface{}) error {
    // Send message
}

func (t *MyTransport) Receive() (interface{}, error) {
    // Receive message
}

func (t *MyTransport) Close() error {
    // Cleanup
}
```

## Advanced Usage

### Error Handling

Always check for errors and handle them appropriately:

```go
result, err := session.CallTool(ctx, params)
if err != nil {
    log.Printf("Tool call failed: %v", err)
    return err
}
```

### Context and Timeouts

Use contexts to control execution time:

```go
// Create context with 30-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// All operations will respect this timeout
session, err := client.Connect(ctx, transport, nil)
```

### Handling Different Content Types

Tool responses can contain different content types:

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

### Listing Capabilities

Check what the server supports before calling:

```go
// List tools
tools, err := session.ListTools(ctx, nil)

// List resources (might not be supported)
resources, err := session.ListResources(ctx, nil)

// List prompts (might not be supported)
prompts, err := session.ListPrompts(ctx, nil)
```

## Building Your Own Client

### Step 1: Create Client

```go
client := mcp.NewClient(&mcp.Implementation{
    Name:    "my-client",
    Version: "1.0.0",
}, nil)
```

### Step 2: Choose Transport

For local servers:
```go
transport := &mcp.CommandTransport{
    Command: exec.Command("path/to/server"),
}
```

### Step 3: Connect

```go
ctx := context.Background()
session, err := client.Connect(ctx, transport, nil)
if err != nil {
    log.Fatal(err)
}
defer session.Close()
```

### Step 4: Call Tools

```go
result, err := session.CallTool(ctx, &mcp.CallToolParams{
    Name: "tool_name",
    Arguments: map[string]interface{}{
        "param1": "value1",
        "param2": 123,
    },
})
```

### Step 5: Process Results

```go
for _, content := range result.Content {
    if textContent, ok := content.(mcp.TextContent); ok {
        fmt.Println(textContent.Text)
    }
}
```

## Common Patterns

### Sequential Tool Calls

```go
// Call tool 1
result1, err := session.CallTool(ctx, params1)
if err != nil {
    return err
}

// Use result1 to inform tool 2
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
if err != nil {
    return err
}

for _, resource := range resourceList.Resources {
    contents, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
        URI: resource.URI,
    })
    if err != nil {
        log.Printf("Failed to read %s: %v", resource.URI, err)
        continue
    }

    // Process contents
    processContents(contents)
}
```

## Testing

### Unit Tests

Test your client code with a mock server:

```go
func TestMyClient(t *testing.T) {
    // Create mock server
    server := createMockServer()
    defer server.Close()

    // Create client and connect
    client := createClient()
    session, err := client.Connect(ctx, transport, nil)
    require.NoError(t, err)

    // Test tool calls
    result, err := session.CallTool(ctx, params)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests

Test against real servers:

```bash
# Start server
go run src/mcp_server/basic/main.go &
SERVER_PID=$!

# Run client tests
go test ./src/mcp_client/...

# Cleanup
kill $SERVER_PID
```

## Troubleshooting

### Connection Failures

**Problem**: `Failed to connect to server`

**Solutions**:
- Verify the server path is correct
- Check that the server binary exists and is executable
- Ensure the server supports stdio transport
- Check server logs for startup errors

### Tool Call Failures

**Problem**: `Tool call failed: unknown tool`

**Solutions**:
- List available tools first with `session.ListTools()`
- Check tool name spelling
- Verify the server has registered the tool

### Timeout Errors

**Problem**: `context deadline exceeded`

**Solutions**:
- Increase the context timeout
- Check if the server is responding slowly
- Verify network connectivity (for remote servers)
- Check if the tool operation is actually stuck

### JSON Parsing Errors

**Problem**: `failed to unmarshal response`

**Solutions**:
- Verify the server is returning valid JSON
- Check the tool's output format
- Use the MCP SDK's built-in parsers

## Next Steps

1. **Modify the Client**: Edit `main.go` to call different tools or servers
2. **Create Your Own Tools**: Add custom tool interactions
3. **Build an Application**: Use this client as a foundation for your app
4. **Add Error Recovery**: Implement retry logic and fallbacks
5. **Add Logging**: Integrate structured logging for debugging

## Additional Resources

- [MCP Client Documentation](https://docs.claude.com/en/docs/agents-and-tools/mcp)
- [Go SDK Reference](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk)
- [MCP Examples Repository](https://github.com/modelcontextprotocol)
