# Watch History Feature

## Overview

The Watch History feature tracks user viewing activity, allowing users to see their watch history, continue watching incomplete movies, and view statistics about their viewing habits.

## Database Model

### WatchHistory Table

The `watch_histories` table stores individual watch history entries:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint | Primary key |
| `user_id` | uint | Foreign key to users table (indexed) |
| `movie_id` | uint | Foreign key to movies table (indexed) |
| `watched_at` | timestamp | When the movie was watched (indexed) |
| `watch_duration` | int | Total watch time in seconds |
| `progress` | float32 | Completion percentage (0-100) |
| `completed` | bool | Whether the movie was fully watched |
| `last_position` | int | Last playback position in seconds |
| `watch_count` | int | Number of times watched (default: 1) |
| `quality` | string | Video quality (e.g., "1080p", "720p") |
| `device_type` | string | Device type (e.g., "web", "mobile", "tv") |

### Relationships

- `user_id` → `users.id` (many-to-one)
- `movie_id` → `movies.id` (many-to-one)

## API Endpoints

All endpoints require authentication (JWT Bearer token).

### 1. Record Watch History

**POST** `/watch-history`

Records or updates a watch history entry for a movie.

**Request Body:**
```json
{
  "movie_id": 1,
  "watch_duration": 3600,
  "last_position": 2700,
  "progress": 75.5,
  "completed": false,
  "quality": "1080p",
  "device_type": "web"
}
```

**Response (200 OK):**
```json
{
  "message": "Watch history recorded successfully",
  "watch_history": {
    "id": 1,
    "user_id": 1,
    "movie_id": 1,
    "watched_at": "2024-01-01T12:00:00Z",
    "watch_duration": 3600,
    "progress": 75.5,
    "completed": false,
    "last_position": 2700,
    "watch_count": 1,
    "quality": "1080p",
    "device_type": "web"
  }
}
```

**Behavior:**
- If a watch history entry exists for the same movie on the same day, it updates that entry
- Watch duration is accumulated (added to existing duration)
- Watch count is incremented
- Other fields are replaced with new values

### 2. Get Watch History

**GET** `/watch-history`

Retrieves the user's watch history with pagination and filtering.

**Query Parameters:**
- `page` (int, optional): Page number (default: 1)
- `page_size` (int, optional): Items per page (default: 20, max: 100)
- `movie_id` (int, optional): Filter by specific movie
- `completed` (bool, optional): Filter by completion status

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "movie_id": 1,
      "watched_at": "2024-01-01T12:00:00Z",
      "watch_duration": 3600,
      "progress": 75.5,
      "completed": false,
      "last_position": 2700,
      "watch_count": 1,
      "quality": "1080p",
      "device_type": "web",
      "movie": {
        "id": 1,
        "title": "The Matrix",
        "description": "...",
        "director": "The Wachowskis",
        "release_date": "1999-03-31T00:00:00Z",
        "genres": ["Action", "Sci-Fi"],
        "rating": 8.7
      }
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 42,
    "total_pages": 3
  }
}
```

### 3. Get Continue Watching

**GET** `/watch-history/continue-watching`

Retrieves movies that the user is currently watching (not completed, progress > 5%).

**Query Parameters:**
- `limit` (int, optional): Number of results (default: 10, max: 50)

**Response (200 OK):**
```json
[
  {
    "watch_history_id": 1,
    "movie_id": 1,
    "movie": {
      "id": 1,
      "title": "The Matrix",
      "description": "...",
      "director": "The Wachowskis",
      "release_date": "1999-03-31T00:00:00Z",
      "genres": ["Action", "Sci-Fi"],
      "rating": 8.7
    },
    "last_position": 2700,
    "progress": 75.5,
    "watched_at": "2024-01-01T12:00:00Z"
  }
]
```

### 4. Get Watch Statistics

**GET** `/watch-history/stats`

Retrieves aggregated watch statistics for the user.

**Response (200 OK):**
```json
{
  "total_movies_watched": 42,
  "total_watch_time": 151200,
  "completed_movies": 35,
  "in_progress_movies": 7,
  "favorite_genres": ["Action", "Drama", "Sci-Fi"],
  "last_watched_at": "2024-01-01T12:00:00Z",
  "average_progress": 82.5
}
```

**Fields:**
- `total_movies_watched`: Count of unique movies watched
- `total_watch_time`: Total watch time in seconds
- `completed_movies`: Count of fully watched movies
- `in_progress_movies`: Count of incomplete movies (progress > 5%)
- `favorite_genres`: Top 5 most watched genres
- `last_watched_at`: Timestamp of most recent watch
- `average_progress`: Average completion percentage

### 5. Delete Watch History Entry

**DELETE** `/watch-history/{id}`

Deletes a specific watch history entry.

**Path Parameters:**
- `id` (int): Watch history entry ID

**Response (200 OK):**
```json
{
  "message": "Watch history deleted successfully"
}
```

**Error Responses:**
- `403 Forbidden`: Attempting to delete another user's watch history
- `404 Not Found`: Watch history entry not found

### 6. Clear All Watch History

**DELETE** `/watch-history/clear`

Deletes all watch history entries for the authenticated user.

**Response (200 OK):**
```json
{
  "message": "Watch history cleared successfully",
  "deleted_count": 42
}
```

## Usage Examples

### Recording Watch Progress

```bash
# Record watch progress for a movie
curl -X POST http://localhost:8080/watch-history \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 1,
    "watch_duration": 1800,
    "last_position": 1800,
    "progress": 50.0,
    "completed": false,
    "quality": "1080p",
    "device_type": "web"
  }'
