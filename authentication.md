# 🎓 Learning Platform - Documentation

## 📋 Tổng quan hệ thống

Nền tảng học tập trực tuyến với các tính năng:
- ✅ Quản lý nội dung học tập (Lessons, Quizzes)
- ✅ AI-powered auto-grading
- ✅ Gamification (Points, Streaks, Leaderboard)
- ✅ Multi-modal learning (Text, Video, Audio)
- ✅ Spaced Repetition System
- ✅ Real-time progress tracking

---

## 🏗️ Kiến trúc hệ thống

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                          │
│  Web (React) / Mobile (React Native) / Admin Dashboard  │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   API Gateway                            │
│  Authentication / Rate Limiting / Request Routing       │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                Backend Services                          │
│  • Auth Service                                          │
│  • Content Service                                       │
│  • Learning Service                                      │
│  • Assessment Service                                    │
│  • Gamification Service                                  │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              Database (PostgreSQL)                       │
│  • User & Profile Management                            │
│  • Content Management                                    │
│  • Progress Tracking                                     │
│  • Assessment & Scoring                                  │
└─────────────────────────────────────────────────────────┘
```

---

## 👥 Hệ thống Roles & Permissions

### **1. Roles Definition**

```sql
-- Bảng roles (cần thêm vào schema)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Bảng user_roles (many-to-many)
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Bảng permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource VARCHAR(100) NOT NULL,  -- lessons, quizzes, users, etc.
    action VARCHAR(50) NOT NULL,     -- create, read, update, delete
    description TEXT,
    UNIQUE(resource, action)
);

-- Bảng role_permissions
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
```

### **2. Roles Hierarchy**

| Role | Level | Description |
|------|-------|-------------|
| **SUPER_ADMIN** | 5 | Toàn quyền hệ thống |
| **ADMIN** | 4 | Quản lý nội dung, users, reports |
| **TEACHER** | 3 | Tạo/sửa lessons, chấm bài, xem reports |
| **CONTENT_CREATOR** | 2 | Tạo lessons, media (không chấm bài) |
| **STUDENT** | 1 | Học bài, làm quiz, xem tiến độ |
| **GUEST** | 0 | Chỉ xem preview (chưa login) |

### **3. Permission Matrix**

| Resource | SUPER_ADMIN | ADMIN | TEACHER | CONTENT_CREATOR | STUDENT |
|----------|-------------|-------|---------|-----------------|---------|
| **Users** |
| Create User | ✅ | ✅ | ❌ | ❌ | ❌ |
| View Users | ✅ | ✅ | ✅ (own students) | ❌ | ❌ |
| Edit User | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete User | ✅ | ✅ | ❌ | ❌ | ❌ |
| Assign Roles | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Lessons** |
| Create Lesson | ✅ | ✅ | ✅ | ✅ | ❌ |
| View Lessons | ✅ | ✅ | ✅ | ✅ | ✅ (enrolled) |
| Edit Lesson | ✅ | ✅ | ✅ (own) | ✅ (own) | ❌ |
| Delete Lesson | ✅ | ✅ | ✅ (own) | ❌ | ❌ |
| Publish Lesson | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Quizzes** |
| Create Quiz | ✅ | ✅ | ✅ | ✅ | ❌ |
| View Quiz | ✅ | ✅ | ✅ | ✅ | ✅ (enrolled) |
| Edit Quiz | ✅ | ✅ | ✅ (own) | ✅ (own) | ❌ |
| View Answers | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Submissions** |
| Submit Answer | ✅ | ✅ | ✅ | ✅ | ✅ |
| View Own Submissions | ✅ | ✅ | ✅ | ✅ | ✅ |
| View All Submissions | ✅ | ✅ | ✅ (own lessons) | ❌ | ❌ |
| Grade Submissions | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Reports** |
| View Analytics | ✅ | ✅ | ✅ (own classes) | ❌ | ❌ |
| Export Data | ✅ | ✅ | ✅ | ❌ | ❌ |
| View Leaderboard | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 🔐 Authentication Flow

### **1. Registration Flow**

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     │ POST /api/auth/register
     │ {email, password, display_name}
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Validate input
     │    - Email format
     │    - Password strength (min 8 chars, 1 uppercase, 1 number)
     │    - Check email exists
     │
     │ 2. Hash password (bcrypt, rounds=12)
     │
     │ 3. BEGIN TRANSACTION
     │    INSERT INTO users (email, password_hash, status='pending')
     │    INSERT INTO user_profiles (display_name, locale)
     │    INSERT INTO user_roles (role='STUDENT')
     │    INSERT INTO user_streaks (current_len=0)
     │    INSERT INTO user_points (lifetime=0)
     │    COMMIT
     │
     │ 4. Generate verification token (JWT, 24h expiry)
     │    payload: {user_id, email, type: 'email_verification'}
     │
     │ 5. Send verification email
     │    Link: https://app.com/verify?token=xxx
     │
     ▼
┌─────────────────┐
│  Response       │
│  {              │
│    success: true│
│    message: ""  │
│  }              │
└─────────────────┘
```

