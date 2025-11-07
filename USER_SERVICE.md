# User Service Documentation

## Overview

The User Service has been successfully implemented with registration, profile management, and role-based access control (RBAC) features.

## Features Implemented

### 1. User Model (`pkg/models/user.go`)
- Complete user entity with fields:
  - ID, Username, Email, Password (hashed)
  - Role (admin, moderator, user)
  - Profile fields: FirstName, LastName, Bio, Avatar
  - IsActive flag for account status
  - Timestamps: CreatedAt, UpdatedAt, DeletedAt (soft delete)
- UserProfile type for public representation (excludes password)
- Three role types: `admin`, `moderator`, `user`

### 2. Authentication Enhancements

#### Password Security (`internal/auth/password.go`)
- Bcrypt-based password hashing
- Secure password verification

#### JWT Improvements (`internal/auth/jwt.go`)
- Enhanced token generation with user ID and role
- Token includes: user_id, username, role, expiration
- Middleware automatically extracts claims to context

#### Login (`internal/auth/login.go`)
- Database-backed authentication
- Password verification using bcrypt
- Returns JWT token with user profile
- Account status checking (IsActive)

### 3. User Registration (`internal/users/register.go`)
- Public registration endpoint
- Input validation (username, email, password)
- Duplicate username/email checking
- Automatic user role assignment
- Password hashing
- Returns JWT token and user profile

### 4. Profile Management (`internal/users/profile.go`)

**Endpoints:**
- `GET /profile` - Get authenticated user's profile
- `PUT /profile` - Update authenticated user's profile
- `DELETE /profile` - Soft delete user account
- `GET /users/:id` - Get any user's public profile (no auth required)

**Update Fields:**
- FirstName, LastName, Bio, Avatar

### 5. Role-Based Authorization (`internal/middleware/rbac.go`)

**Middleware Functions:**
- `RequireRole(roles...)` - Check if user has one of specified roles
- `RequireAdmin()` - Convenience function for admin-only routes
- `RequireModerator()` - Convenience function for moderator/admin routes

**Features:**
- Fetches user from database to verify current role
- Checks account active status
- Stores user_id and user_role in context
- Returns appropriate error responses (401, 403)

### 6. Admin User Management (`internal/users/admin.go`)

**Endpoints (Admin Only):**
- `GET /admin/users` - List all users (paginated)
  - Query params: page, limit
- `PUT /admin/users/:id/role` - Update user role
- `POST /admin/users/:id/deactivate` - Deactivate user account
- `POST /admin/users/:id/activate` - Activate user account

### 7. Database Integration (`internal/storage/db.go`)
- Added User model to auto-migration
- User table created with all constraints
- Unique indexes on username and email

## API Endpoints

### Public Endpoints

```
POST /register          - Register new user
POST /login            - Login with credentials
GET  /users/:id        - Get user profile by ID
```

### Authenticated Endpoints

```
GET    /profile        - Get own profile
PUT    /profile        - Update own profile
DELETE /profile        - Delete own account
```

### Admin Endpoints (Requires Admin Role)

```
GET    /admin/users              - List all users
PUT    /admin/users/:id/role     - Update user role
POST   /admin/users/:id/activate - Activate user
POST   /admin/users/:id/deactivate - Deactivate user
```

## Usage Examples

### Register a New User

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "securepassword123"
  }'
```

### Get Own Profile

```bash
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer <your-jwt-token>"
```

### Update Profile

```bash
curl -X PUT http://localhost:8080/profile \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Smith",
    "bio": "Updated bio"
  }'
```

### Admin: List Users

```bash
curl -X GET "http://localhost:8080/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer <admin-jwt-token>"
```

### Admin: Update User Role

```bash
curl -X PUT http://localhost:8080/admin/users/2/role \
  -H "Authorization: Bearer <admin-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "moderator"
  }'
```

## User Roles

### User (Default)
- Basic authenticated access
- Can manage own profile
- Can access movie features

### Moderator
- All User permissions
- Can moderate content (if implemented)
- Can be used for future content moderation features

### Admin
- All permissions
- Can manage all users
- Can change user roles
- Can activate/deactivate accounts

## Security Features

1. **Password Security**
   - Bcrypt hashing with default cost
   - Passwords never exposed in API responses

2. **JWT Token Security**
   - 24-hour token expiration
   - Signed with secret key from environment
   - Contains user ID, username, and role

3. **Input Validation**
   - Username: 3-50 characters
   - Email: valid email format
   - Password: minimum 8 characters
   - Bio: maximum 500 characters

4. **Account Protection**
   - Soft delete support (DeletedAt)
   - Account deactivation (IsActive flag)
   - Inactive accounts cannot login

5. **Authorization**
   - Role-based access control
   - Admin-only routes protected
   - User context validation

## Files Created/Modified

### New Files:
- `pkg/models/user.go` - User model and types
- `internal/auth/password.go` - Password hashing utilities
- `internal/users/register.go` - Registration handler
- `internal/users/profile.go` - Profile management handlers
- `internal/users/admin.go` - Admin user management
- `internal/middleware/rbac.go` - Role-based authorization

### Modified Files:
- `internal/auth/jwt.go` - Enhanced JWT with user claims
- `internal/auth/login.go` - Database-backed authentication
- `internal/storage/db.go` - Added User model migration
- `cmd/server/main.go` - Added user routes

## Next Steps

1. **Create Initial Admin User:**
   - You'll need to manually create an admin user in the database or create a seed script
   - Example SQL:
   ```sql
   INSERT INTO users (username, email, password, role, is_active, created_at, updated_at)
   VALUES ('admin', 'admin@example.com', '<bcrypt-hash>', 'admin', true, datetime('now'), datetime('now'));
   ```

2. **Generate Swagger Documentation:**
   ```bash
   swag init -g cmd/server/main.go
   ```

3. **Test the API:**
   - Register a test user
   - Login and get JWT token
   - Test profile endpoints
   - Test admin endpoints with admin account

4. **Optional Enhancements:**
   - Email verification
   - Password reset functionality
   - Rate limiting on registration
   - OAuth integration
   - Two-factor authentication

## Configuration

Ensure the following environment variable is set:

```bash
export JWT_SECRET="your-secret-key-here"
```

## Testing

Run the server:
```bash
go run cmd/server/main.go
```

Access Swagger UI:
```
http://localhost:8080/swagger/index.html
```

## Build Status

✅ All files compile successfully
✅ Database migration configured
✅ All endpoints registered
✅ Role-based authorization implemented