```

### Getting Continue Watching List

```bash
# Get movies user is currently watching
curl -X GET http://localhost:8080/watch-history/continue-watching?limit=5 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Getting Watch Statistics

```bash
# Get user's watch statistics
curl -X GET http://localhost:8080/watch-history/stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Getting Full Watch History

```bash
# Get paginated watch history
curl -X GET "http://localhost:8080/watch-history?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Filter by completed movies
curl -X GET "http://localhost:8080/watch-history?completed=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Filter by specific movie
curl -X GET "http://localhost:8080/watch-history?movie_id=1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Implementation Details

### Watch History Update Logic

When recording watch history:
1. Check if an entry exists for the same movie on the same day (today)
2. If exists:
   - Add new `watch_duration` to existing duration
   - Update `last_position`, `progress`, `completed`, `quality`, `device_type`
   - Increment `watch_count`
   - Update `watched_at` timestamp
3. If doesn't exist:
   - Create new entry with `watch_count = 1`

### Continue Watching Logic

Movies appear in continue watching if:
- `completed = false`
- `progress > 5%` (to filter out movies that were barely started)
- Returns the most recent watch entry for each movie
- Ordered by most recently watched

### Favorite Genres Calculation

1. Retrieve all movies the user has watched
2. Extract genres from each movie
3. Count frequency of each genre
4. Return top 5 most frequent genres

## Testing

The watch history feature includes comprehensive test coverage:

```bash
# Run watch history tests
go test ./internal/watchhistory/... -v

# Run with coverage
go test ./internal/watchhistory/... -cover
```

### Test Coverage

Tests cover:
- ✅ Recording new watch history
- ✅ Updating existing watch history
- ✅ Movie not found error handling
- ✅ Unauthorized access handling
- ✅ Retrieving watch history with pagination
- ✅ Filtering watch history
- ✅ Continue watching functionality
- ✅ Watch statistics calculation
- ✅ Deleting watch history entries
- ✅ Authorization checks for deletion
- ✅ Clearing all watch history

## Database Migrations

The `WatchHistory` model is automatically migrated when the application starts:

```go
// internal/storage/db.go
err = DB.AutoMigrate(
    &User,
    &Movie,
    // ...
    &WatchHistory,
)
```

For production environments, consider creating explicit migration files.

## Performance Considerations

### Indexes

The following fields are indexed for optimal query performance:
- `user_id` - For filtering by user
- `movie_id` - For filtering by movie
- `watched_at` - For sorting and date-based queries

### Optimization Tips

1. **Continue Watching Query**: Uses a subquery with GROUP BY for efficiency
2. **Watch Stats**: Multiple optimized queries instead of one complex query
3. **Pagination**: Always use pagination for watch history to avoid loading large datasets

## Future Enhancements

Potential features to add:
- Watch history export (CSV, JSON)
- Weekly/monthly viewing reports
- Viewing streaks tracking
- Recommendations based on watch history
- Social features (share watch history with friends)
- Watch time goals and achievements
- Parental controls integration

## Swagger Documentation

All endpoints are documented in Swagger. Access the documentation at:

```
http://localhost:8080/swagger/index.html
```

Look for the **watch-history** tag in the API documentation.

## Files Modified/Created

### New Files
- `pkg/models/watchhistory.go` - Data models
- `internal/watchhistory/handlers.go` - API handlers
- `internal/watchhistory/handlers_test.go` - Test suite

### Modified Files
- `internal/storage/db.go` - Added WatchHistory migration
- `cmd/server/main.go` - Added watch history routes
- `internal/testutil/testutil.go` - Added SeedTestMovies helper
- `docs/` - Auto-generated Swagger documentation

## Integration with Existing Features

The watch history feature integrates with:
- **Authentication**: All endpoints require JWT authentication
- **Movies**: Foreign key relationship to movies table
- **Users**: Foreign key relationship to users table
- **Recommendations**: Can be used to improve recommendation accuracy (future enhancement)
