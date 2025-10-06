# English App Infrastructure

Tập trung tất cả các microservices và infrastructure vào một docker-compose duy nhất.

## 🚀 Quick Start

### 1. Khởi động tất cả services
```bash
./start.sh
```

### 2. Dừng tất cả services
```bash
./stop.sh
```

### 3. Xem logs của service cụ thể
```bash
docker-compose logs -f user-services
docker-compose logs -f lesson-services
docker-compose logs -f content-services
docker-compose logs -f golang-init
```

## 📋 Services

### Microservices
- **user-services**: Port 8001
- **lesson-services**: Port 8005  
- **content-services**: Port 8003
- **golang-init**: Port 8004

### Infrastructure
- **PostgreSQL**: Port 5432
- **Redis**: Port 6379
- **RabbitMQ**: Port 5672 (Management: 15672)
- **MongoDB**: Port 27017

### Monitoring
- **Grafana**: Port 3000 (admin/admin)
- **Prometheus**: Port 9090
- **Loki**: Port 3100

## ⚙️ Configuration

### Environment Variables
Tất cả cấu hình được quản lý trong file `docker-compose.env`:

```bash
# Database
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=english_app

# Redis
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_USER=user
RABBITMQ_PASSWORD=password

# Service Ports
USER_SERVICES_PORT=8001
LESSON_SERVICES_PORT=8005
CONTENT_SERVICES_PORT=8003
GOLANG_INIT_PORT=8004
```

### Custom Configuration
Để thay đổi cấu hình, chỉnh sửa file `docker-compose.env` và restart services:

```bash
docker-compose --env-file docker-compose.env down
docker-compose --env-file docker-compose.env up -d
```

## 🔧 Development Commands

### Build specific service
```bash
docker-compose build user-services
```

### Restart specific service
```bash
docker-compose restart user-services
```

### View service status
```bash
docker-compose ps
```

### Access service shell
```bash
docker-compose exec user-services sh
```

## 📊 Monitoring

### Grafana Dashboards
- URL: http://localhost:3000
- Username: admin
- Password: admin

### Prometheus Metrics
- URL: http://localhost:9090

### RabbitMQ Management
- URL: http://localhost:15672
- Username: user
- Password: password

## 🗂️ Project Structure

```
infrastructure/
├── docker-compose.yml          # Main compose file
├── docker-compose.env          # Environment variables
├── start.sh                    # Start script
├── stop.sh                     # Stop script
├── prometheus.yml              # Prometheus config
├── promtail-config.yml         # Promtail config
└── README.md                   # This file
```

## 🐛 Troubleshooting

### Port conflicts
Nếu gặp lỗi port đã được sử dụng, thay đổi port trong `docker-compose.env`:

```bash
USER_SERVICES_PORT=8005  # Thay vì 8001
```

### Database connection issues
Kiểm tra health check của database:

```bash
docker-compose ps
```

### Service không start
Xem logs để debug:

```bash
docker-compose logs user-services
```

### Clean restart
Xóa tất cả containers và volumes:

```bash
docker-compose down -v
docker system prune -f
./start.sh
```
