# Networking & Security Configuration

## Overview

This document describes the networking and security improvements implemented for the microservices architecture.

## Network Architecture

### Network Isolation

The system uses two separate Docker networks:

1. **Frontend Network** (`english-app-frontend`)
   - Public-facing services
   - Traefik reverse proxy
   - Monitoring dashboards (Grafana)
   - BFF services

2. **Backend Network** (`english-app-backend`)
   - Internal services only
   - Databases (PostgreSQL, MongoDB, Redis)
   - Message queues (RabbitMQ)
   - Microservices
   - Monitoring exporters

### Network Configuration

```yaml
networks:
  frontend:
    driver: bridge
    name: english-app-frontend
  backend:
    driver: bridge
    name: english-app-backend
    internal: true  # Isolate backend network
```

## Security Hardening

### 1. Password Security

All services now use strong passwords:
- PostgreSQL: `POSTGRES_PASSWORD`
- Redis: `REDIS_PASSWORD` (with AUTH enabled)
- RabbitMQ: `RABBITMQ_PASSWORD`
- Grafana: `GRAFANA_ADMIN_PASSWORD`

### 2. Port Exposure

**Development Mode:**
- All ports exposed for debugging
- Database ports accessible externally

**Production Mode:**
- Only Traefik (80, 443) and Grafana (3000) exposed
- All database ports removed
- All microservice ports removed
- All exporter ports removed

### 3. HTTPS Configuration

Traefik configured with:
- Let's Encrypt SSL certificates
- Automatic HTTP to HTTPS redirect
- TLS 1.2+ only

```yaml
traefik:
  command:
    - "--entrypoints.websecure.address=:443"
    - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
    - "--entrypoints.web.http.redirections.entrypoint.to=websecure"
```

## Persistence & Data Protection

### Persistent Volumes

All critical data is persisted:

```yaml
volumes:
  postgres_data:      # Database data
  redis_data:         # Cache data
  rabbitmq_data:      # Queue data
  prometheus_data:    # Metrics data
  loki_data:         # Log data
  grafana_data:      # Dashboard data
  traefik_letsencrypt: # SSL certificates
```

### Data Backup Strategy

1. **Database Backups**
   ```bash
   # PostgreSQL backup
   docker exec postgres pg_dump -U user english_app > backup.sql
   
   # MongoDB backup
   docker exec mongodb mongodump --out /backup
   ```

2. **Volume Backups**
   ```bash
   # Backup all volumes
   docker run --rm -v postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_data.tar.gz -C /data .
   ```

## Deployment Configurations

### Development

```bash
# Start development environment
docker-compose up -d

# With hot reload
docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d
```

### Production

```bash
# Start production environment
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# With production environment file
docker-compose --env-file docker-compose.prod.env -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Monitoring & Observability

### Metrics Collection

- **Prometheus**: Metrics collection and storage
- **Grafana**: Dashboards and visualization
- **Loki**: Log aggregation
- **Promtail**: Log collection

### Security Monitoring

- All services log to structured format
- Prometheus metrics for security events
- Grafana dashboards for security monitoring

## Best Practices

### 1. Environment Variables

Never commit production passwords to version control:

```bash
# Create production environment file
cp docker-compose.prod.env.example docker-compose.prod.env

# Edit with secure passwords
nano docker-compose.prod.env
```

### 2. SSL Certificate Management

Let's Encrypt certificates are automatically managed:
- Certificates stored in `traefik_letsencrypt` volume
- Automatic renewal
- Fallback to HTTP for certificate challenges

### 3. Network Security

- Backend network is internal-only
- No direct access to databases from outside
- All communication through microservices

### 4. Service Discovery

Services communicate using Docker service names:
- `postgres:5432`
- `redis:6379`
- `rabbitmq:5672`

## Troubleshooting

### Network Connectivity

```bash
# Check network connectivity
docker network ls
docker network inspect english-app-frontend
docker network inspect english-app-backend

# Test service connectivity
docker exec -it user-services ping postgres
docker exec -it user-services ping redis
```

### Security Verification

```bash
# Check exposed ports
docker ps --format "table {{.Names}}\t{{.Ports}}"

# Verify no external database access
telnet localhost 5432  # Should fail in production
```

### SSL Certificate Issues

```bash
# Check certificate status
docker logs traefik | grep -i certificate

# Force certificate renewal
docker restart traefik
```

## Migration Guide

### From Development to Production

1. **Update Environment Variables**
   ```bash
   cp docker-compose.prod.env.example docker-compose.prod.env
   # Edit with production values
   ```

2. **Deploy Production Configuration**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

3. **Verify Security**
   ```bash
   # Check no external database ports
   docker ps | grep -E "(5432|6379|5672)"
   # Should only show internal services
   ```

### Backup Before Migration

```bash
# Backup all data
docker-compose exec postgres pg_dump -U user english_app > backup.sql
docker-compose exec mongodb mongodump --out /backup
```

## Security Checklist

- [ ] All passwords changed from defaults
- [ ] No external database ports in production
- [ ] HTTPS enabled with valid certificates
- [ ] All volumes persisted
- [ ] Network isolation implemented
- [ ] Monitoring configured
- [ ] Backup strategy in place
- [ ] Environment variables secured
- [ ] SSL certificates auto-renewing
- [ ] Security monitoring active
