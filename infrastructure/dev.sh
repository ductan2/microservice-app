#!/bin/bash

# Development script with auto reload
# This script starts all services with auto reload enabled

echo "🚀 Starting microservices with auto reload..."

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose not found. Please install docker-compose first."
    exit 1
fi

# Start services with override file for development
docker-compose -f docker-compose.yml -f docker-compose.override.yml up --build

echo "✅ All services started with auto reload enabled!"
echo ""
echo "📋 Service URLs:"
echo "  - User Services: http://localhost:8001"
echo "  - Lesson Services: http://localhost:8005"
echo "  - Content Services: http://localhost:8004"
echo "  - Notification Services: http://localhost:8003"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo "  - RabbitMQ Management: http://localhost:15672"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000"
echo ""
echo "🔄 Auto reload is enabled for all services!"
echo "   - Go services use Air"
echo "   - Python services use uvicorn --reload"
echo "   - Node.js services use nodemon"