### **2. Email Verification Flow**

```
User clicks verification link
     │
     │ GET /api/auth/verify?token=xxx
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Verify JWT token
     │    - Check signature
     │    - Check expiry
     │    - Check type='email_verification'
     │
     │ 2. UPDATE users
     │    SET email_verified = true,
     │        status = 'active',
     │        updated_at = NOW()
     │    WHERE id = token.user_id
     │
     │ 3. Send welcome email
     │
     ▼
Redirect to /login?verified=true
```

### **3. Login Flow (Standard)**

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     │ POST /api/auth/login
     │ {email, password, remember_me}
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Query user by email
     │    SELECT * FROM users WHERE email = ?
     │
     │ 2. Check user exists & status='active'
     │
     │ 3. Verify password
     │    bcrypt.compare(password, user.password_hash)
     │
     │ 4. Check email_verified = true
     │
     │ 5. Get user roles
     │    SELECT r.name FROM roles r
     │    JOIN user_roles ur ON r.id = ur.role_id
     │    WHERE ur.user_id = ?
     │
     │ 6. Get user profile
     │    SELECT * FROM user_profiles WHERE user_id = ?
     │
     │ 7. Generate tokens
     │    ACCESS_TOKEN (JWT, 15min expiry)
     │    {
     │      user_id: uuid,
     │      email: string,
     │      roles: ['STUDENT'],
     │      type: 'access'
     │    }
     │
     │    REFRESH_TOKEN (JWT, 7 days or 30 days if remember_me)
     │    {
     │      user_id: uuid,
     │      type: 'refresh',
     │      jti: unique_id  // for revocation
     │    }
     │
     │ 8. Store refresh token in DB
     │    INSERT INTO refresh_tokens
     │    (user_id, token_jti, expires_at, ip, user_agent)
     │
     │ 9. Update last_login
     │    UPDATE users SET updated_at = NOW()
     │
     ▼
┌─────────────────┐
│  Response       │
│  {              │
│    access_token │
│    refresh_token│
│    user: {      │
│      id,        │
│      email,     │
│      roles,     │
│      profile    │
│    }            │
│  }              │
└─────────────────┘
     │
     ▼
Client stores:
- access_token in memory (or sessionStorage)
- refresh_token in httpOnly cookie (secure, sameSite)
```

### **4. Token Refresh Flow**

```
┌─────────┐
│ Client  │  Access token hết hạn (401 Unauthorized)
└────┬────┘
     │
     │ POST /api/auth/refresh
     │ Cookie: refresh_token=xxx
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Verify refresh token
     │    - Check signature
     │    - Check expiry
     │    - Check type='refresh'
     │
     │ 2. Check token in DB (not revoked)
     │    SELECT * FROM refresh_tokens
     │    WHERE token_jti = ? AND revoked_at IS NULL
     │
     │ 3. Get user & roles (như login)
     │
     │ 4. Generate NEW access token (15min)
     │
     │ 5. (Optional) Rotate refresh token
     │    - Revoke old refresh token
     │    - Generate new refresh token
     │
     ▼
┌─────────────────┐
│  Response       │
│  {              │
│    access_token │
│    refresh_token│  (if rotated)
│  }              │
└─────────────────┘
```

### **5. OAuth2 Login Flow (Google/Facebook)**

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     │ GET /api/auth/oauth/google
     ▼
┌─────────────────┐
│  Auth Service   │  Redirect to Google
└────┬────────────┘
     │
     ▼
┌─────────────────┐
│  Google OAuth   │  User authorizes
└────┬────────────┘
     │
     │ Callback: /api/auth/oauth/google/callback?code=xxx
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Exchange code for tokens
     │    POST https://oauth2.googleapis.com/token
     │
     │ 2. Get user info
     │    GET https://www.googleapis.com/oauth2/v2/userinfo
     │    {email, name, picture, email_verified}
     │
     │ 3. Check if user exists
     │    SELECT * FROM users WHERE email = ?
     │
     │ 4a. If exists:
     │     - Update profile (picture, name if changed)
     │     - Generate tokens (như login flow)
     │
     │ 4b. If NOT exists:
     │     - CREATE user (email_verified=true, status='active')
     │     - CREATE profile (from OAuth data)
     │     - Assign STUDENT role
     │     - Generate tokens
     │
     ▼
Redirect to frontend with tokens
```

