# Frontend Integration Guide

This guide provides everything frontend developers need to integrate with the English Learning App microservice backend.

## 🏗️ Architecture Overview

The backend consists of 5 microservices orchestrated through an API Gateway (Traefik) and a Backend-for-Frontend (BFF) service:

```
Frontend (React/Vue/Angular)
         ↓
API Gateway (Traefik) - Port 80
         ↓
BFF Service (Go) - Port 8010
         ↓
┌─────────────────────────────────────────┐
│  Microservices                         │
│  • User Service (Go) - Port 8001        │
│  • Content Service (Go/GraphQL) - 8003 │
│  • Lesson Service (Python/FastAPI) - 8005│
│  • Notification Service (Node.js) - 8004│
└─────────────────────────────────────────┘
```

## 🚀 Quick Start

### 1. Start the Backend Services

```bash
# Navigate to infrastructure directory
cd infrastructure

# Start all services with Docker Compose
docker-compose up -d

# Check service health
curl http://localhost/health
```

### 2. Service URLs

| Service | URL | Purpose |
|---------|-----|---------|
| **API Gateway** | `http://localhost` | Main entry point |
| **BFF Service** | `http://localhost/api/bff` | Aggregated API |
| **User Service** | `http://localhost/api/user` | Authentication |
| **Content Service** | `http://localhost/api/content` | Content management |
| **Lesson Service** | `http://localhost/api/lesson` | Progress tracking |
| **Notification Service** | `http://localhost/api/notification` | Email notifications |

## 🔐 Authentication

### JWT Token Flow

1. **Register/Login** → Get JWT tokens
2. **Include token** in all authenticated requests
3. **Token expires** → Use refresh token or re-login

### Authentication Headers

```javascript
// Include in all authenticated requests
const headers = {
  'Authorization': `Bearer ${accessToken}`,
  'Content-Type': 'application/json'
};
```

### User Context Headers

After successful authentication, the backend provides these headers:
- `X-User-ID`: User's unique identifier
- `X-User-Email`: User's email address  
- `X-Session-ID`: Current session identifier

## 📡 API Endpoints

### 🔑 Authentication (User Service)

**Base URL:** `http://localhost/api/user/api/v1`

#### Register User
```http
POST /api/user/api/v1/register
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Registration successful! Please check your email to verify your account.",
    "email": "user@example.com"
  }
}
```

#### Login
```http
POST /api/user/api/v1/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "mfa_code": "123456"  // Optional if MFA enabled
}
```

**Response:**
```json
{
  "status": "success,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-01T12:00:00Z",
    "mfa_required": false,
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "email_verified": true,
      "status": "active",
      "profile": {
        "display_name": "John Doe",
        "avatar_url": "https://...",
        "locale": "en",
        "time_zone": "UTC"
      }
    }
  }
}
```

#### Get User Profile
```http
GET /api/user/api/v1/profile
Authorization: Bearer <access_token>
```

#### Update Profile
```http
PUT /api/user/api/v1/profile
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "display_name": "John Smith",
  "avatar_url": "https://example.com/avatar.jpg",
  "locale": "en",
  "time_zone": "America/New_York"
}
```

#### Password Management
```http
# Request password reset
POST /api/user/api/v1/password/reset/request
{
  "email": "user@example.com"
}

# Confirm password reset
POST /api/user/api/v1/password/reset/confirm
{
  "token": "reset-token",
  "new_password": "NewSecurePass123!"
}

# Change password (authenticated)
POST /api/user/api/v1/password/change
Authorization: Bearer <access_token>
{
  "old_password": "OldPass123!",
  "new_password": "NewSecurePass123!"
}
```

#### Multi-Factor Authentication (MFA)
```http
# Setup MFA
POST /api/user/api/v1/mfa/setup
Authorization: Bearer <access_token>
{
  "type": "totp",
  "label": "My Authenticator"
}

# Verify MFA
POST /api/user/api/v1/mfa/verify
Authorization: Bearer <access_token>
{
  "method_id": "uuid",
  "code": "123456"
}

# Get MFA methods
GET /api/user/api/v1/mfa/methods
Authorization: Bearer <access_token>
```

### 📚 Content Management (Content Service)

**Base URL:** `http://localhost/api/content`

#### GraphQL Endpoint
```http
POST /api/content/graphql
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "query": "query { topics { id slug name } }"
}
```

#### Key GraphQL Queries

**Get Topics and Levels:**
```graphql
query {
  topics {
    id
    slug
    name
    createdAt
  }
  levels {
    id
    code
    name
  }
  tags {
    id
    slug
    name
  }
}
```

**Get Lessons:**
```graphql
query {
  lessons(
    filter: { 
      isPublished: true,
      topicId: "uuid",
      levelId: "uuid"
    }
    page: 1
    pageSize: 10
  ) {
    items {
      id
      code
      title
      description
      topic {
        id
        name
      }
      level {
        id
        name
      }
      isPublished
      createdAt
      sections {
        id
        type
        body
      }
    }
    totalCount
  }
}
```

