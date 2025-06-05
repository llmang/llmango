#!/bin/bash

# Test script for the agent endpoint
# This demonstrates the complete end-to-end workflow

echo "🚀 Testing Agent System Integration"
echo "=================================="

# Check if the server is running
if ! curl -s http://localhost:8080 > /dev/null; then
    echo "❌ Server is not running on localhost:8080"
    echo "Please start the server first with: go run main.go"
    exit 1
fi

echo "✅ Server is running"

# Test the agent endpoint
echo ""
echo "📤 Sending request to /agents endpoint..."

response=$(curl -s -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello! Can you help me analyze the sentiment of this text: I absolutely love this new AI system!",
    "workflowUID": "example_workflow"
  }')

echo "📥 Response received:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

# Check if the response contains expected fields
if echo "$response" | grep -q '"status"'; then
    echo ""
    echo "✅ Agent endpoint is working correctly!"
    echo "🎉 Complete end-to-end workflow test passed!"
else
    echo ""
    echo "❌ Unexpected response format"
    exit 1
fi

echo ""
echo "🔗 You can also test manually with:"
echo "curl -X POST http://localhost:8080/agents \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"input\": \"Your message here\", \"workflowUID\": \"example_workflow\"}'"