### **6. Logout Flow**

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     │ POST /api/auth/logout
     │ Headers: Authorization: Bearer {access_token}
     │ Cookie: refresh_token=xxx
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Verify access token (get user_id)
     │
     │ 2. Revoke refresh token
     │    UPDATE refresh_tokens
     │    SET revoked_at = NOW()
     │    WHERE user_id = ? AND token_jti = ?
     │
     │ 3. (Optional) Blacklist access token
     │    INSERT INTO token_blacklist (jti, expires_at)
     │    (only if access token has jti claim)
     │
     ▼
┌─────────────────┐
│  Response       │
│  {              │
│    success: true│
│  }              │
└─────────────────┘
     │
     ▼
Client clears:
- access_token from memory
- refresh_token cookie
```

### **7. Password Reset Flow**

```
Step 1: Request Reset
─────────────────────
User forgets password
     │
     │ POST /api/auth/forgot-password
     │ {email}
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Check email exists
     │    (Always return success to prevent email enumeration)
     │
     │ 2. Generate reset token (JWT, 1h expiry)
     │    payload: {user_id, email, type: 'password_reset'}
     │
     │ 3. Store token hash in DB
     │    INSERT INTO password_reset_tokens
     │    (user_id, token_hash, expires_at)
     │
     │ 4. Send reset email
     │    Link: https://app.com/reset-password?token=xxx
     │
     ▼
Response: {success: true, message: "Check your email"}


Step 2: Reset Password
──────────────────────
User clicks reset link
     │
     │ POST /api/auth/reset-password
     │ {token, new_password}
     ▼
┌─────────────────┐
│  Auth Service   │
└────┬────────────┘
     │
     │ 1. Verify JWT token (signature, expiry, type)
     │
     │ 2. Check token in DB (not used)
     │    SELECT * FROM password_reset_tokens
     │    WHERE token_hash = HASH(token)
     │      AND used_at IS NULL
     │      AND expires_at > NOW()
     │
     │ 3. Validate new password
     │
     │ 4. BEGIN TRANSACTION
     │    - Hash new password
     │    - UPDATE users SET password_hash = ?
     │    - UPDATE password_reset_tokens SET used_at = NOW()
     │    - Revoke ALL refresh tokens (force re-login)
     │    COMMIT
     │
     │ 5. Send confirmation email
     │
     ▼
Response: {success: true}
```

---

## 🔒 Security Best Practices

### **1. Password Security**
```javascript
// Hashing
const bcrypt = require('bcrypt');
const SALT_ROUNDS = 12;
const hash = await bcrypt.hash(password, SALT_ROUNDS);

