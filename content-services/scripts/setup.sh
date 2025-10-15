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



go run import_data.go \
  --user-id "2254a7dd-bfa3-4217-be59-e6964db03a26" \
  --file data.json \
  --db-host postgres \
  --db-port 5432 \
  --db-user user \
  --db-password password \
  --db-name english_app