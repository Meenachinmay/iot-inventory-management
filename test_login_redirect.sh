#!/bin/bash

echo "Testing login redirect functionality..."

# Test the login endpoint with a valid UUID
echo "Sending login request with valid UUID..."
response=$(curl -s -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=00000000-0000-0000-0000-000000000001" \
  -i http://localhost:8080/ui/login)

# Print the full response for debugging
echo "Full response:"
echo "$response"
echo "------------------------"

# Extract and display the exact header
echo "Exact header found:"
echo "$response" | grep -i "HX-Redirect" || echo "No HX-Redirect header found"
echo "------------------------"

# Check if the response contains HX-Redirect header (case insensitive)
if echo "$response" | grep -i -q "HX-Redirect"; then
  echo "✅ HX-Redirect header found in response"
else
  echo "❌ HX-Redirect header not found in response"
fi

# Check if the response contains the dashboard URL (case insensitive)
if echo "$response" | grep -i -q "HX-Redirect: /ui/dashboard"; then
  echo "✅ Redirect to dashboard confirmed"
else
  echo "❌ Redirect to dashboard not found"
fi

echo "Test completed."