// Password requirements
- Min 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character (!@#$%^&*)
- Not in common password list (check against top 10k)
```

### **2. JWT Security**
```javascript
// Access Token (short-lived)
{
  "iss": "learning-platform",
  "sub": "user_id",
  "iat": 1234567890,
  "exp": 1234568790,  // 15 minutes
  "type": "access",
  "roles": ["STUDENT"]
}

// Refresh Token (long-lived)
{
  "iss": "learning-platform",
  "sub": "user_id",
  "iat": 1234567890,
  "exp": 1235172690,  // 7 days
  "type": "refresh",
  "jti": "unique-token-id"  // for revocation
}

// Secret keys (env variables)
ACCESS_TOKEN_SECRET=<strong-random-key-256bit>
REFRESH_TOKEN_SECRET=<different-strong-random-key>
```

### **3. Rate Limiting**
```javascript
// Login endpoint
- 5 attempts per 15 minutes per IP
- 10 attempts per hour per email
- Exponential backoff after failed attempts

// Registration endpoint
- 3 registrations per hour per IP
- Email verification required

// Password reset
- 3 requests per hour per email
- 10 requests per hour per IP
```

### **4. CSRF Protection**
```javascript
// Use SameSite cookies
Set-Cookie: refresh_token=xxx; 
  HttpOnly; 
  Secure; 
  SameSite=Strict; 
  Path=/api/auth/refresh

// CSRF tokens for state-changing operations
X-CSRF-Token: <token>
```

### **5. SQL Injection Prevention**
```javascript
// Always use parameterized queries
// ❌ BAD
const query = `SELECT * FROM users WHERE email = '${email}'`;

// ✅ GOOD (SQLAlchemy ORM)
const user = await User.findOne({where: {email}});

// ✅ GOOD (Raw query with params)
const user = await db.query(
  'SELECT * FROM users WHERE email = $1', 
  [email]
);
```

---

## 📊 Database Schema for Auth

```sql
-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_jti VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_jti ON refresh_tokens(token_jti);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- Password reset tokens
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- OAuth providers
CREATE TABLE oauth_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,  -- google, facebook, github
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- Audit log
CREATE TABLE auth_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,  -- login, logout, register, etc.
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_auth_logs_user_id ON auth_logs(user_id);
CREATE INDEX idx_auth_logs_created_at ON auth_logs(created_at);
```

---

## 🚀 API Endpoints Summary

### **Auth Endpoints**
```
POST   /api/auth/register              # Đăng ký
GET    /api/auth/verify                # Xác thực email
POST   /api/auth/login                 # Đăng nhập
POST   /api/auth/refresh               # Refresh token
POST   /api/auth/logout                # Đăng xuất
POST   /api/auth/forgot-password       # Quên mật khẩu
POST   /api/auth/reset-password        # Đặt lại mật khẩu
GET    /api/auth/oauth/:provider       # OAuth login
GET    /api/auth/oauth/:provider/callback  # OAuth callback
```

### **User Management (Admin)**
```
GET    /api/admin/users                # List users (paginated)
GET    /api/admin/users/:id            # Get user detail
PUT    /api/admin/users/:id            # Update user
DELETE /api/admin/users/:id            # Delete user
POST   /api/admin/users/:id/roles      # Assign roles
DELETE /api/admin/users/:id/roles/:role_id  # Remove role
GET    /api/admin/users/:id/activity   # User activity log
```

### **Role Management (Super Admin)**
```
GET    /api/admin/roles                # List all roles
POST   /api/admin/roles                # Create role
PUT    /api/admin/roles/:id            # Update role
DELETE /api/admin/roles/:id            # Delete role
GET    /api/admin/permissions          # List all permissions
POST   /api/admin/roles/:id/permissions  # Assign permission
```

---

## 🧪 Testing Authentication

```javascript
// Example test cases
describe('Authentication Flow', () => {
  test('Register new user', async () => {
    const res = await request(app)
      .post('/api/auth/register')
      .send({
        email: 'test@example.com',
        password: 'SecurePass123!',
        display_name: 'Test User'
      });
    
    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
  });

  test('Login with valid credentials', async () => {
    const res = await request(app)
      .post('/api/auth/login')
      .send({
        email: 'test@example.com',
        password: 'SecurePass123!'
      });
    
    expect(res.status).toBe(200);
    expect(res.body.access_token).toBeDefined();
    expect(res.body.refresh_token).toBeDefined();
  });

  test('Access protected route with token', async () => {
    const token = 'valid_jwt_token';
    const res = await request(app)
      .get('/api/user/profile')
      .set('Authorization', `Bearer ${token}`);
    
    expect(res.status).toBe(200);
  });

  test('Reject invalid token', async () => {
    const res = await request(app)
      .get('/api/user/profile')
      .set('Authorization', 'Bearer invalid_token');
    
    expect(res.status).toBe(401);
  });
});
```

---

## 📝 Environment Variables

```bash
# Server
NODE_ENV=production
PORT=3000
API_VERSION=v1

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/learning_db
DB_POOL_MIN=2
DB_POOL_MAX=10

# JWT
ACCESS_TOKEN_SECRET=your-256-bit-secret-key-here
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_SECRET=your-different-256-bit-secret
REFRESH_TOKEN_EXPIRY=7d
REFRESH_TOKEN_EXPIRY_REMEMBER=30d

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@yourapp.com
SMTP_PASS=your-smtp-password
EMAIL_FROM=Learning Platform <noreply@yourapp.com>

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-secret
GOOGLE_CALLBACK_URL=https://api.yourapp.com/auth/oauth/google/callback

FACEBOOK_APP_ID=your-facebook-app-id
FACEBOOK_APP_SECRET=your-facebook-secret
FACEBOOK_CALLBACK_URL=https://api.yourapp.com/auth/oauth/facebook/callback

# Frontend URL
FRONTEND_URL=https://yourapp.com

# Rate Limiting
RATE_LIMIT_WINDOW=15m
RATE_LIMIT_MAX_REQUESTS=100

# Security
BCRYPT_ROUNDS=12
PASSWORD_MIN_LENGTH=8
```

---

## 📚 Additional Resources

- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [PostgreSQL Security](https://www.postgresql.org/docs/current/auth-methods.html)

---

## 🤝 Contributing

Xem [CONTRIBUTING.md](CONTRIBUTING.md) để biết chi tiết về quy trình development.

## 📄 License

MIT License - xem [LICENSE](LICENSE) file.