# API Documentation with Swagger

## 🎉 Congratulations! Your API Now Has Interactive Documentation

Your Movie DB API now has **Swagger/OpenAPI documentation** - an interactive web interface where you can view, understand, and **test your API directly from your browser**!

## 📖 What is Swagger?

Swagger is a tool that provides:
- **Interactive UI** - See all your endpoints in a beautiful web interface
- **Live Testing** - Try out API calls directly in the browser (no Postman needed!)
- **Auto-generated** - Documentation is created from your code comments
- **Always Up-to-Date** - Changes to code automatically reflect in docs when regenerated

## 🚀 How to Access Swagger UI

### Step 1: Start Your Server

Make sure you have the required environment variables set:

```bash
export JWT_SECRET=your-secret-key
export REDIS_ADDR=127.0.0.1:6379  # optional
```

Then start the server:

```bash
go run cmd/server/main.go
```

### Step 2: Open Swagger UI in Your Browser

Navigate to:

```
http://localhost:8080/swagger/index.html
```

**That's it!** You should see a beautiful documentation page listing all your API endpoints.

## 🎯 Using the Swagger UI

### Viewing Endpoints

The Swagger UI organizes endpoints by **tags**:
- **auth** - Authentication endpoints (login)
- **movies** - Movie management endpoints (CRUD operations)

Click on any endpoint to expand it and see:
- Description
- Required parameters
- Request body format
- Response examples
- Status codes

### Testing Endpoints (Try it out!)

This is the powerful part - you can test APIs without Postman:

#### 1. **Test Login (No Auth Required)**

1. Find the `POST /login` endpoint under **auth** tag
2. Click "Try it out"
3. You'll see an editable JSON request body
4. Modify the credentials (or use defaults: admin/password)
5. Click "Execute"
6. See the response below - you'll get a JWT token!

Example:
```json
{
  "username": "admin",
  "password": "password"
}
```

#### 2. **Test Protected Endpoints (Requires Auth)**

For endpoints like creating a movie, you need to authenticate first:

1. **Get a Token**: Use the login endpoint first (as shown above)
2. **Copy the Token**: From the response, copy the JWT token
3. **Click "Authorize"**: At the top of Swagger UI, click the green "Authorize" button
4. **Enter Token**: Paste `Bearer <your-token>` (with "Bearer " prefix)
   - Example: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
5. **Click "Authorize"** then **"Close"**

Now you can test any protected endpoint!

#### 3. **Create a Movie**

1. Find `POST /movies` under **movies** tag
2. Click "Try it out"
3. Edit the JSON request body (or use the example):
```json
{
  "title": "The Matrix",
  "description": "A computer hacker learns from mysterious rebels about the true nature of his reality",
  "director": "The Wachowskis",
  "release_date": "1999-03-31T00:00:00Z",
  "genres": ["Action", "Sci-Fi"],
  "rating": 8.7
}
```
4. Click "Execute"
5. See your newly created movie in the response!

#### 4. **List Movies with Filters**

1. Find `GET /movies` under **movies** tag
2. Click "Try it out"
3. Try different filters:
   - `title`: Search by movie title
   - `director`: Filter by director name
   - `genre`: Filter by genre
   - `page`: Page number for pagination
   - `limit`: Items per page
4. Click "Execute"
5. See filtered results!

## 🔄 Updating Documentation

When you **modify your API** (add new endpoints, change parameters, etc.), you need to regenerate the docs.

### Option 1: Use the Script

```bash
./scripts/generate-docs.sh
```

### Option 2: Manual Command

```bash
swag init -g cmd/server/main.go -o docs
```

**Important**: You must restart your server after regenerating docs for changes to take effect.

## 📝 How the Documentation Works

### 1. Annotations in Code

Documentation is generated from special comments in your code:

