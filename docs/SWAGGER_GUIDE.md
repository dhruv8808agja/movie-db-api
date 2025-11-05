# Swagger/OpenAPI Documentation Guide

## What is Swagger/OpenAPI?

**OpenAPI** is a specification for describing REST APIs in a standardized format. **Swagger** is a set of tools that work with OpenAPI specifications.

### Why Use It?

1. **Interactive Documentation** - Automatically generates a web UI where you can:
   - View all your API endpoints
   - See request/response formats
   - **Try out APIs directly in the browser** (no Postman needed!)

2. **Client Generation** - Generate client libraries in multiple languages automatically

3. **API Contract** - Serves as a contract between frontend and backend teams

4. **Always Up-to-Date** - Documentation is generated from your code comments, so it stays in sync

## How It Works in Go

For Go projects, we use a tool called **swag** that:

1. **Reads special comments** in your code (called annotations)
2. **Generates** OpenAPI specification files (JSON/YAML)
3. **Serves** a beautiful web UI (Swagger UI) to interact with your API

## Example: Before and After

### Before (no documentation):
You have to manually tell developers:
- "To create a movie, POST to /movies with JSON body containing title, description, etc."
- "The rating must be between 0 and 10"
- "You need a Bearer token in the Authorization header"

### After (with Swagger):
Developers visit `http://localhost:8080/swagger/index.html` and see:
- All endpoints listed clearly
- Click "Try it out" to test APIs directly
- See example requests and responses
- Know exactly what authentication is needed

## The Annotations

We'll add special comments above our handler functions. Here's what they look like:

```go
// CreateMovie creates a new movie
// @Summary      Create a new movie
// @Description  Creates a new movie with the provided details
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        movie  body      models.Movie  true  "Movie object"
// @Success      201    {object}  models.Movie
// @Failure      400    {object}  map[string]string
// @Security     BearerAuth
// @Router       /movies [post]
func CreateMovie(c *gin.Context) {
    // ... your code
}
```

### What Each Annotation Means:

- `@Summary` - Short description (shown in list)
- `@Description` - Detailed description
- `@Tags` - Groups related endpoints together (e.g., "movies", "auth")
- `@Accept` - What content type the endpoint accepts (json, xml, etc.)
- `@Produce` - What content type it returns
- `@Param` - Describes parameters (body, query, path, header)
- `@Success` - Successful response code and type
- `@Failure` - Error response codes and types
- `@Security` - What authentication is required
- `@Router` - The endpoint path and HTTP method

## The Process

1. **Install swag CLI tool** - Command-line tool to generate docs
2. **Add annotations** - Special comments above your functions
3. **Run swag init** - Generates the documentation files
4. **Import generated docs** - Add to your main.go
5. **Visit Swagger UI** - Open in browser to see and test your API

## File Structure

After setup, you'll have:
```
movie-db-api/
├── docs/
│   ├── docs.go         # Generated - contains OpenAPI spec
│   ├── swagger.json    # Generated - OpenAPI spec in JSON
│   └── swagger.yaml    # Generated - OpenAPI spec in YAML
└── cmd/server/main.go  # Import docs and setup route
```

The `docs/` folder is auto-generated - you never edit these files manually!

## Let's Get Started!

Follow along as we implement this step by step...
