# Recommendation Service

A comprehensive movie recommendation engine with multiple recommendation strategies.

## Features

### 1. Content-Based Filtering
Recommends movies based on:
- **Genre similarity**: Movies with similar genres to what the user has liked
- **Director preference**: Movies from directors whose work the user enjoys
- **Rating patterns**: Movies with similar ratings to user preferences

### 2. Collaborative Filtering
Recommends movies using:
- **User similarity**: Finds users with similar taste (cosine similarity)
- **Peer recommendations**: Suggests movies liked by similar users
- **Minimum threshold**: Requires at least 3 ratings to establish patterns

### 3. Trending Movies
Falls back to:
- **Popular content**: Most-viewed movies in last 30 days
- **Highly rated**: Top-rated movies when no interaction data exists

## API Endpoints

### Get Personalized Recommendations
```bash
GET /recommendations?limit=10
Authorization: Bearer <token>
```

**Response:**
```json
{
  "recommendations": [
    {
      "movie": {
        "id": 1,
        "title": "The Matrix",
        "director": "The Wachowskis",
        "genres": ["Action", "Sci-Fi"],
        "rating": 8.7
      },
      "score": 15.5,
      "reason_type": "content-based",
      "reason": "Based on your interest in Sci-Fi"
    }
  ],
  "count": 1
}
```

### Rate a Movie
```bash
POST /movies/:movieId/rate
Authorization: Bearer <token>

{
  "rating": 8.5
}
```

### Get User's Ratings
```bash
GET /ratings
Authorization: Bearer <token>
```

### Track Interaction
```bash
POST /interactions
Authorization: Bearer <token>

{
  "movie_id": 1,
  "interaction_type": "view",  // view, watch, like
  "watch_duration": 3600       // optional, in seconds
}
```

### Get Similar Movies
```bash
GET /movies/:movieId/similar?limit=10
```

## Algorithm Details

### Content-Based Scoring
- Genre match: +0.3 per matching genre
- Director match: +0.5
- Movie rating: +0.2 × rating value

### Collaborative Filtering
1. Build user rating vectors
2. Calculate cosine similarity between users
3. Filter users with similarity > 0.3
4. Aggregate recommendations weighted by similarity
5. Recommend highly-rated movies (rating ≥ 7.0)

### Trending Algorithm
- Counts interactions in last 30 days
- Weights by interaction count
- Falls back to highest-rated movies

## Database Schema

### UserMovieRating
```go
type UserMovieRating struct {
    UserID    uint
    MovieID   uint
    Rating    float32  // 0-10
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### UserInteraction
```go
type UserInteraction struct {
    ID              uint
    UserID          uint
    MovieID         uint
    InteractionType string   // view, watch, like, rate
    Rating          *float32 // optional
    WatchDuration   *int     // optional, in seconds
    CreatedAt       time.Time
}
```

## Testing

Run tests with coverage:
```bash
go test ./internal/recommendations/... -cover
```

Current coverage: **85.6%**

## Usage Example

```go
// Create engine
engine := recommendations.NewEngine()

// Get recommendations for user
recs, err := engine.GetRecommendations(userID, 10)

// Get content-based recommendations
contentRecs, err := engine.getContentBasedRecommendations(userID, 10)

// Get collaborative recommendations
collabRecs, err := engine.getCollaborativeRecommendations(userID, 10)

// Get trending movies
trending, err := engine.getTrendingMovies(10)
```

## Future Enhancements

1. **Matrix Factorization**: SVD/ALS for better collaborative filtering
2. **Deep Learning**: Neural collaborative filtering
3. **Real-time Updates**: Stream processing for instant recommendations
4. **A/B Testing**: Track recommendation effectiveness
5. **Diversity**: Ensure varied recommendations
6. **Personalized Trending**: Weight trending by user preferences
7. **Cold Start**: Better handling for new users/movies
8. **Explanation**: More detailed reasoning for recommendations
