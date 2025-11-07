# PostgreSQL Migration Guide

## Overview

The application has been updated to support both PostgreSQL and SQLite databases. This guide will help you migrate from SQLite to PostgreSQL.

## Why PostgreSQL?

- **Better Performance**: PostgreSQL handles concurrent connections better than SQLite
- **Scalability**: Designed for production workloads with large datasets
- **Advanced Features**: Full-text search, JSON support, complex queries
- **ACID Compliance**: Better transaction handling and data integrity
- **Production Ready**: Suitable for deployment in cloud environments

## Prerequisites

### Install PostgreSQL

**macOS (using Homebrew):**
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Windows:**
Download and install from: https://www.postgresql.org/download/windows/

**Docker (Recommended for Development):**
```bash
docker run --name moviedb-postgres \
  -e POSTGRES_PASSWORD=yourpassword \
  -e POSTGRES_DB=moviedb \
  -p 5432:5432 \
  -d postgres:16-alpine
```

## Migration Steps

### Step 1: Create PostgreSQL Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE moviedb;

# Create user (optional)
CREATE USER moviedb_user WITH ENCRYPTED PASSWORD 'your_secure_password';

# Grant privileges
GRANT ALL PRIVILEGES ON DATABASE moviedb TO moviedb_user;

# Exit psql
\q
```

### Step 2: Configure Environment Variables

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` and configure PostgreSQL settings:
```bash
# Database Configuration
DB_DRIVER=postgres

# PostgreSQL Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=moviedb
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# ... other configurations
```

### Step 3: Run Database Migrations

The application will automatically create tables when it starts:

```bash
# Run the application
go run cmd/server/main.go
```

The GORM auto-migration will create all necessary tables:
- `users` - User accounts and profiles
- `movies` - Movie metadata
- `videos` - Video files
- `upload_sessions` - Chunked upload tracking
- `transcoded_videos` - Transcoded video variants
- `transcoding_jobs` - Video transcoding jobs

### Step 4: Verify Database Connection

Check the application logs for:
```
Connecting to PostgreSQL database...
Database connection established (postgres) and migrated
```

You can also verify tables in PostgreSQL:
```bash
psql -U postgres -d moviedb -c "\dt"
```

Expected output:
```
                 List of relations
 Schema |        Name        | Type  |  Owner
--------+--------------------+-------+----------
 public | movies             | table | postgres
 public | transcoded_videos  | table | postgres
 public | transcoding_jobs   | table | postgres
 public | upload_sessions    | table | postgres
 public | users              | table | postgres
 public | videos             | table | postgres
```

## Data Migration (Optional)

If you have existing data in SQLite and want to migrate it to PostgreSQL:

### Option 1: Using pgloader (Recommended)

```bash
# Install pgloader
brew install pgloader  # macOS
sudo apt install pgloader  # Ubuntu

# Create migration script
cat > migrate.load << EOF
LOAD DATABASE
     FROM sqlite:///path/to/movies.db
     INTO postgresql://postgres:password@localhost/moviedb
;
EOF

# Run migration
pgloader migrate.load
```

### Option 2: Manual Export/Import

```bash
# Export from SQLite
sqlite3 movies.db .dump > dump.sql

# Edit dump.sql to make it PostgreSQL compatible
# (convert SQLite syntax to PostgreSQL)

# Import to PostgreSQL
psql -U postgres -d moviedb -f dump.sql
```

### Option 3: Application-Level Migration Script

Create a Go script to migrate data:

```go
package main

import (
    "log"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/dhruv8808agja/movie-db-api/pkg/models"
)

func main() {
    // Connect to SQLite
    sqliteDB, err := gorm.Open(sqlite.Open("movies.db"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Connect to PostgreSQL
    dsn := "host=localhost user=postgres password=yourpass dbname=moviedb sslmode=disable"
    postgresDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Migrate Users
    var users []models.User
    sqliteDB.Find(&users)
    for _, user := range users {
        postgresDB.Create(&user)
    }

    // Migrate Movies
    var movies []models.Movie
    sqliteDB.Find(&movies)
    for _, movie := range movies {
        postgresDB.Create(&movie)
    }

    // ... migrate other tables

    log.Println("Migration completed!")
}
```

