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
- `POST /query` -> GraphQL endpoint
- `GET /` -> GraphQL Playground (development)

## GraphQL Examples

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


## SQL
```sql
-- taxonomies
CREATE TABLE topics (
  id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  slug       text UNIQUE NOT NULL,
  name       text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE levels (
  id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  code       text UNIQUE NOT NULL, -- e.g., A1, A2, B1...
  name       text NOT NULL
);

CREATE TABLE tags (
  id   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  slug text UNIQUE NOT NULL,
  name text NOT NULL
);

CREATE TABLE media_assets (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  storage_key  text NOT NULL UNIQUE, -- S3/MinIO object key
  kind         text NOT NULL CHECK (kind IN ('image','audio')),
  mime_type    text NOT NULL,
  bytes        integer,
  duration_ms  integer,              -- for audio
  sha256       text NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  uploaded_by  uuid                  -- logical FK to users.id
);
CREATE INDEX media_sha_idx ON media_assets(sha256);

-- lessons are modular, versioned
CREATE TABLE lessons (
  id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  code          text UNIQUE,                    -- human ID, optional
  title         text NOT NULL,
  description   text,
  topic_id      uuid REFERENCES topics(id),
  level_id      uuid REFERENCES levels(id),
  is_published  boolean NOT NULL DEFAULT false,
  version       integer NOT NULL DEFAULT 1,
  created_by    uuid,                           -- logical FK
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  published_at  timestamptz
);
CREATE INDEX lessons_topic_level_pub_idx ON lessons(topic_id, level_id, is_published);

-- lesson sections (content blocks)
CREATE TABLE lesson_sections (
  id          uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  lesson_id   uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
  ord         integer NOT NULL,
  type        text NOT NULL CHECK (type IN ('text','dialog','audio','image','exercise')),
  body        jsonb NOT NULL DEFAULT '{}'::jsonb, -- flexible payload
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX lesson_sections_ord ON lesson_sections(lesson_id, ord);

-- flashcards
CREATE TABLE flashcard_sets (
  id          uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  title       text NOT NULL,
  description text,
  topic_id    uuid REFERENCES topics(id),
  level_id    uuid REFERENCES levels(id),
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid
);

CREATE TABLE flashcards (
  id               uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  set_id           uuid NOT NULL REFERENCES flashcard_sets(id) ON DELETE CASCADE,
  front_text       text NOT NULL,
  back_text        text NOT NULL,
  front_media_id   uuid REFERENCES media_assets(id),
  back_media_id    uuid REFERENCES media_assets(id),
  ord              integer NOT NULL,
  hints            text[],
  created_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX flashcards_set_ord ON flashcards(set_id, ord);

-- quizzes
CREATE TABLE quizzes (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  lesson_id    uuid REFERENCES lessons(id) ON DELETE SET NULL,
  title        text NOT NULL,
  description  text,
  total_points integer NOT NULL DEFAULT 0,
  time_limit_s integer,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE quiz_questions (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  quiz_id      uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
  ord          integer NOT NULL,
  type         text NOT NULL CHECK (type IN ('mcq','multi_select','fill_blank','audio_transcribe','match','ordering')),
  prompt       text NOT NULL,
  prompt_media uuid REFERENCES media_assets(id),
  points       integer NOT NULL DEFAULT 1,
  metadata     jsonb NOT NULL DEFAULT '{}'::jsonb
);
CREATE UNIQUE INDEX quiz_questions_ord ON quiz_questions(quiz_id, ord);

CREATE TABLE question_options (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  question_id  uuid NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
  ord          integer NOT NULL,
  label        text NOT NULL,    -- shown text
  is_correct   boolean NOT NULL DEFAULT false,
  feedback     text
);
CREATE UNIQUE INDEX question_options_ord ON question_options(question_id, ord);

-- tagging
CREATE TABLE content_tags (
  tag_id    uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  kind      text NOT NULL CHECK (kind IN ('lesson','quiz','flashcard_set')),
  object_id uuid NOT NULL,
  PRIMARY KEY(tag_id, kind, object_id)
);

-- outbox
CREATE TABLE outbox (
  id           bigserial PRIMARY KEY,
  aggregate_id uuid NOT NULL,
  topic        text NOT NULL,      -- content.events
  type         text NOT NULL,      -- LessonPublished, QuizCreated, …
  payload      jsonb NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz
);
CREATE INDEX outbox_unpub_idx ON outbox(published_at) WHERE published_at IS NULL;

```