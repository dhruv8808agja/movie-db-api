# Testing Documentation

This document describes the testing setup and how to run tests for the Movie DB API.

## Overview

The project includes comprehensive unit and integration tests covering:
- Movie validation logic
- JWT authentication
- Rate limiting middleware
- Movie CRUD operations
- Authentication flow

## Test Structure

```
internal/
├── auth/
│   ├── auth_integration_test.go  # Auth endpoint integration tests
│   └── jwt_test.go                # JWT unit tests
├── middleware/
│   └── rate_limiter_test.go      # Rate limiting tests
├── movies/
│   ├── validation_test.go         # Movie validation unit tests
│   └── movies_integration_test.go # Movie endpoints integration tests
└── testutil/
    └── testutil.go                # Test utilities and helpers
```

## Running Tests

### Prerequisites

1. **Set JWT_SECRET environment variable** (required for all tests):
   ```bash
   export JWT_SECRET=test-secret-key
   ```

2. **Redis** (optional, but recommended for rate limiting tests):
   - Rate limiting tests require Redis running on `localhost:6379`
   - If Redis is unavailable, rate limiting tests will be skipped
   - Other tests will run without Redis

### Run All Tests

```bash
JWT_SECRET=test-secret go test ./internal/... -v
```

### Run Specific Test Suites

**Movie validation tests:**
```bash
JWT_SECRET=test-secret go test ./internal/movies -v -run TestValidateMovie
```

**JWT authentication tests:**
```bash
JWT_SECRET=test-secret go test ./internal/auth -v -run TestJWT
```

**Integration tests:**
```bash
JWT_SECRET=test-secret go test ./internal/movies -v -run TestCreate
JWT_SECRET=test-secret go test ./internal/auth -v -run TestLogin
```

**Rate limiting tests:**
```bash
JWT_SECRET=test-secret go test ./internal/middleware -v -run TestRateLimiter
```

### Run Tests with Coverage

```bash
JWT_SECRET=test-secret go test ./internal/... -cover
```