**Create Lesson:**
```graphql
mutation {
  createLesson(input: {
    title: "Basic Grammar"
    description: "Learn English grammar basics"
    topicId: "uuid"
    levelId: "uuid"
    createdBy: "user-uuid"
  }) {
    id
    title
    code
  }
}
```

**Get Flashcards:**
```graphql
query {
  flashcardSets(
    topicId: "uuid"
    levelId: "uuid"
    page: 1
    pageSize: 10
  ) {
    items {
      id
      title
      description
      cards {
        id
        frontText
        backText
        hints
      }
    }
    totalCount
  }
}
```

**Get Quizzes:**
```graphql
query {
  quizzes(
    lessonId: "uuid"
    page: 1
    pageSize: 10
  ) {
    items {
      id
      title
    }
  }
}
```

### 🎯 Progress Tracking (Lesson Service)

**Base URL:** `http://localhost/api/lesson/api/v1`

#### Start a Lesson
```http
POST /api/lesson/api/v1/progress/lessons/start
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "user_id": "uuid",
  "lesson_id": "uuid"
}
```

#### Update Lesson Progress
```http
PUT /api/lesson/api/v1/progress/lessons/{user_id}/{lesson_id}/progress
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "section_id": "uuid",
  "completed": true,
  "score": 85
}
```

#### Get User's Lessons
```http
GET /api/lesson/api/v1/progress/lessons/user/{user_id}
Authorization: Bearer <access_token>
```

#### Quiz Attempts
```http
# Start quiz attempt
POST /api/lesson/api/v1/progress/quiz/attempts
Authorization: Bearer <access_token>
{
  "user_id": "uuid",
  "quiz_id": "uuid"
}

# Submit quiz answers
POST /api/lesson/api/v1/progress/quiz/attempts/{attempt_id}/submit
Authorization: Bearer <access_token>
{
  "answers": [
    {
      "question_id": "uuid",
      "answer": "selected_option_id",
      "is_correct": true
    }
  ]
}
```

#### Spaced Repetition (Flashcards)
```http
# Get due cards
GET /api/lesson/api/v1/progress/sr/cards/due/{user_id}
Authorization: Bearer <access_token>

# Submit review
POST /api/lesson/api/v1/progress/sr/reviews
Authorization: Bearer <access_token>
{
  "card_id": "uuid",
  "difficulty": 3,  // 1-5 scale
  "response_time_ms": 5000
}
```

#### Gamification & Leaderboards
```http
# Get user stats
GET /api/lesson/api/v1/progress/stats/{user_id}
Authorization: Bearer <access_token>

# Get leaderboard
GET /api/lesson/api/v1/progress/leaderboard
Authorization: Bearer <access_token>
```

### 📧 Notifications (Notification Service)

**Base URL:** `http://localhost/api/notification`

#### Send Email
```http
POST /api/notification/send-email
Content-Type: application/json

{
  "to": "user@example.com",
  "subject": "Welcome to English Learning!",
  "html": "<h1>Welcome!</h1><p>Start your learning journey.</p>",
  "text": "Welcome! Start your learning journey."
}
```

## 🔧 Frontend Integration Examples

### JavaScript/TypeScript Integration

```typescript
// API Client Configuration
class EnglishLearningAPI {
  private baseURL = 'http://localhost';
  private accessToken: string | null = null;

  constructor() {
    this.accessToken = localStorage.getItem('access_token');
  }

  private async request(endpoint: string, options: RequestInit = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const headers = {
      'Content-Type': 'application/json',
      ...(this.accessToken && { 'Authorization': `Bearer ${this.accessToken}` }),
      ...options.headers,
    };

    const response = await fetch(url, { ...options, headers });
    
    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    return response.json();
  }

  // Authentication
  async login(email: string, password: string) {
    const response = await this.request('/api/user/api/v1/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });

    if (response.status === 'success') {
      this.accessToken = response.data.access_token;
      localStorage.setItem('access_token', this.accessToken);
    }

    return response;
  }

  async logout() {
    await this.request('/api/user/api/v1/logout', { method: 'POST' });
    this.accessToken = null;
    localStorage.removeItem('access_token');
  }

  // Content Management
  async getLessons(filters: any = {}) {
    const query = `
      query GetLessons($filter: LessonFilterInput, $page: Int, $pageSize: Int) {
        lessons(filter: $filter, page: $page, pageSize: $pageSize) {
          items {
            id
            title
            description
            topic { id name }
            level { id name }
            isPublished
          }
          totalCount
        }
      }
    `;

    return this.request('/api/content/graphql', {
      method: 'POST',
      body: JSON.stringify({
        query,
        variables: { filter: filters, page: 1, pageSize: 10 }
      }),
    });
  }

  // Progress Tracking
  async startLesson(userId: string, lessonId: string) {
    return this.request('/api/lesson/api/v1/progress/lessons/start', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, lesson_id: lessonId }),
    });
  }

  async updateProgress(userId: string, lessonId: string, progress: any) {
    return this.request(`/api/lesson/api/v1/progress/lessons/${userId}/${lessonId}/progress`, {
      method: 'PUT',
      body: JSON.stringify(progress),
    });
  }
}

// Usage
const api = new EnglishLearningAPI();

// Login
const loginResponse = await api.login('user@example.com', 'password');

// Get lessons
const lessons = await api.getLessons({ isPublished: true });

// Start a lesson
await api.startLesson('user-id', 'lesson-id');
```