```go
// CreateMovie creates a new movie in the database
// @Summary      Create a new movie
// @Description  Create a new movie with the provided details
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

### 2. What Each Annotation Means

| Annotation | Purpose | Example |
|------------|---------|---------|
| `@Summary` | Short description | `Create a new movie` |
| `@Description` | Detailed explanation | `Creates a new movie with validation` |
| `@Tags` | Group endpoints | `movies` |
| `@Accept` | Input format | `json` |
| `@Produce` | Output format | `json` |
| `@Param` | Request parameters | `movie body models.Movie true` |
| `@Success` | Success response | `201 {object} models.Movie` |
| `@Failure` | Error response | `400 {object} map[string]string` |
| `@Security` | Auth requirement | `BearerAuth` |
| `@Router` | Endpoint path | `/movies [post]` |

### 3. Generated Files

When you run `swag init`, it creates:

```
docs/
├── docs.go         # Go code with embedded OpenAPI spec
├── swagger.json    # OpenAPI spec in JSON format
└── swagger.yaml    # OpenAPI spec in YAML format
```

**Never edit these files manually!** They're auto-generated.

## 🔐 Authentication in Swagger

Your API uses **JWT Bearer tokens**. Here's how it works in Swagger:

1. The green **"Authorize"** button appears because of this annotation in `main.go`:
```go
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
```

2. Protected endpoints have a 🔒 lock icon and this annotation:
```go
// @Security     BearerAuth
```

3. To test protected endpoints:
   - Get token from `/login`
   - Click "Authorize"
   - Enter: `Bearer <token>`
   - All subsequent requests will include this token

## 📊 Example Workflow

### Complete Testing Flow:

1. **Start Server**
   ```bash
   JWT_SECRET=test go run cmd/server/main.go
   ```

2. **Open Swagger UI**
   - Visit: http://localhost:8080/swagger/index.html

3. **Login**
   - Expand `POST /login`
   - Try it out with: `{"username": "admin", "password": "password"}`
   - Copy the returned token

4. **Authorize**
   - Click green "Authorize" button
   - Paste: `Bearer <your-token>`
   - Click "Authorize" then "Close"

5. **Create Movie**
   - Expand `POST /movies`
   - Try it out with sample movie data
   - Execute

6. **List Movies**
   - Expand `GET /movies`
   - Try different filters
   - Execute

7. **Update Movie**
   - Expand `PUT /movies/{id}`
   - Enter movie ID (e.g., 1)
   - Modify movie data
   - Execute

8. **Delete Movie**
   - Expand `DELETE /movies/{id}`
   - Enter movie ID
   - Execute

## 🛠️ Troubleshooting

### Swagger UI Not Loading

- **Check server is running**: `curl http://localhost:8080/swagger/index.html`
- **Check route is registered**: Look for `/swagger/*any` in console output
- **Clear browser cache**: Hard refresh with Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows)

### "Failed to Load API Definition"

- **Regenerate docs**: Run `./scripts/generate-docs.sh`
- **Restart server**: Stop and start your server again
- **Check for syntax errors**: Look at server console for errors

### "Authorization has been revoked"

- Your token may have expired (24 hours)
- Get a new token from `/login` and re-authorize

### Changes Not Showing

- **Regenerate docs** after code changes
- **Restart server** to load new docs
- **Hard refresh browser** to clear cache

## 📚 Learn More

- [Official Swagger Documentation](https://swagger.io/docs/)
- [swag GitHub Repository](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest.html)

## 🎓 Adding Documentation to New Endpoints

When you add a new endpoint, follow these steps:

### 1. Add Annotations

```go
// YourFunction does something
// @Summary      Short description
// @Description  Longer description
// @Tags         group-name
// @Accept       json
// @Produce      json
// @Param        name  type  DataType  required  "description"
// @Success      200   {object}  ResponseType
// @Security     BearerAuth  (if protected)
// @Router       /path [method]
func YourFunction(c *gin.Context) {
    // ... code
}
```

### 2. Regenerate Docs

```bash
./scripts/generate-docs.sh
```

### 3. Restart Server

```bash
# Stop current server (Ctrl+C)
# Start again
go run cmd/server/main.go
```

### 4. Test in Swagger

- Visit http://localhost:8080/swagger/index.html
- Find your new endpoint
- Try it out!

## 🎉 Benefits You Get

✅ **No More Postman Collections** - Test directly in browser
✅ **Self-Documenting API** - Docs update with code
✅ **Team Collaboration** - Share one URL for all docs
✅ **Client Generation** - Generate SDKs for any language
✅ **API Contract** - Clear contract between frontend/backend
✅ **Professional Look** - Impressive for stakeholders

Enjoy your interactive API documentation! 🚀
