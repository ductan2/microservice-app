# content-services

GraphQL API for content management (lessons, flashcards, quizzes) using gqlgen.

## Setup

### Environment

This service reads configuration from a `.env` file (loaded via `godotenv`).

1) Create `content-services/.env` from the example:

```bash
cd content-services
cp .env.example .env
```

2) Edit values as needed. Available variables:

```
# Server
PORT=8004

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DB=content

# GraphQL
# Enable GraphQL Playground at '/'
GRAPHQL_PLAYGROUND=true
```

First time setup (generate GraphQL code):

```bash
./scripts/setup.sh
```

Or manually:

```bash
go mod tidy
go run github.com/99designs/gqlgen generate
```

## Run

```bash
make run
```

Or directly:

```bash
go run ./cmd/server
```

Set a custom port via `PORT` environment variable (defaults to 8004):

```bash
PORT=9000 go run ./cmd/server
```

## Endpoints

- `GET /health` -> `{ "status": "ok" }`
- `POST /graphql` -> GraphQL endpoint
- `GET /` -> GraphQL Playground (development, optional)

## GraphQL Examples

### Query Taxonomy

```graphql
query {
  topics {
    id
    slug
    name
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

### Create a Topic

```graphql
mutation {
  createTopic(input: { slug: "travel", name: "Travel" }) {
    id
    slug
    name
  }
}
```

### Query Lessons

```graphql
query {
  lessons(
    filter: { isPublished: true }
    pagination: { page: 1, pageSize: 10 }
  ) {
    data {
      id
      title
      description
      sections {
        type
        body
      }
    }
    total
  }
}
```

### Create Lesson

```graphql
mutation {
  createLesson(input: {
    title: "Basic Grammar"
    description: "Learn English grammar basics"
    topicId: "uuid"
    levelId: "uuid"
  }) {
    id
    title
  }
}
```