## Configuration Options

### Database Driver Selection

You can switch between databases by changing the `DB_DRIVER` environment variable:

```bash
# Use PostgreSQL
DB_DRIVER=postgres

# Use SQLite (for development/testing)
DB_DRIVER=sqlite
DB_NAME=movies.db
```

### Connection Pool Settings (Advanced)

For production, you may want to configure connection pooling in `internal/storage/db.go`:

```go
sqlDB, err := DB.DB()
if err != nil {
    log.Fatal(err)
}

// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
sqlDB.SetMaxIdleConns(10)

// SetMaxOpenConns sets the maximum number of open connections to the database
sqlDB.SetMaxOpenConns(100)

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
sqlDB.SetConnMaxLifetime(time.Hour)
```

### SSL/TLS Configuration

For production environments, enable SSL:

```bash
# In .env
DB_SSLMODE=require  # or verify-full for strict verification
```

## Testing the Migration

### 1. Create a Test User

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 3. Create a Movie

```bash
curl -X POST http://localhost:8080/movies \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Movie",
    "description": "Test Description",
    "director": "Test Director",
    "release_date": "2024-01-01T00:00:00Z",
    "genres": ["Action", "Drama"],
    "rating": 8.5
  }'
```

### 4. Verify in Database

```bash
psql -U postgres -d moviedb

-- Check users
SELECT id, username, email, role FROM users;

-- Check movies
SELECT id, title, director FROM movies;
```

## Troubleshooting

### Connection Refused

```
Error: failed to connect database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solution:** Ensure PostgreSQL is running
```bash
# Check status
brew services list  # macOS
sudo systemctl status postgresql  # Linux

# Start PostgreSQL
brew services start postgresql@16  # macOS
sudo systemctl start postgresql  # Linux
```

### Authentication Failed

```
Error: pq: password authentication failed for user "postgres"
```

**Solution:** Check your credentials in `.env` or reset PostgreSQL password
```bash
# For local development, you can modify pg_hba.conf
# to use 'trust' authentication (not recommended for production)
sudo vi /etc/postgresql/*/main/pg_hba.conf
# Change 'md5' or 'scram-sha-256' to 'trust' for local connections
sudo systemctl restart postgresql
```

### Database Does Not Exist

```
Error: pq: database "moviedb" does not exist
```

**Solution:** Create the database
```bash
createdb -U postgres moviedb
```

### Migration Failed

If auto-migration fails, check:
1. Database user has CREATE TABLE privileges
2. Schema is correct for PostgreSQL
3. Check application logs for specific errors

## Docker Compose Setup (Recommended)

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: moviedb-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: yourpassword
      POSTGRES_DB: moviedb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: moviedb-redis
    ports:
      - "6379:6379"

  minio:
    image: minio/minio:latest
    container_name: moviedb-minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  minio_data:
```

Run with Docker Compose:
```bash
docker-compose up -d
```

## Performance Tips

1. **Create Indexes** for frequently queried columns:
```sql
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_movies_title ON movies(title);
CREATE INDEX idx_movies_release_date ON movies(release_date);
```

2. **Enable Query Logging** for development:
```go
// In storage/db.go
DB, err = gorm.Open(dialector, &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

3. **Use Connection Pooling** (see Configuration Options above)

## Rollback to SQLite

If you need to rollback to SQLite:

```bash
# In .env
DB_DRIVER=sqlite
DB_NAME=movies.db

# Restart the application
```

## Next Steps

1. Set up automated backups for PostgreSQL
2. Configure SSL/TLS for production
3. Implement database monitoring
4. Set up read replicas for scaling
5. Consider using a connection pooler like PgBouncer

## Support

If you encounter issues:
1. Check application logs
2. Check PostgreSQL logs: `tail -f /var/log/postgresql/postgresql-*.log`
3. Verify environment variables
4. Test connection with `psql`

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [GORM PostgreSQL Driver](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL)
- [Docker PostgreSQL](https://hub.docker.com/_/postgres)
