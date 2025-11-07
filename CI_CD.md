# CI/CD Pipeline Documentation

## Overview

This project uses GitHub Actions for Continuous Integration and Continuous Deployment (CI/CD). The pipeline automatically tests, builds, and deploys the application.

## Pipeline Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Commit    │────▶│  CI Tests   │────▶│   Build     │
│   to Repo   │     │   & Lint    │     │   Docker    │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
                                        ┌─────────────┐
                                        │   Deploy    │
                                        │ to Staging/ │
                                        │ Production  │
                                        └─────────────┘
```

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

**Jobs:**

#### Test Job
- Sets up PostgreSQL, Redis, and MinIO services
- Runs Go version 1.21
- Downloads and verifies dependencies
- Runs `go vet` for code issues
- Runs `staticcheck` for static analysis
- Executes tests with race detection and coverage
- Uploads coverage to Codecov

#### Lint Job
- Runs `golangci-lint` with latest version
- Checks code quality and style

#### Build Job
- Builds the application binary
- Uploads artifact for 7 days

**Environment Variables:**
```yaml
DB_DRIVER: postgres
DB_HOST: localhost
DB_PORT: 5432
DB_USER: postgres
DB_PASSWORD: postgres
DB_NAME: moviedb_test
JWT_SECRET: test-secret-key-for-ci
```

### 2. Docker Workflow (`.github/workflows/docker.yml`)

**Triggers:**
- Push to `main` branch
- Tags matching `v*`
- Pull requests to `main`

**Features:**
- Builds multi-platform images (amd64, arm64)
- Pushes to GitHub Container Registry (ghcr.io)
- Uses BuildKit cache for faster builds
- Generates artifact attestations
- Tags with:
  - Branch name
  - Git SHA
  - Semantic version (if tag)

**Image Naming:**
```
ghcr.io/<username>/movie-db-api:main
ghcr.io/<username>/movie-db-api:sha-<git-sha>
ghcr.io/<username>/movie-db-api:v1.0.0
```

### 3. Deploy Workflow (`.github/workflows/deploy.yml`)

**Triggers:**
- Push to `main` (staging) or on release (production)
- Manual workflow dispatch

**Deployment Steps:**
1. SSH to deployment server
2. Pull latest code
3. Pull Docker image from registry
4. Update environment variables
5. Restart services
6. Run health check
7. Send Slack notification

**Rollback:**
- Automatic rollback on deployment failure
- Reverts to previous Docker image
- Notifies team via Slack

## Setup Instructions

### 1. Repository Secrets

Configure the following secrets in GitHub:
`Settings → Secrets and variables → Actions → New repository secret`

**Required Secrets:**

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `DB_PASSWORD` | Database password | `SecurePass123!` |
| `JWT_SECRET` | JWT signing key | `your-64-char-secret` |
| `MINIO_SECRET_KEY` | MinIO secret | `your-minio-secret` |
| `DEPLOY_HOST` | Server hostname | `api.example.com` |
| `DEPLOY_USER` | SSH username | `deploy` |
| `DEPLOY_SSH_KEY` | Private SSH key | `-----BEGIN RSA...` |
| `DEPLOY_PORT` | SSH port (optional) | `22` |
| `SLACK_WEBHOOK` | Slack webhook URL | `https://hooks.slack...` |

**Optional Secrets:**
- `CODECOV_TOKEN` - For Codecov integration
- `SENTRY_DSN` - For error tracking

### 2. Enable GitHub Packages

1. Go to repository `Settings → Actions → General`
2. Under "Workflow permissions", select:
   - ✅ Read and write permissions
   - ✅ Allow GitHub Actions to create and approve pull requests

### 3. Container Registry Access

Images are pushed to GitHub Container Registry (ghcr.io):

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull image
docker pull ghcr.io/username/movie-db-api:main
```

### 4. Server Setup

Prepare your deployment server:

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Create deployment directory
sudo mkdir -p /opt/movie-db-api
sudo chown $USER:$USER /opt/movie-db-api

# Clone repository
cd /opt/movie-db-api
git clone <repository-url> .

# Setup environment
cp .env.production.example .env
# Edit .env with production values

# Setup SSH key for GitHub Actions
mkdir -p ~/.ssh
# Add deploy key to authorized_keys
```

## Local Development with CI/CD

### Run CI Tests Locally

```bash
# Using Make
make ci-test

# Or manually
docker compose up -d postgres redis minio
sleep 5
go test -v -race ./...
docker compose down
```

### Build Docker Image Locally

```bash
# Using Make
make docker-build

# Or using Docker
docker build -t movie-db-api:local .

# Run locally
docker run -p 8080:8080 \
  -e DB_DRIVER=postgres \
  -e DB_HOST=host.docker.internal \
  movie-db-api:local
```

### Test Production Deployment Locally

