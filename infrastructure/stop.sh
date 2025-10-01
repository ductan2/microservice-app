#!/bin/bash

# Script để dừng tất cả services
echo "🛑 Stopping English App Microservices..."

# Dừng tất cả services
docker-compose --env-file docker-compose.env down

echo "✅ All services stopped successfully!"
echo ""
echo "💡 To remove volumes as well, run: docker-compose down -v"