Generate detailed coverage report:
```bash
JWT_SECRET=test-secret go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Test Categories

### Unit Tests

#### Movie Validation (internal/movies/validation_test.go)
Tests the `ValidateMovie` function with various scenarios:
- ✓ Valid movie data
- ✓ Empty/whitespace title
- ✓ Title too long (>200 chars)
- ✓ Description too long (>1000 chars)
- ✓ Director validation
- ✓ Release date bounds (1800 - current year + 5)
- ✓ Rating bounds (0-10)
- ✓ Genre validation (non-empty, max 10 genres)
- ✓ Multiple validation errors
- ✓ Edge cases

**Total:** 18 test cases

#### JWT Authentication (internal/auth/jwt_test.go)
Tests JWT token generation and middleware:
- ✓ Token generation
- ✓ Token parsing and validation
- ✓ Different usernames produce different tokens
- ✓ Valid token authentication
- ✓ Missing Authorization header
- ✓ Invalid auth format
- ✓ Invalid/malformed tokens
- ✓ Expired tokens
- ✓ Tampered tokens

**Total:** 9 test cases

#### Rate Limiter Configuration (internal/middleware/rate_limiter_test.go)
Tests rate limiter configuration:
- ✓ Default configuration values
- ✓ Custom environment variable values
- ✓ Invalid value handling

**Total:** 3 test cases

### Integration Tests

#### Movie Endpoints (internal/movies/movies_integration_test.go)
Tests complete HTTP request/response flow:
- ✓ Create movie (success & validation errors)
- ✓ Create movies in bulk
- ✓ Get movie by ID (success, not found, invalid ID)
- ✓ Update movie (success, not found, validation errors)
- ✓ Delete movie (success, not found)
- ✓ List movies with filters

**Total:** 14 test cases

#### Authentication (internal/auth/auth_integration_test.go)
Tests authentication endpoints and middleware integration:
- ✓ Login with valid credentials
- ✓ Login with invalid credentials (5 scenarios)
- ✓ Login with invalid JSON
- ✓ JWT middleware integration
- ✓ Unauthorized access attempts
- ✓ Complete auth flow (login → access protected route)

**Total:** 6 test cases

#### Rate Limiting (internal/middleware/rate_limiter_test.go)
Tests rate limiting behavior:
- ✓ Requests within limit
- ✓ Requests exceeding limit
- ✓ Different IPs tracked separately
- ✓ Rate limit headers
- ✓ Graceful degradation when Redis unavailable

**Total:** 5 test cases

## Test Utilities

The `testutil` package provides helpers for test setup:

### Database Setup
```go
testutil.SetupTestDB()  // Creates in-memory SQLite database
```

### Redis Setup
```go
testutil.SetupTestRedis()     // Creates Redis client (DB 15)
testutil.CleanupTestRedis()    // Flushes test Redis DB
```

### Environment Setup
```go
testutil.SetTestEnv()          // Sets test environment variables
testutil.ClearTestEnv()        // Clears test environment variables
```

### Complete Storage Init
```go
testutil.InitTestStorage()     // Initializes both DB and Redis
```

### Test Data Helpers
```go
testutil.CreateTestMovie()           // Creates a single test movie
testutil.CreateTestMovies(count)     // Creates multiple test movies
```

## Test Patterns

### Unit Test Pattern
```go
func TestValidateMovie_ValidMovie(t *testing.T) {
    movie := &models.Movie{
        Title:   "Test Movie",
        Rating:  8.5,
        // ... other fields
    }

    err := ValidateMovie(movie)
    assert.NoError(t, err)
}
```

### Integration Test Pattern
```go
func TestCreateMovie_Success(t *testing.T) {
    // Setup
    testutil.SetTestEnv()
    defer testutil.ClearTestEnv()
    storage.DB = testutil.SetupTestDB()

    // Create router and endpoint
    router := setupRouter()
    router.POST("/movies", CreateMovie)

    // Make request
    movie := testutil.CreateTestMovie()
    jsonData, _ := json.Marshal(movie)
    req := httptest.NewRequest("POST", "/movies", bytes.NewBuffer(jsonData))
    w := httptest.NewRecorder()

    // Execute
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

## Continuous Integration

For CI/CD pipelines, use:

```bash
#!/bin/bash
# Set required environment variables
export JWT_SECRET=test-secret-for-ci

# Run tests with race detector
go test ./internal/... -race -count=1

# Generate coverage
go test ./internal/... -coverprofile=coverage.out -covermode=atomic

# Check coverage threshold (optional)
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
```

## Troubleshooting

### Tests fail with "JWT_SECRET environment variable is not set"
**Solution:** Set the JWT_SECRET environment variable before running tests:
```bash
export JWT_SECRET=test-secret
```

### Rate limiter tests are skipped
**Cause:** Redis is not available
**Solution:**
- Start Redis: `redis-server`
- Or run without rate limiter tests: `go test ./internal/auth ./internal/movies -v`

### Database-related test failures
**Cause:** Tests may interfere with each other if run in parallel
**Solution:** Use `-count=1` flag to disable test caching and run fresh:
```bash
go test ./internal/... -count=1
```

## Test Coverage Summary

| Package      | Tests | Coverage |
|--------------|-------|----------|
| auth         | 15    | High     |
| middleware   | 8     | High     |
| movies       | 32    | High     |
| **Total**    | **55**| **High** |

## Best Practices

1. **Always set JWT_SECRET** before running tests
2. **Use testutil helpers** for consistent test setup
3. **Clean up resources** with `defer` statements
4. **Test both success and failure paths**
5. **Use descriptive test names** following Go conventions
6. **Run tests with `-count=1`** to ensure fresh runs
7. **Check coverage** regularly to maintain quality