### React Hook Example

```typescript
// useAuth.ts
import { useState, useEffect } from 'react';

export const useAuth = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      // Verify token and get user data
      fetchUserProfile();
    } else {
      setLoading(false);
    }
  }, []);

  const fetchUserProfile = async () => {
    try {
      const response = await fetch('/api/user/api/v1/profile', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setUser(data.data);
      }
    } catch (error) {
      console.error('Auth error:', error);
    } finally {
      setLoading(false);
    }
  };

  const login = async (email: string, password: string) => {
    const response = await fetch('/api/user/api/v1/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();
    if (data.status === 'success') {
      localStorage.setItem('access_token', data.data.access_token);
      setUser(data.data.user);
    }
    return data;
  };

  const logout = async () => {
    await fetch('/api/user/api/v1/logout', { method: 'POST' });
    localStorage.removeItem('access_token');
    setUser(null);
  };

  return { user, loading, login, logout };
};
```

## 🛠️ Development Setup

### Environment Variables

Create a `.env` file in your frontend project:

```env
# API Configuration
REACT_APP_API_BASE_URL=http://localhost
REACT_APP_USER_SERVICE_URL=http://localhost/api/user
REACT_APP_CONTENT_SERVICE_URL=http://localhost/api/content
REACT_APP_LESSON_SERVICE_URL=http://localhost/api/lesson
REACT_APP_NOTIFICATION_SERVICE_URL=http://localhost/api/notification

# Development
REACT_APP_ENVIRONMENT=development
```

### CORS Configuration

The backend is configured to allow CORS from `http://localhost:3001` by default. Update the CORS configuration in the BFF service if using different ports.

### Error Handling

All services return consistent error responses:

```json
{
  "status": "error",
  "message": "Error description",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": "Specific error details"
  }
}
```

### Health Checks

Check service health before making requests:

```typescript
const checkServiceHealth = async () => {
  try {
    const response = await fetch('http://localhost/health');
    return response.ok;
  } catch {
    return false;
  }
};
```

## 📊 Monitoring & Debugging

### Service Health Endpoints

- **API Gateway:** `http://localhost/health`
- **User Service:** `http://localhost/api/user/health`
- **Content Service:** `http://localhost/api/content/health`
- **Lesson Service:** `http://localhost/api/lesson/health`
- **Notification Service:** `http://localhost/api/notification/health`

### Monitoring Dashboards

- **Grafana:** `http://localhost:3000` (admin/admin)
- **Prometheus:** `http://localhost:9090`
- **Traefik Dashboard:** `http://localhost:8080`

### Common Issues

1. **CORS Errors:** Ensure frontend URL is whitelisted in BFF service
2. **Authentication Failures:** Check JWT token expiration
3. **Service Unavailable:** Verify all services are running with `docker-compose ps`
4. **GraphQL Errors:** Use GraphQL Playground at `http://localhost/api/content/`

## 🔒 Security Best Practices

1. **Store tokens securely** (httpOnly cookies recommended for production)
2. **Implement token refresh** before expiration
3. **Validate all user inputs** on the frontend
4. **Use HTTPS** in production
5. **Implement rate limiting** for API calls
6. **Sanitize GraphQL queries** to prevent injection

## 📱 Mobile Integration

For React Native or mobile apps:

```typescript
// Mobile-specific configuration
const API_BASE_URL = Platform.OS === 'ios' 
  ? 'http://localhost' 
  : 'http://10.0.2.2'; // Android emulator

// Handle network state changes
import NetInfo from '@react-native-community/netinfo';

const useNetworkStatus = () => {
  const [isConnected, setIsConnected] = useState(true);
  
  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener(state => {
      setIsConnected(state.isConnected);
    });
    return unsubscribe;
  }, []);
  
  return isConnected;
};
```

## 🚀 Production Deployment

### Environment Configuration

```env
# Production URLs
REACT_APP_API_BASE_URL=https://api.yourapp.com
REACT_APP_USER_SERVICE_URL=https://api.yourapp.com/api/user
# ... other services
```

### Build Configuration

```json
{
  "scripts": {
    "build": "react-scripts build",
    "start": "react-scripts start",
    "test": "react-scripts test"
  }
}
```

## 📞 Support

For technical support or questions:

1. Check service health endpoints
2. Review Docker Compose logs: `docker-compose logs [service-name]`
3. Monitor Grafana dashboards for system metrics
4. Check Traefik dashboard for request routing

---

**Happy coding! 🎉** This microservice architecture provides a robust, scalable foundation for your English learning application.