```bash
# Start production stack
make docker-prod-up

# Or manually
docker compose -f docker-compose.prod.yml up -d

# Check logs
docker compose -f docker-compose.prod.yml logs -f

# Health check
curl http://localhost:8080/health
```

## Deployment Workflows

### Staging Deployment

```bash
# Automatic on push to main
git push origin main

# Or manual trigger
gh workflow run deploy.yml -f environment=staging
```

### Production Deployment

```bash
# Create release
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Or manual trigger
gh workflow run deploy.yml -f environment=production
```

### Rollback

```bash
# Using Make
make prod-rollback

# Or manually
ssh deploy@server
cd /opt/movie-db-api
VERSION=previous docker compose -f docker-compose.prod.yml up -d
```

## Monitoring CI/CD

### GitHub Actions Dashboard

- View workflow runs: `Actions` tab in repository
- Check build status: Badge in README
- Review logs: Click on workflow run

### Workflow Status Badge

Add to README.md:

```markdown
![CI](https://github.com/username/movie-db-api/workflows/CI/badge.svg)
![Docker](https://github.com/username/movie-db-api/workflows/Docker/badge.svg)
```

### Health Checks

```bash
# Application health
curl https://api.example.com/health

# Docker container health
docker ps --format "table {{.Names}}\t{{.Status}}"

# Service logs
docker compose logs -f app
```

## Troubleshooting

### CI Tests Failing

**Issue:** Tests fail in CI but pass locally

**Solutions:**
```bash
# Run tests with same environment as CI
make ci-test

# Check service health
docker compose ps

# View service logs
docker compose logs postgres redis minio
```

### Docker Build Fails

**Issue:** Docker build fails or times out

**Solutions:**
```bash
# Clear Docker cache
docker builder prune -a

# Build locally first
make docker-build

# Check Dockerfile syntax
docker build --no-cache .
```

### Deployment Fails

**Issue:** Deployment fails or times out

**Solutions:**
```bash
# Check SSH connection
ssh deploy@server

# Verify environment variables
ssh deploy@server "cat /opt/movie-db-api/.env"

# Check Docker images
ssh deploy@server "docker images"

# View deployment logs
ssh deploy@server "cd /opt/movie-db-api && docker compose logs"
```

### Health Check Fails

**Issue:** Health check returns error after deployment

**Solutions:**
```bash
# Check application logs
docker compose logs app

# Check database connection
docker exec movie-db-postgres-prod psql -U postgres -d moviedb -c "SELECT 1"

# Check service connectivity
docker compose exec app wget -O- http://localhost:8080/health
```

## Best Practices

### 1. Branch Strategy

```
main (staging) → automatic deployment to staging
tags (v*) → manual deployment to production
feature/* → CI tests only
```

### 2. Commit Messages

Follow conventional commits:
```
feat: add user authentication
fix: resolve database connection issue
docs: update API documentation
test: add integration tests
ci: update deployment workflow
```

### 3. Version Tagging

Use semantic versioning:
```bash
v1.0.0 - Major release
v1.1.0 - Minor feature
v1.1.1 - Patch/bugfix
```

### 4. Testing Strategy

- Write tests for all new features
- Maintain >80% code coverage
- Run tests locally before pushing
- Review CI results before merging

### 5. Deployment Checklist

Before deploying to production:
- ✅ All CI tests pass
- ✅ Code reviewed and approved
- ✅ Documentation updated
- ✅ Database migrations tested
- ✅ Environment variables verified
- ✅ Backup created
- ✅ Rollback plan ready

## Performance Optimization

### Build Cache

GitHub Actions uses layer caching:
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

### Parallel Jobs

Tests and linting run in parallel:
```yaml
jobs:
  test: ...
  lint: ...  # Runs simultaneously
```

### Artifacts

Build artifacts are retained for 7 days:
```yaml
retention-days: 7
```

## Security

### Secret Management

- Never commit secrets to repository
- Use GitHub Secrets for sensitive data
- Rotate secrets regularly
- Use environment-specific secrets

### Image Scanning

Add vulnerability scanning:
```yaml
- name: Scan image
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
```

### HTTPS/TLS

- Use SSL/TLS in production
- Enable HSTS headers
- Use secure connection strings

## Maintenance

### Regular Updates

- Update GitHub Actions versions
- Update Docker base images
- Update Go dependencies
- Review security advisories

### Backup Strategy

- Automated daily database backups
- Test restore procedures monthly
- Store backups off-site
- Retain backups for 30 days

### Monitoring

- Set up application monitoring
- Configure alerts for failures
- Monitor resource usage
- Track deployment metrics

## Support

For issues with CI/CD:
1. Check workflow logs in GitHub Actions
2. Review this documentation
3. Consult team lead
4. Create issue in repository

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Documentation](https://docs.docker.com/)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Semantic Versioning](https://semver.org/)
