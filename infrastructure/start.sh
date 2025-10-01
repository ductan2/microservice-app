#!/bin/bash

# Script để khởi động tất cả services
echo "🚀 Starting English App Microservices..."

# Kiểm tra xem file .env có tồn tại không
if [ ! -f "docker-compose.env" ]; then
    echo "⚠️  File docker-compose.env not found. Creating from template..."
    cp docker-compose.env docker-compose.env.backup 2>/dev/null || true
fi

# Khởi động tất cả services
echo "📦 Building and starting all services..."
docker-compose --env-file docker-compose.env up --build -d

echo "✅ All services started successfully!"
echo ""
echo "🌐 Service URLs:"
echo "  - User Services: http://localhost:8001"
echo "  - Lesson Services: http://localhost:8002"
echo "  - Content Services: http://localhost:8003"
echo "  - Golang Init: http://localhost:8004"
echo ""
echo "📊 Monitoring URLs:"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9090"
echo "  - RabbitMQ Management: http://localhost:15672 (user/password)"
echo ""
echo "🔍 To view logs: docker-compose logs -f [service-name]"
echo "🛑 To stop all: docker-compose down"
