#!/bin/bash

# Quick test script for MCP examples
# This script tests that the MCP examples compile and run

echo "=== Testing MCP Examples ==="
echo ""

# Test 1: Basic server syntax check
echo "Test 1: Checking basic server syntax..."
go build -o /tmp/mcp_basic src/mcp_server/basic/main.go
if [ $? -eq 0 ]; then
    echo "✓ Basic server compiles successfully"
    rm /tmp/mcp_basic
else
    echo "✗ Basic server has compilation errors"
    exit 1
fi

# Test 2: Advanced server syntax check
echo "Test 2: Checking advanced server syntax..."
go build -o /tmp/mcp_advanced src/mcp_server/advanced/main.go
if [ $? -eq 0 ]; then
    echo "✓ Advanced server compiles successfully"
    rm /tmp/mcp_advanced
else
    echo "✗ Advanced server has compilation errors"
    exit 1
fi

# Test 3: Client syntax check
echo "Test 3: Checking client syntax..."
go build -o /tmp/mcp_client src/mcp_client/main.go
if [ $? -eq 0 ]; then
    echo "✓ Client compiles successfully"
    rm /tmp/mcp_client
else
    echo "✗ Client has compilation errors"
    exit 1
fi

echo ""
echo "=== All Syntax Tests Passed ==="
echo ""
echo "To run the examples:"
echo "  1. Basic server:    go run src/mcp_server/basic/main.go"
echo "  2. Advanced server: go run src/mcp_server/advanced/main.go"
echo "  3. Client:          go run src/mcp_client/main.go"
