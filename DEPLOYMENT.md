# Deployment Guide

## Quick Start

### Development

```bash
# 1. Clone and setup
git clone <repository-url>
cd movie-db-api
make init

# 2. Start services
make docker-up

# 3. Run application
make run
```

### Production

```bash
# 1. Setup server
ssh user@server
cd /opt/movie-db-api

# 2. Configure environment
cp .env.production.example .env.production
# Edit .env.production with secure values

# 3. Start services
make docker-prod-up

# 4. Verify deployment
make health
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Load Balancer                   │
│                   (Nginx/ALB)                    │
└────────────────────┬────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
┌───────▼──────┐         ┌────────▼────────┐
│  App Server  │         │   App Server    │
│  (Docker)    │         │   (Docker)      │
└───────┬──────┘         └────────┬────────┘
        │                         │
        └────────────┬────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
┌───────▼──────┐  ┌──────▼──────┐  ┌────────▼─────┐
│  PostgreSQL  │  │    Redis    │  │    MinIO     │
│  (Database)  │  │   (Cache)   │  │  (Storage)   │
└──────────────┘  └─────────────┘  └──────────────┘
```

## Deployment Methods

### Method 1: Docker Compose (Recommended)

**Pros:**
- Easy setup
- All services in one place
- Good for small to medium deployments

**Setup:**
```bash
# Production deployment
docker compose -f docker-compose.prod.yml up -d

# Check status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f app
```

### Method 2: Kubernetes

**Pros:**
- High availability
- Auto-scaling
- Better for large deployments

**Setup:**
```bash
# Apply configurations
kubectl apply -f k8s/

# Check pods
kubectl get pods

# View logs
kubectl logs -f deployment/movie-db-api
```

### Method 3: Cloud Platforms

#### AWS ECS

```bash
# Build and push
docker build -t movie-db-api .
aws ecr get-login-password | docker login --username AWS --password-stdin <ecr-url>
docker tag movie-db-api:latest <ecr-url>/movie-db-api:latest
docker push <ecr-url>/movie-db-api:latest

# Deploy to ECS
aws ecs update-service --cluster movie-db --service api --force-new-deployment
```

#### Google Cloud Run

```bash
# Build and deploy
gcloud builds submit --tag gcr.io/<project-id>/movie-db-api
gcloud run deploy movie-db-api --image gcr.io/<project-id>/movie-db-api --platform managed
```

#### Heroku

```bash
# Deploy
heroku container:push web -a movie-db-api
heroku container:release web -a movie-db-api
```

## Environment Configuration

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | Database password | `SecurePass123!` |
| `JWT_SECRET` | JWT signing key | Min 32 chars |
| `MINIO_SECRET_KEY` | MinIO secret | Min 32 chars |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | API port | `8080` |
| `GIN_MODE` | Gin mode | `release` |
| `LOG_LEVEL` | Log level | `info` |

## Database Migration

### Initial Setup

```bash
# Using Make
make db-create
make db-migrate

# Or manually
docker exec movie-db-postgres createdb -U postgres moviedb
go run cmd/server/main.go  # Auto-migrates on startup
```

### Backup

```bash
# Create backup
make db-backup

# Or manually
docker exec movie-db-postgres pg_dump -U postgres moviedb > backup.sql
```

### Restore

```bash
# Restore from backup
make db-restore BACKUP_FILE=backup.sql

# Or manually
docker exec -i movie-db-postgres psql -U postgres -d moviedb < backup.sql
```

## Monitoring

### Health Checks

```bash
# Application health
curl https://api.example.com/health

# Response
{
  "status": "ok",
  "service": "movie-db-api"
}
```

### Metrics

```bash
# Prometheus metrics
curl https://api.example.com/metrics

# Key metrics:
# - http_requests_total
# - http_request_duration_seconds
# - go_goroutines
```

### Logs

```bash
# Application logs
make logs

# Or with Docker
docker compose logs -f app

# Filter logs
docker compose logs --tail=100 app | grep ERROR
```

## Scaling

### Horizontal Scaling

```bash
# Scale to 3 instances
docker compose -f docker-compose.prod.yml up -d --scale app=3

# Verify
docker compose ps
```

### Vertical Scaling

Update `docker-compose.prod.yml`:
```yaml
app:
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 2G
      reservations:
        cpus: '1'
        memory: 1G
```

## Security

### SSL/TLS Setup

