#!/bin/bash

# Script to regenerate Swagger documentation
set -e

echo "Generating Swagger documentation..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Error: swag is not installed"
    echo "Install it with: go install github.com/swaggo/swag/cmd/swag@latest"
    exit 1
fi

# Generate docs
swag init -g cmd/server/main.go -o docs

echo "✓ Swagger documentation generated successfully!"
echo ""
echo "Files created:"
echo "  - docs/docs.go"
echo "  - docs/swagger.json"
echo "  - docs/swagger.yaml"
echo ""
echo "Start your server and visit: http://localhost:8080/swagger/index.html"
