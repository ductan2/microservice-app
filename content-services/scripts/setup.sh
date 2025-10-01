#!/bin/bash
set -e

echo "🚀 Setting up content-services..."

# Navigate to content-services directory
cd "$(dirname "$0")/.."

echo "📦 Installing dependencies..."
go mod tidy
go mod download

echo "🎨 Generating GraphQL code..."
go run github.com/99designs/gqlgen generate

echo "✅ Setup complete! GraphQL code has been generated."
echo ""
echo "Generated files:"
echo "  - graph/generated/generated.go"
echo "  - graph/model/models_gen.go"
echo ""
echo "To run the server:"
echo "  make run"