```bash
# Using Let's Encrypt
certbot --nginx -d api.example.com

# Or add certificates
mkdir -p nginx/ssl
cp cert.pem nginx/ssl/
cp key.pem nginx/ssl/
```

### Firewall Configuration

```bash
# Allow only necessary ports
ufw allow 22/tcp   # SSH
ufw allow 80/tcp   # HTTP
ufw allow 443/tcp  # HTTPS
ufw enable
```

### Security Scan

```bash
# Scan for vulnerabilities
make security-scan
make vuln-check

# Scan Docker image
trivy image movie-db-api:latest
```

## Backup Strategy

### Automated Backups

Create backup script `/opt/scripts/backup.sh`:
```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"

# Database backup
docker exec movie-db-postgres pg_dump -U postgres moviedb > $BACKUP_DIR/db_$DATE.sql

# Compress
gzip $BACKUP_DIR/db_$DATE.sql

# Upload to S3
aws s3 cp $BACKUP_DIR/db_$DATE.sql.gz s3://backups/movie-db/

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +30 -delete
```

Add to crontab:
```bash
# Daily backup at 2 AM
0 2 * * * /opt/scripts/backup.sh
```

## Rollback Procedure

### Quick Rollback

```bash
# Using Make
make prod-rollback

# Or manually
VERSION=previous docker compose -f docker-compose.prod.yml up -d
```

### Full Rollback

```bash
# 1. Stop current version
docker compose -f docker-compose.prod.yml down

# 2. Checkout previous version
git checkout <previous-tag>

# 3. Restore database
make db-restore BACKUP_FILE=backup_before_deployment.sql

# 4. Deploy
docker compose -f docker-compose.prod.yml up -d

# 5. Verify
make health
```

## Troubleshooting

### Application Won't Start

**Check logs:**
```bash
docker compose logs app
```

**Common issues:**
- Database connection failed → Check DB_PASSWORD
- Port already in use → Change SERVER_PORT
- Missing environment variable → Check .env

### Database Connection Issues

**Test connection:**
```bash
docker exec movie-db-postgres psql -U postgres -d moviedb -c "SELECT 1"
```

**Solutions:**
- Wait for database to be ready (check healthcheck)
- Verify credentials in .env
- Check network connectivity

### High Memory Usage

**Check resource usage:**
```bash
docker stats
```

**Solutions:**
- Increase memory limits
- Check for memory leaks
- Optimize database queries
- Add connection pooling

### Slow Performance

**Profile application:**
```bash
go tool pprof http://localhost:8080/debug/pprof/profile
```

**Solutions:**
- Add database indexes
- Enable Redis caching
- Optimize slow queries
- Scale horizontally

## Maintenance

### Update Dependencies

```bash
# Update Go dependencies
make deps-update

# Update Docker images
docker compose pull
docker compose up -d
```

### Database Maintenance

```bash
# Vacuum database
docker exec movie-db-postgres psql -U postgres -d moviedb -c "VACUUM ANALYZE"

# Check database size
docker exec movie-db-postgres psql -U postgres -c "\l+"
```

### Cleanup

```bash
# Remove old Docker images
docker image prune -a

# Remove old logs
find /var/log -name "*.log" -mtime +30 -delete

# Clear Redis cache
docker exec movie-db-redis redis-cli FLUSHALL
```

## Performance Tuning

### PostgreSQL

Edit `postgresql.conf`:
```ini
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 2621kB
min_wal_size = 1GB
max_wal_size = 4GB
```

### Redis

Edit redis configuration:
```ini
maxmemory 512mb
maxmemory-policy allkeys-lru
```

### Application

Connection pooling in `.env`:
```bash
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=3600
```

## Monitoring Setup

### Prometheus + Grafana

```bash
# Add to docker-compose
prometheus:
  image: prom/prometheus
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana
  ports:
    - "3000:3000"
```

### Alert Manager

Configure alerts in `alertmanager.yml`:
```yaml
route:
  receiver: 'slack'

receivers:
  - name: 'slack'
    slack_configs:
      - api_url: '<slack-webhook>'
        channel: '#alerts'
```

## Support

For deployment issues:
1. Check logs: `make logs`
2. Verify health: `make health`
3. Review documentation
4. Contact DevOps team

## Additional Resources

- [CI/CD Documentation](./CI_CD.md)
- [PostgreSQL Migration](./POSTGRESQL_MIGRATION.md)
- [User Service](./USER_SERVICE.md)
- [Docker Documentation](https://docs.docker.com/)
