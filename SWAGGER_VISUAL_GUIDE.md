# Visual Guide: What You'll See in Swagger UI

## 🎨 Step-by-Step Visual Walkthrough

### Step 1: Open Swagger UI

When you visit **http://localhost:8080/swagger/index.html**, you'll see:

```
┌─────────────────────────────────────────────────────────────┐
│  Movie Database API                                    1.0  │
│  A movie database API with authentication, rate limiting... │
│                                                              │
│  Base URL: http://localhost:8080                           │
│                                              [Authorize 🔓] │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ▼ auth - Authentication endpoints                         │
│                                                              │
│    POST  /login                    User login              │
│                                                              │
│  ▼ movies - Movie management endpoints                     │
│                                                              │
│    GET   /movies                   List movies with filters│
│    POST  /movies            🔒     Create a new movie      │
│    POST  /movies/bulk       🔒     Create multiple movies  │
│    GET   /movies/{id}       🔒     Get movie by ID         │
│    PUT   /movies/{id}       🔒     Update movie            │
│    DELETE /movies/{id}      🔒     Delete movie            │
│    DELETE /movies           🔒     Delete multiple movies  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Legend:**
- 🔒 = Requires authentication
- ▼ = Click to expand/collapse group

---

### Step 2: Test Login Endpoint

Click on **POST /login**, you'll see:

```
┌─────────────────────────────────────────────────────────────┐
│  POST /login                                 [Try it out]  │
├─────────────────────────────────────────────────────────────┤
│  User login                                                 │
│  Authenticate with username and password to receive a JWT   │
│  token                                                       │
│                                                              │
│  Parameters                                                  │
│                                                              │
│  credentials * (body) - Login credentials                   │
│  ┌──────────────────────────────────────────────┐          │
│  │ {                                             │          │
│  │   "username": "admin",                        │          │
│  │   "password": "password"                      │          │
│  │ }                                             │          │
│  └──────────────────────────────────────────────┘          │
│                                                [Execute]     │
└─────────────────────────────────────────────────────────────┘
```

After clicking **Execute**, you'll see:

```
┌─────────────────────────────────────────────────────────────┐
│  Responses                                                   │
├─────────────────────────────────────────────────────────────┤
│  Code: 200  Success                                         │
│                                                              │
│  Response body:                                              │
│  ┌──────────────────────────────────────────────┐          │
│  │ {                                             │          │
│  │   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6..."│  [Copy]  │
│  │ }                                             │          │
│  └──────────────────────────────────────────────┘          │
│                                                              │
│  Response headers:                                           │
│  content-type: application/json                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### Step 3: Authorize with Token

Click the green **Authorize** button at the top:

```
┌─────────────────────────────────────────────────────────────┐
│  Available authorizations                                    │
├─────────────────────────────────────────────────────────────┤
│  BearerAuth (apiKey)                                        │
│                                                              │
│  Name: Authorization                                         │
│  In: header                                                  │
│                                                              │
│  Value: *                                                    │
│  ┌──────────────────────────────────────────────┐          │
│  │ Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6...      │          │
│  └──────────────────────────────────────────────┘          │
│                                                              │
│                           [Authorize]  [Close]              │
└─────────────────────────────────────────────────────────────┘
```

**Important**: Don't forget to add "Bearer " before the token!

After authorizing, the button changes:

```
[Authorize 🔓]  →  [Authorize 🔐 Logout]
```

Now all 🔒 endpoints can be accessed!

---

### Step 4: Create a Movie

Click on **POST /movies** (now unlocked):

```
┌─────────────────────────────────────────────────────────────┐
│  POST /movies                                [Try it out]  │
├─────────────────────────────────────────────────────────────┤
│  Create a new movie                                    🔐   │
│  Create a new movie with the provided details. Requires     │
│  authentication.                                             │
│                                                              │
│  Parameters                                                  │
│                                                              │
│  movie * (body) - Movie object to create                    │
│  ┌──────────────────────────────────────────────┐          │
│  │ {                                             │          │
│  │   "title": "The Matrix",                      │          │
│  │   "description": "A computer hacker learns..",│          │
│  │   "director": "The Wachowskis",               │          │
│  │   "release_date": "1999-03-31T00:00:00Z",     │          │
│  │   "genres": ["Action", "Sci-Fi"],             │          │
│  │   "rating": 8.7                               │          │
│  │ }                                             │          │
│  └──────────────────────────────────────────────┘          │
│                                                              │
│  Example Value │ Model    ▼                                 │
│                                                [Execute]     │
└─────────────────────────────────────────────────────────────┘
```

**Notice:**
- **Example Value** tab: Shows example JSON
- **Model** tab: Shows field types and constraints
- Fields are **pre-filled** with example data!

---

### Step 5: View Model Schema

Click the **Model** tab:

```
┌─────────────────────────────────────────────────────────────┐
│  Movie {                                                     │
│    id           integer($uint)      Example: 1              │
│    title        string              Example: The Matrix     │
│    description  string              Example: A computer...  │
│    director     string              Example: The Wachowskis │
│    release_date string($date-time)  Example: 1999-03-31...  │
│    genres       [string]            Example: Action,Sci-Fi  │
│    rating       number($float)      Example: 8.7            │
│                                     Minimum: 0              │
│                                     Maximum: 10             │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

**This shows:**
- Field names and types
- Example values for each field
- Validation rules (min/max for rating)

---

### Step 6: Test GET with Filters

Click on **GET /movies**:

```
┌─────────────────────────────────────────────────────────────┐
│  GET /movies                                 [Try it out]  │
├─────────────────────────────────────────────────────────────┤
│  List movies with filters                                   │
│  Get a paginated list of movies with optional search        │
│  filters. Public endpoint (no authentication required).     │
│                                                              │
│  Parameters                                                  │
│                                                              │
│  page (query)      integer  Page number         [  1  ]    │
│  limit (query)     integer  Items per page      [ 10  ]    │
│  title (query)     string   Filter by title     [     ]    │
│  director (query)  string   Filter by director  [     ]    │
│  genre (query)     string   Filter by genre     [     ]    │
│                                                              │
│                                                [Execute]     │
└─────────────────────────────────────────────────────────────┘
```

**Try entering:**
- title: "Matrix"
- page: 1
- limit: 5

Then click **Execute**!

---

### Step 7: View Paginated Response

```
┌─────────────────────────────────────────────────────────────┐
│  Responses                                                   │
├─────────────────────────────────────────────────────────────┤
│  Code: 200  Success                                         │
│                                                              │
│  Response body:                                              │
│  ┌──────────────────────────────────────────────┐          │
│  │ {                                             │          │
│  │   "data": [                                   │          │
│  │     {                                         │          │
│  │       "id": 1,                                │          │
│  │       "title": "The Matrix",                  │          │
│  │       "description": "...",                   │          │
│  │       "director": "The Wachowskis",           │          │
│  │       "release_date": "1999-03-31T00:00:00Z", │          │
│  │       "genres": ["Action", "Sci-Fi"],         │          │
│  │       "rating": 8.7                           │          │
│  │     }                                         │          │
│  │   ],                                          │          │
│  │   "pagination": {                             │          │
│  │     "page": 1,                                │          │
│  │     "page_size": 10,                          │          │
│  │     "total": 1,                               │          │
│  │     "total_pages": 1,                         │          │
│  │     "has_next": false,                        │          │
│  │     "has_prev": false                         │          │
│  │   }                                           │          │
│  │ }                                             │          │
│  └──────────────────────────────────────────────┘          │
│                                                              │
│  Curl command:                          [Copy]              │
│  curl -X GET "http://localhost:8080/movies?page=1&limit=10" │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Bonus**: Swagger shows the equivalent **curl command** you could run!

---

## 🎯 Key Features Highlighted

### 1. Color Coding
- **Green** = GET requests
- **Blue** = POST requests
- **Orange** = PUT requests
- **Red** = DELETE requests

### 2. Security Indicators
- 🔓 = Open lock (not authenticated)
- 🔐 = Closed lock (authenticated)
- 🔒 = Lock icon next to endpoint (requires auth)

### 3. Interactive Elements
- **Try it out** button = Enable editing
- **Execute** button = Send request
- **Example Value** tab = See sample JSON
- **Model** tab = See structure and types
- **Copy** button = Copy response/curl command

### 4. Response Information
- HTTP status code
- Response body (with syntax highlighting)
- Response headers
- Curl command equivalent

---

## 💡 Pro Tips for Navigation

### Keyboard Shortcuts
- **Ctrl/Cmd + F** = Search for endpoints
- **Click** endpoint name = Expand/collapse
- **Tab** = Navigate between fields

### Quick Actions
- **Double-click** JSON = Select all
- **Right-click** response = Copy
- **Click** "Authorize" = Opens auth modal
- **Click** lock icon = Shows which auth is needed

### Understanding Symbols
- `*` = Required parameter
- `(body)` = In request body
- `(query)` = In URL query string
- `(path)` = In URL path (like `/movies/{id}`)
- `(header)` = In HTTP headers

---

## 🎨 What Makes This Special

Your Swagger UI includes:

✅ **Complete endpoint coverage** - All 9 endpoints documented
✅ **Example values** - Every field has realistic examples
✅ **Authentication flow** - JWT auth fully integrated
✅ **Validation rules** - Min/max shown for rating
✅ **Request/response models** - Clear structure definitions
✅ **Error responses** - All possible errors documented
✅ **Filters and pagination** - Query parameters clearly shown
✅ **Professional presentation** - Clean, organized, easy to use

---

## 🚀 Getting Started Checklist

- [ ] Start server with JWT_SECRET set
- [ ] Open http://localhost:8080/swagger/index.html
- [ ] Test POST /login to get token
- [ ] Click "Authorize" and paste "Bearer <token>"
- [ ] Try POST /movies to create a movie
- [ ] Try GET /movies with different filters
- [ ] Try GET /movies/{id} to get specific movie
- [ ] Try PUT /movies/{id} to update
- [ ] Try DELETE /movies/{id} to delete

---

Enjoy exploring your API visually! 🎊
