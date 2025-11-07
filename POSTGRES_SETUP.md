# Quick PostgreSQL Setup Guide

## Overview

The application now supports both PostgreSQL and SQLite databases. PostgreSQL is recommended for production use.

## Quick Start with Docker Compose (Recommended)

The easiest way to get started is using Docker Compose:

### 1. Start Services

```bash
# Start PostgreSQL, Redis, and MinIO
docker-compose up -d

# Check services are running
docker-compose ps
```

### 2. Configure Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with these PostgreSQL settings:
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=moviedb
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-here
```

### 3. Run the Application

```bash
# The database tables will be created automatically
go run cmd/server/main.go
```

You should see:
```
Connecting to PostgreSQL database...
Database connection established (postgres) and migrated
```

### 4. Create Your First User

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "password123",
    "first_name": "Admin",
    "last_name": "User"
  }'
```

## Accessing Services

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | - |
| Swagger UI | http://localhost:8080/swagger/index.html | - |
| PostgreSQL | localhost:5432 | postgres/postgres |
| MinIO Console | http://localhost:9001 | minioadmin/minioadmin123 |
| Redis | localhost:6379 | (no auth) |

## Database Access

Connect to PostgreSQL:

```bash
# Using psql
docker exec -it movie-db-postgres psql -U postgres -d moviedb

# Or from host if you have psql installed
psql -h localhost -U postgres -d moviedb
```

Useful PostgreSQL commands:
```sql
-- List tables
\dt

-- Describe a table
\d users

-- View users
SELECT id, username, email, role FROM users;

-- View movies
SELECT id, title, director, rating FROM movies;
```

## Switching Between Databases

### Use PostgreSQL (Production)

In `.env`:
```bash
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=moviedb
```

### Use SQLite (Development/Testing)

In `.env`:
```bash
DB_DRIVER=sqlite
DB_NAME=movies.db
```

## Managing Docker Services

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# View PostgreSQL logs
docker-compose logs -f postgres

# Restart a service
docker-compose restart postgres

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v
```

## Backup and Restore

### Backup Database

```bash
# Create backup
docker exec movie-db-postgres pg_dump -U postgres moviedb > backup.sql

# Or with timestamp
docker exec movie-db-postgres pg_dump -U postgres moviedb > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Restore Database

```bash
# Restore from backup
docker exec -i movie-db-postgres psql -U postgres -d moviedb < backup.sql
```

## Troubleshooting

### Services Won't Start

```bash
# Check if ports are already in use
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
lsof -i :9000  # MinIO

# Stop conflicting services or change ports in docker-compose.yml
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Reset Everything

```bash
# Stop and remove everything including volumes
docker-compose down -v

# Start fresh
docker-compose up -d

# The app will recreate tables on next startup
```

## Performance Tips

1. **Connection Pooling**: Configured automatically by GORM
2. **Indexes**: Created automatically for primary keys and unique constraints
3. **Query Optimization**: Use GORM's preloading for related data

## Production Deployment

For production, consider:

1. **Use managed PostgreSQL** (AWS RDS, Google Cloud SQL, etc.)
2. **Enable SSL**: Set `DB_SSLMODE=require`
3. **Use strong passwords**: Change default credentials
4. **Set up backups**: Automated daily backups
5. **Monitor performance**: Use tools like pgAdmin, pg_stat_statements
6. **Connection pooling**: Consider PgBouncer for many connections

## Next Steps

1. ✅ PostgreSQL is configured
2. Create your admin user
3. Test the API endpoints
4. Set up production database
5. Configure automated backups

## Full Documentation

For detailed migration guide and advanced configuration, see:
- [POSTGRESQL_MIGRATION.md](./POSTGRESQL_MIGRATION.md) - Complete migration guide
- [USER_SERVICE.md](./USER_SERVICE.md) - User service documentation

## Support

Common issues and solutions:

**Port 5432 already in use:**
- Stop local PostgreSQL: `brew services stop postgresql`
- Or change port in docker-compose.yml

**Permission denied:**
- Ensure Docker has necessary permissions
- Check file ownership if using mounted volumes

**Connection timeout:**
- Wait for healthcheck to pass: `docker-compose ps`
- Check firewall settings
