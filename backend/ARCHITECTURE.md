# SkillSync Backend Architecture

> **Comprehensive Technical Documentation**
>
> Last Updated: 2026-01-03

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Overview](#2-system-overview)
3. [Directory Structure](#3-directory-structure)
4. [Database Architecture](#4-database-architecture)
5. [Service Components](#5-service-components)
6. [Internal Packages](#6-internal-packages)
7. [Data Flow & Pipeline](#7-data-flow--pipeline)
8. [Matching Algorithm](#8-matching-algorithm)
9. [External Dependencies](#9-external-dependencies)
10. [Configuration & Environment](#10-configuration--environment)
11. [API Reference](#11-api-reference)

---

## 1. Executive Summary

**SkillSync** is a hybrid job-matching platform backend built in **Go 1.25.5**. It combines:

- **Skill-based matching** (keyword extraction and overlap scoring)
- **Semantic matching** (vector embeddings via Google Gemini)

The system ingests job listings from external APIs (Adzuna), processes user resumes, extracts skills, generates vector embeddings, and produces ranked job recommendations using a weighted hybrid scoring algorithm.

### Key Technologies

| Category | Technology |
|----------|------------|
| Language | Go 1.25.5 |
| Database | PostgreSQL with pgvector extension |
| Embeddings | Google Gemini `text-embedding-004` (768 dimensions) |
| Job Source | Adzuna API |
| DB Driver | pgx/v5 (connection pooling) |
| Config | godotenv (.env files) |

---

## 2. System Overview

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SKILLSYNC BACKEND                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────┐     ┌────────────────────┐     ┌──────────────────┐  │
│  │   EXTERNAL       │     │   BACKGROUND       │     │   API LAYER      │  │
│  │   INTEGRATIONS   │     │   WORKERS          │     │                  │  │
│  ├──────────────────┤     ├────────────────────┤     ├──────────────────┤  │
│  │                  │     │                    │     │                  │  │
│  │  ┌────────────┐  │     │  ┌──────────────┐  │     │  ┌────────────┐  │  │
│  │  │ Adzuna    │  │────▶│  │ jobs-cron    │  │     │  │ /matches   │  │  │
│  │  │ Job API   │  │     │  │              │  │     │  │ endpoint   │  │  │
│  │  └────────────┘  │     │  └──────────────┘  │     │  └────────────┘  │  │
│  │                  │     │         │          │     │        │         │  │
│  │  ┌────────────┐  │     │         ▼          │     │        │         │  │
│  │  │ Gemini    │  │     │  ┌──────────────┐  │     │        ▼         │  │
│  │  │ Embedding │  │◀───▶│  │ embedding-   │  │     │  ┌────────────┐  │  │
│  │  │ API       │  │     │  │ worker       │  │     │  │ Matcher    │  │  │
│  │  └────────────┘  │     │  └──────────────┘  │     │  │ Logic      │  │  │
│  │                  │     │         │          │     │  └────────────┘  │  │
│  └──────────────────┘     │         ▼          │     │        │         │  │
│                           │  ┌──────────────┐  │     │        │         │  │
│                           │  │ skills-      │  │     │        │         │  │
│                           │  │ worker       │  │     │        │         │  │
│                           │  └──────────────┘  │     │        │         │  │
│                           │         │          │     │        │         │  │
│                           │         ▼          │     │        │         │  │
│                           │  ┌──────────────┐  │     │        │         │  │
│                           │  │ user-worker  │  │     │        │         │  │
│                           │  └──────────────┘  │     │        │         │  │
│                           │         │          │     │        │         │  │
│                           └─────────┼──────────┘     └────────┼─────────┘  │
│                                     │                         │            │
│                                     ▼                         ▼            │
│                           ┌─────────────────────────────────────────────┐  │
│                           │              PostgreSQL + pgvector          │  │
│                           │                                             │  │
│                           │  ┌─────────┐  ┌─────────┐  ┌─────────────┐  │  │
│                           │  │  jobs   │  │  users  │  │   skills    │  │  │
│                           │  └────┬────┘  └────┬────┘  └──────┬──────┘  │  │
│                           │       │            │              │         │  │
│                           │       ▼            ▼              ▼         │  │
│                           │  ┌─────────────────────────────────────┐    │  │
│                           │  │   job_skills   │   user_skills      │    │  │
│                           │  └─────────────────────────────────────┘    │  │
│                           └─────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Service Architecture Pattern

The backend follows a **microservices-inspired design** with:

- **1 HTTP API server** (stateless, handles requests)
- **4 background workers** (batch processing, can run as cron jobs)
- **Shared internal packages** (reusable business logic)
- **Single shared database** (PostgreSQL with vector support)

---

## 3. Directory Structure

```
backend/
├── .env                          # Environment configuration
├── .gitignore                    # Git ignore rules
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
│
├── cmd/                          # ━━━ SERVICE ENTRY POINTS ━━━
│   ├── api/
│   │   └── main.go              # HTTP API server (port 8080)
│   ├── embedding-worker/
│   │   └── main.go              # Generates vector embeddings
│   ├── jobs-cron/
│   │   └── main.go              # Fetches jobs from Adzuna API
│   ├── skills-worker/
│   │   └── main.go              # Extracts skills from job descriptions
│   └── user-worker/
│       └── main.go              # Processes user resumes
│
├── internal/                     # ━━━ SHARED BUSINESS LOGIC ━━━
│   ├── adzuna/
│   │   ├── client.go            # Adzuna API HTTP client
│   │   └── types.go             # API response structs
│   ├── db/
│   │   ├── db.go                # Database connection & queries
│   │   └── models.go            # Data models (Job, Skill)
│   ├── embedding/
│   │   └── client.go            # Google Gemini embedding client
│   ├── jobs/                     # (Empty - reserved for future)
│   ├── match/
│   │   └── matcher.go           # Hybrid matching algorithm
│   ├── resume/
│   │   └── parse.go             # Resume text parser
│   └── skills/
│       └── extract.go           # Skill extraction logic
│
└── migrations/                   # ━━━ DATABASE SCHEMA ━━━
    ├── 001_create_jobs.sql      # Jobs table
    ├── 002_create_skills.sql    # Skills + job_skills junction
    ├── 003_create_users.sql     # Users + user_skills junction
    └── 004_add_embeddings.sql   # Vector columns (pgvector)
```

---

## 4. Database Architecture

### 4.1 Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DATABASE SCHEMA (PostgreSQL)                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────┐                      ┌─────────────────────┐       │
│  │       jobs          │                      │       users         │       │
│  ├─────────────────────┤                      ├─────────────────────┤       │
│  │ id           SERIAL │ PK                   │ id           SERIAL │ PK    │
│  │ job_hash     TEXT   │ UNIQUE               │ email        TEXT   │ UNIQUE│
│  │ title        TEXT   │                      │ resume_text  TEXT   │       │
│  │ company      TEXT   │                      │ embedding    VECTOR │ (768) │
│  │ location     TEXT   │                      │ created_at   TIMESTAMP      │
│  │ description  TEXT   │                      └──────────┬──────────┘       │
│  │ apply_url    TEXT   │                                 │                  │
│  │ source       TEXT   │ DEFAULT 'adzuna'                │                  │
│  │ embedding    VECTOR │ (768)                           │                  │
│  │ created_at   TIMESTAMP                                │                  │
│  └──────────┬──────────┘                                 │                  │
│             │                                            │                  │
│             │                ┌─────────────────────┐     │                  │
│             │                │      skills         │     │                  │
│             │                ├─────────────────────┤     │                  │
│             │                │ id     SERIAL       │ PK  │                  │
│             │                │ name   TEXT         │ UNIQUE                 │
│             │                └──────────┬──────────┘     │                  │
│             │                           │                │                  │
│             │         ┌─────────────────┼────────────────┤                  │
│             │         │                 │                │                  │
│             ▼         ▼                 ▼                ▼                  │
│  ┌─────────────────────────┐    ┌─────────────────────────┐                 │
│  │      job_skills         │    │      user_skills        │                 │
│  ├─────────────────────────┤    ├─────────────────────────┤                 │
│  │ job_id    INT           │ FK │ user_id   INT           │ FK              │
│  │ skill_id  INT           │ FK │ skill_id  INT           │ FK              │
│  │ PRIMARY KEY (job_id,    │    │ PRIMARY KEY (user_id,   │                 │
│  │             skill_id)   │    │             skill_id)   │                 │
│  └─────────────────────────┘    └─────────────────────────┘                 │
│                                                                             │
│  ━━━ VECTOR EXTENSION ━━━                                                   │
│  pgvector: Enables VECTOR(768) type and cosine distance operator (<->)      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Table Definitions

#### `jobs` Table

```sql
CREATE TABLE IF NOT EXISTS jobs (
    id          SERIAL PRIMARY KEY,
    job_hash    TEXT NOT NULL UNIQUE,       -- SHA256(title|company|location)
    title       TEXT NOT NULL,
    company     TEXT NOT NULL,
    location    TEXT NOT NULL,
    description TEXT NOT NULL,
    apply_url   TEXT NOT NULL,
    source      TEXT NOT NULL DEFAULT 'adzuna',
    embedding   VECTOR(768),                 -- Gemini text-embedding-004
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Key Points:**
- `job_hash` prevents duplicate job insertions from multiple API fetches
- `embedding` stores the 768-dimensional vector from Gemini
- `source` tracks the origin API (currently only "adzuna")

#### `skills` Table

```sql
CREATE TABLE IF NOT EXISTS skills (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);
```

**Key Points:**
- Normalized skill storage (each skill appears once)
- Case-sensitive unique constraint

#### `job_skills` Junction Table

```sql
CREATE TABLE IF NOT EXISTS job_skills (
    job_id   INT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (job_id, skill_id)
);
```

**Key Points:**
- Many-to-many relationship between jobs and skills
- Cascade deletion maintains referential integrity

#### `users` Table

```sql
CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    email       TEXT UNIQUE,
    resume_text TEXT,                        -- Raw resume content
    embedding   VECTOR(768),                 -- Resume embedding
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### `user_skills` Junction Table

```sql
CREATE TABLE IF NOT EXISTS user_skills (
    user_id  INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, skill_id)
);
```

### 4.3 Vector Operations (pgvector)

The `embedding` columns use the **pgvector** PostgreSQL extension:

```sql
-- Cosine Distance calculation
-- Lower value = more similar (0 = identical, 2 = opposite)
j.embedding <-> u.embedding

-- Similarity calculation (used in matching)
-- Higher value = more similar (1 = identical, -1 = opposite)
1 - (j.embedding <-> u.embedding) AS similarity
```

---

## 5. Service Components

### 5.1 API Server (`cmd/api/main.go`)

**Purpose:** HTTP API server exposing job matching functionality.

**Port:** 8080 (configurable via `PORT` env var)

**Flow:**
```
┌──────────────────────────────────────────────────────────────────┐
│                        API REQUEST FLOW                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Client Request                                                 │
│        │                                                         │
│        ▼                                                         │
│   GET /matches?user_id=123                                       │
│        │                                                         │
│        ▼                                                         │
│   ┌────────────────────────────┐                                 │
│   │  Parse user_id parameter   │                                 │
│   │  Validate is integer       │                                 │
│   └─────────────┬──────────────┘                                 │
│                 │                                                │
│                 ▼                                                │
│   ┌────────────────────────────┐                                 │
│   │  match.MatchJobs(db, id)   │                                 │
│   │  (calls matcher logic)     │                                 │
│   └─────────────┬──────────────┘                                 │
│                 │                                                │
│                 ▼                                                │
│   ┌────────────────────────────┐                                 │
│   │  Return JSON response:     │                                 │
│   │  - JobID, Title, Company   │                                 │
│   │  - SkillScore, EmbedScore  │                                 │
│   │  - MatchScore (hybrid)     │                                 │
│   │  - MissingSkills list      │                                 │
│   └────────────────────────────┘                                 │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

**Key Implementation Details:**

```go
// CORS enabled for frontend
w.Header().Set("Access-Control-Allow-Origin", "*")

// Uses standard library net/http
http.HandleFunc("/matches", handler)
http.ListenAndServe(":"+port, nil)
```

---

### 5.2 Jobs Cron (`cmd/jobs-cron/main.go`)

**Purpose:** Fetches job listings from Adzuna API and stores them in the database.

**Execution:** Run periodically (cron job or manual)

**Flow:**
```
┌──────────────────────────────────────────────────────────────────┐
│                     JOBS-CRON PIPELINE                           │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   1. Initialize                                                  │
│      ├─ Load .env (ADZUNA_APP_ID, ADZUNA_APP_KEY, DATABASE_URL)  │
│      └─ Create DB connection pool                                │
│                                                                  │
│   2. Create Adzuna Client                                        │
│      └─ adzuna.NewClient(appID, appKey)                          │
│                                                                  │
│   3. Fetch Jobs                                                  │
│      └─ client.FetchJobs("software", "india", 3)                 │
│                │                                                 │
│                ▼                                                 │
│      ┌─────────────────────────────────────────────┐             │
│      │ Fetches 3 pages × 50 results = 150 jobs     │             │
│      │ From: api.adzuna.com/v1/api/jobs/in/search  │             │
│      └─────────────────────────────────────────────┘             │
│                                                                  │
│   4. For each job:                                               │
│      ├─ Generate job_hash = SHA256(title|company|location)       │
│      └─ INSERT with ON CONFLICT DO NOTHING (idempotent)          │
│                                                                  │
│   5. Log: "Inserted X new jobs"                                  │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

**Hash Generation:**
```go
func makeJobHash(title, company, location string) string {
    key := title + "|" + company + "|" + location
    hash := sha256.Sum256([]byte(key))
    return hex.EncodeToString(hash[:])
}
```

---

### 5.3 Skills Worker (`cmd/skills-worker/main.go`)

**Purpose:** Extracts skills from job descriptions using keyword matching.

**Flow:**
```
┌──────────────────────────────────────────────────────────────────┐
│                    SKILLS-WORKER PIPELINE                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   1. Query: SELECT id, description FROM jobs                     │
│             WHERE id NOT IN (SELECT DISTINCT job_id              │
│                              FROM job_skills)                    │
│      ↳ Finds jobs that haven't been processed yet                │
│                                                                  │
│   2. For each unprocessed job:                                   │
│      │                                                           │
│      ├─ 2a. skills.Extract(description)                          │
│      │      └─ Matches against KnownSkills list                  │
│      │      └─ Returns: ["go", "postgres", "docker", ...]        │
│      │                                                           │
│      └─ 2b. For each found skill:                                │
│             ├─ GetOrCreateSkill(skillName) → skillID             │
│             └─ InsertJobSkill(jobID, skillID)                    │
│                                                                  │
│   3. Log: "Skill extraction completed"                           │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

> **Note:** The worker uses `context.Background()` which may cause a compilation error since `context` is not imported. This appears to be a bug in the current implementation.

---

### 5.4 Embedding Worker (`cmd/embedding-worker/main.go`)

**Purpose:** Generates vector embeddings for jobs and users using Google Gemini API.

**Flow:**
```
┌──────────────────────────────────────────────────────────────────┐
│                   EMBEDDING-WORKER PIPELINE                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ━━━ PHASE 1: JOB EMBEDDINGS ━━━                                │
│                                                                  │
│   1. Query: SELECT id, (title || ' ' || description) AS text     │
│             FROM jobs WHERE embedding IS NULL                    │
│             LIMIT 50                                             │
│                                                                  │
│   2. For each job without embedding:                             │
│      ├─ embedder.Embed(text)                                     │
│      │      └─ POST to Gemini text-embedding-004                 │
│      │      └─ Returns: []float32 (768 dimensions)               │
│      └─ database.UpdateJobEmbedding(jobID, vec)                  │
│             └─ UPDATE jobs SET embedding = $2::vector            │
│                                                                  │
│   ━━━ PHASE 2: USER EMBEDDINGS ━━━                               │
│                                                                  │
│   3. Query: SELECT id, resume_text FROM users                    │
│             WHERE embedding IS NULL                              │
│               AND resume_text IS NOT NULL                        │
│             LIMIT 20                                             │
│                                                                  │
│   4. For each user without embedding:                            │
│      ├─ embedder.Embed(resume_text)                              │
│      └─ database.UpdateUserEmbedding(userID, vec)                │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

**Gemini API Call:**
```go
// Endpoint
https://generativelanguage.googleapis.com/v1beta/models/text-embedding-004:embedContent?key={API_KEY}

// Request Body
{
  "content": {
    "parts": [
      { "text": "..." }
    ]
  }
}

// Response
{
  "embedding": {
    "values": [0.123, -0.456, ...]  // 768 floats
  }
}
```

---

### 5.5 User Worker (`cmd/user-worker/main.go`)

**Purpose:** Processes user resumes, extracts skills, and stores user-skill associations.

**Flow:**
```
┌──────────────────────────────────────────────────────────────────┐
│                     USER-WORKER PIPELINE                         │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   1. CreateUser("test@example.com")                              │
│      └─ INSERT INTO users ... ON CONFLICT (email) DO UPDATE      │
│      └─ Returns: userID                                          │
│                                                                  │
│   2. rawResume = (hardcoded test resume)                         │
│      "Backend developer with experience in Go, PostgreSQL,       │
│       Docker, Linux, and AWS. Worked on REST APIs."              │
│                                                                  │
│   3. resume.Parse(rawResume)                                     │
│      └─ Currently: strings.TrimSpace(text)                       │
│      └─ Future: PDF/DOCX parsing support                         │
│                                                                  │
│   4. skills.Extract(text)                                        │
│      └─ Returns: ["go", "postgresql", "docker", "linux", "aws"]  │
│                                                                  │
│   5. For each found skill:                                       │
│      ├─ GetOrCreateSkill(skill) → skillID                        │
│      └─ InsertUserSkill(userID, skillID)                         │
│                                                                  │
│   6. Log: "User resume processed successfully"                   │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

> **Note:** Currently uses a hardcoded test resume. In production, this would accept user input.

---

## 6. Internal Packages

### 6.1 Package: `internal/db`

**Files:** `db.go`, `models.go`

**Purpose:** Database connection management and all SQL operations.

#### Connection Pool

```go
type DB struct {
    Pool *pgxpool.Pool
}

func New(databaseURL string) (*DB, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    pool, err := pgxpool.New(ctx, databaseURL)
    // ... connection validation with Ping()
    return &DB{Pool: pool}, nil
}
```

**Features:**
- Uses `pgx/v5` (high-performance PostgreSQL driver)
- Connection pooling via `pgxpool`
- 5-second connection timeout
- Connection validation via Ping

#### Available Methods

| Method | Description |
|--------|-------------|
| `New(url)` | Create connection pool |
| `Close()` | Shutdown pool |
| `InsertJob(job)` | Insert job with conflict handling |
| `GetOrCreateSkill(name)` | Upsert skill, return ID |
| `InsertJobSkill(jobID, skillID)` | Link job to skill |
| `CreateUser(email)` | Upsert user, return ID |
| `InsertUserSkill(userID, skillID)` | Link user to skill |
| `GetJobsWithoutEmbeddings(limit)` | Find jobs needing embeddings |
| `GetUsersWithoutEmbeddings(limit)` | Find users needing embeddings |
| `UpdateJobEmbedding(jobID, vec)` | Store job embedding |
| `UpdateUserEmbedding(userID, vec)` | Store user embedding |

#### Vector Formatting

```go
func (db *DB) formatVector(vec []float32) string {
    // Converts []float32{0.1, 0.2, 0.3}
    // To string: "[0.1,0.2,0.3]"
    // For pgvector's ::vector cast
}
```

#### Data Models

```go
type Job struct {
    ID          int
    JobHash     string
    Title       string
    Company     string
    Location    string
    Description string
    ApplyURL    string
    Source      string
    CreatedAt   time.Time
}

type Skill struct {
    ID   int
    Name string
}
```

---

### 6.2 Package: `internal/adzuna`

**Files:** `client.go`, `types.go`

**Purpose:** HTTP client for Adzuna Job Search API.

#### Client Structure

```go
type Client struct {
    AppID  string
    AppKey string
    Client *http.Client  // 10-second timeout
}
```

#### API Integration

```go
// API Endpoint Pattern
https://api.adzuna.com/v1/api/jobs/in/search/{page}?
    app_id={ID}&
    app_key={KEY}&
    results_per_page=50&
    what={query}&
    where={location}
```

#### Response Types

```go
type apiResponse struct {
    Results []apiJob `json:"results"`
}

type apiJob struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    RedirectURL string `json:"redirect_url"`
    Company     struct {
        DisplayName string `json:"display_name"`
    } `json:"company"`
    Location struct {
        DisplayName string `json:"display_name"`
    } `json:"location"`
}
```

---

### 6.3 Package: `internal/embedding`

**Files:** `client.go`

**Purpose:** Google Gemini embedding API client.

#### Configuration

```go
type GeminiClient struct {
    APIKey string
    Client *http.Client  // 15-second timeout
}
```

#### Embedding Model

- **Model:** `text-embedding-004`
- **Dimensions:** 768
- **API Version:** v1beta

#### Request/Response Flow

```go
// Request
type geminiEmbedRequest struct {
    Content geminiContent `json:"content"`
}

type geminiContent struct {
    Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
    Text string `json:"text"`
}

// Response
type geminiEmbedResponse struct {
    Embedding struct {
        Values []float32 `json:"values"`
    } `json:"embedding"`
}
```

---

### 6.4 Package: `internal/skills`

**Files:** `extract.go`

**Purpose:** Keyword-based skill extraction from text.

#### Known Skills List

```go
var KnownSkills = []string{
    "go", "golang",
    "python",
    "java",
    "javascript",
    "typescript",
    "react",
    "node",
    "sql",
    "postgres",
    "docker",
    "aws",
    "git",
    "linux",
}
```

#### Extraction Algorithm

```go
func Extract(text string) []string {
    text = strings.ToLower(text)
    found := make(map[string]bool)

    for _, skill := range KnownSkills {
        if strings.Contains(text, skill) {
            found[normalize(skill)] = true
        }
    }
    // ... return unique skills
}
```

#### Normalization Rules

```go
func normalize(skill string) string {
    switch skill {
    case "golang":  return "go"
    case "postgres": return "postgresql"
    default: return skill
    }
}
```

---

### 6.5 Package: `internal/resume`

**Files:** `parse.go`

**Purpose:** Resume text parsing (currently minimal).

```go
// Current implementation (placeholder)
func Parse(text string) string {
    return strings.TrimSpace(text)
}

// Future: PDF/DOCX parsing support
```

---

### 6.6 Package: `internal/match`

**Files:** `matcher.go`

**Purpose:** Hybrid job-user matching algorithm.

*See [Section 8: Matching Algorithm](#8-matching-algorithm) for detailed explanation.*

---

## 7. Data Flow & Pipeline

### Complete System Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        COMPLETE DATA PIPELINE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ━━━ STAGE 1: JOB INGESTION ━━━                                             │
│                                                                             │
│  [Adzuna API] ──┬──▶ [jobs-cron] ──▶ [INSERT into jobs table]               │
│                 │                          │                                │
│                 │    job_hash = SHA256     │                                │
│                 │    (prevents duplicates) │                                │
│                 │                          ▼                                │
│                 │                   ┌─────────────┐                         │
│                 │                   │    jobs     │                         │
│                 │                   │  (no embed) │                         │
│                 │                   └─────────────┘                         │
│                 │                                                           │
│  ━━━ STAGE 2: SKILL EXTRACTION ━━━                                          │
│                                                                             │
│  [skills-worker] ─────────▶ For each job without skills:                    │
│         │                   1. Extract from description                     │
│         │                   2. GetOrCreateSkill()                           │
│         │                   3. InsertJobSkill()                             │
│         │                          │                                        │
│         │                          ▼                                        │
│         │                   ┌─────────────┐    ┌─────────────┐              │
│         │                   │   skills    │───▶│ job_skills  │              │
│         │                   └─────────────┘    └─────────────┘              │
│         │                                                                   │
│  ━━━ STAGE 3: EMBEDDING GENERATION ━━━                                      │
│                                                                             │
│  [embedding-worker] ──┬──▶ For each entity without embedding:               │
│         │             │    1. Send text to Gemini API                       │
│         │             │    2. Receive 768-dim vector                        │
│         │             │    3. UPDATE embedding column                       │
│         │             │                                                     │
│         │             ├──▶ Jobs: title + description → embedding            │
│         │             │                                                     │
│         │             └──▶ Users: resume_text → embedding                   │
│         │                                                                   │
│  ━━━ STAGE 4: USER ONBOARDING ━━━                                           │
│                                                                             │
│  [User Signup] ───────▶ [user-worker] ──▶ 1. CreateUser()                   │
│         │                      │          2. Parse resume                   │
│         │                      │          3. Extract skills                 │
│         │                      │          4. InsertUserSkill()              │
│         │                      │                 │                          │
│         │                      │                 ▼                          │
│         │                      │          ┌─────────────┐                   │
│         │                      │          │ user_skills │                   │
│         │                      │          └─────────────┘                   │
│         │                      │                                            │
│  ━━━ STAGE 5: MATCHING & API ━━━                                            │
│                                                                             │
│  [Frontend] ──▶ GET /matches?user_id=X ──▶ [API Server]                     │
│       ▲                                         │                           │
│       │                                         ▼                           │
│       │                                   [Matcher Logic]                   │
│       │                                   1. Skill overlap score            │
│       │                                   2. Embedding similarity           │
│       │                                   3. Weighted hybrid score          │
│       │                                   4. Missing skills calculation     │
│       │                                         │                           │
│       │                                         ▼                           │
│       └───────────────────────── JSON Response: Ranked Jobs                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Worker Execution Order (Recommended)

```
1. jobs-cron          (First: populate jobs)
        ↓
2. skills-worker      (Second: extract skills from jobs)
        ↓
3. user-worker        (Third: process user resumes)
        ↓
4. embedding-worker   (Fourth: generate all embeddings)
        ↓
5. api server         (Always running: serve match requests)
```

---

## 8. Matching Algorithm

### 8.1 Algorithm Overview

The matching algorithm uses a **hybrid approach** combining:

1. **Skill-based matching** (keyword overlap)
2. **Semantic matching** (vector similarity)

### 8.2 Detailed Algorithm

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MATCHING ALGORITHM                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  INPUT: user_id                                                             │
│                                                                             │
│  ━━━ STEP 1: EMBEDDING SIMILARITY ━━━                                       │
│                                                                             │
│  SELECT                                                                     │
│      j.id,                                                                  │
│      1 - (j.embedding <-> u.embedding) AS similarity                        │
│  FROM jobs j                                                                │
│  JOIN users u ON u.id = $user_id                                            │
│  WHERE j.embedding IS NOT NULL AND u.embedding IS NOT NULL                  │
│  ORDER BY similarity DESC                                                   │
│  LIMIT 200                                                                  │
│                                                                             │
│  Result: Map[jobID] → similarity (0.0 to 1.0)                               │
│                                                                             │
│  ━━━ STEP 2: SKILL OVERLAP SCORE ━━━                                        │
│                                                                             │
│  SELECT                                                                     │
│      j.id, j.title, j.company, j.location,                                  │
│      COUNT(matched_skills) / COUNT(job_skills) AS skill_score               │
│  FROM jobs j                                                                │
│  LEFT JOIN job_skills js ON j.id = js.job_id                                │
│  LEFT JOIN user_skills us ON js.skill_id = us.skill_id                      │
│                          AND us.user_id = $user_id                          │
│  GROUP BY j.id                                                              │
│  ORDER BY skill_score DESC                                                  │
│  LIMIT 100                                                                  │
│                                                                             │
│  Formula: skill_score = (user skills matching job) / (total job skills)     │
│                                                                             │
│  ━━━ STEP 3: HYBRID SCORE CALCULATION ━━━                                   │
│                                                                             │
│  For each job:                                                              │
│      MatchScore = 0.7 × SkillScore + 0.3 × EmbedScore                       │
│                                                                             │
│  Weights:                                                                   │
│  ┌────────────────────────────────────────────┐                             │
│  │  Skill Score    │  70%  │  Keyword match   │                             │
│  │  Embed Score    │  30%  │  Semantic match  │                             │
│  └────────────────────────────────────────────┘                             │
│                                                                             │
│  ━━━ STEP 4: MISSING SKILLS IDENTIFICATION ━━━                              │
│                                                                             │
│  SELECT s.name                                                              │
│  FROM skills s                                                              │
│  JOIN job_skills js ON s.id = js.skill_id                                   │
│  WHERE js.job_id = $job_id                                                  │
│    AND s.id NOT IN (SELECT skill_id FROM user_skills                        │
│                     WHERE user_id = $user_id)                               │
│                                                                             │
│  ━━━ STEP 5: FINAL RANKING ━━━                                              │
│                                                                             │
│  Sort all job matches by MatchScore (descending)                            │
│                                                                             │
│  OUTPUT: []JobMatch{                                                        │
│      JobID, Title, Company, Location,                                       │
│      SkillScore, EmbedScore, MatchScore,                                    │
│      MissingSkills[]                                                        │
│  }                                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 8.3 Score Interpretation

| Score Range | Interpretation |
|-------------|----------------|
| 0.9 - 1.0 | Excellent match (near-perfect skill overlap + semantic fit) |
| 0.7 - 0.9 | Good match |
| 0.5 - 0.7 | Moderate match (some skill gaps) |
| 0.3 - 0.5 | Weak match |
| 0.0 - 0.3 | Poor match |

### 8.4 Response Structure

```go
type JobMatch struct {
    JobID    int
    Title    string
    Company  string
    Location string

    SkillScore float64    // 0.0 - 1.0 (keyword match)
    EmbedScore float64    // 0.0 - 1.0 (semantic similarity)
    MatchScore float64    // 0.0 - 1.0 (weighted hybrid)

    MissingSkills []string // Skills to learn
}
```

---

## 9. External Dependencies

### 9.1 Go Module Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/jackc/pgx/v5` | v5.7.6 | PostgreSQL driver |
| `github.com/jackc/puddle/v2` | v2.2.2 | Connection pool |
| `github.com/joho/godotenv` | v1.5.1 | .env file loading |
| `github.com/gin-gonic/gin` | v1.11.0 | HTTP framework (unused?) |
| `google.golang.org/protobuf` | v1.36.9 | Protobuf support |

> **Note:** Gin framework is in dependencies but the API uses standard `net/http`. Consider removing if unused.

### 9.2 External APIs

| API | Usage | Rate Limits |
|-----|-------|-------------|
| **Adzuna Job API** | Job listings ingestion | Depends on plan |
| **Google Gemini Embedding API** | Vector generation | Depends on quota |

### 9.3 Database Requirements

| Requirement | Version | Notes |
|-------------|---------|-------|
| PostgreSQL | 12+ | Required for pgvector |
| pgvector extension | 0.5.0+ | Must be installed |

---

## 10. Configuration & Environment

### Required Environment Variables

```bash
# Database
DATABASE_URL=postgresql://user:password@host:5432/skillsync

# Adzuna API (for jobs-cron)
ADZUNA_APP_ID=your_app_id
ADZUNA_APP_KEY=your_app_key

# Google Gemini (for embedding-worker)
GEMINI_API_KEY=your_gemini_api_key

# API Server (optional)
PORT=8080  # defaults to 8080 if not set
```

### .env File Location

The `.env` file should be placed in the `backend/` root directory. All services use `godotenv.Load()` which automatically loads from the current directory.

---

## 11. API Reference

### GET `/matches`

Retrieves ranked job matches for a user.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | integer | Yes | User ID to match jobs for |

**Request Example:**
```http
GET /matches?user_id=123 HTTP/1.1
Host: localhost:8080
```

**Response Example:**
```json
[
  {
    "JobID": 42,
    "Title": "Senior Go Developer",
    "Company": "TechCorp",
    "Location": "Bangalore, India",
    "SkillScore": 0.75,
    "EmbedScore": 0.82,
    "MatchScore": 0.771,
    "MissingSkills": ["kubernetes", "terraform"]
  },
  {
    "JobID": 87,
    "Title": "Backend Engineer",
    "Company": "StartupXYZ",
    "Location": "Remote",
    "SkillScore": 0.60,
    "EmbedScore": 0.78,
    "MatchScore": 0.654,
    "MissingSkills": ["python", "machine-learning"]
  }
]
```

**Error Responses:**

| Status | Body | Description |
|--------|------|-------------|
| 400 | `Missing user_id parameter` | user_id not provided |
| 400 | `Invalid user_id parameter` | user_id not an integer |
| 500 | `Internal Server Error` | Database or matching error |

**Headers:**
```http
Content-Type: application/json
Access-Control-Allow-Origin: *
```

---

## Appendix A: SQL Query Reference

### Idempotent Insert Patterns

```sql
-- Jobs (dedup by hash)
INSERT INTO jobs (job_hash, title, ...)
VALUES ($1, $2, ...)
ON CONFLICT (job_hash) DO NOTHING;

-- Skills (upsert, return ID)
INSERT INTO skills (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- Junction tables (ignore duplicates)
INSERT INTO job_skills (job_id, skill_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
```

### Find Unprocessed Records

```sql
-- Jobs without skills
SELECT id, description FROM jobs
WHERE id NOT IN (SELECT DISTINCT job_id FROM job_skills);

-- Jobs without embeddings
SELECT id, title || ' ' || description AS text
FROM jobs
WHERE embedding IS NULL
LIMIT 50;

-- Users without embeddings (must have resume)
SELECT id, resume_text
FROM users
WHERE embedding IS NULL
  AND resume_text IS NOT NULL
LIMIT 20;
```

### Vector Similarity

```sql
-- Cosine similarity (1 = identical)
SELECT
    j.id,
    1 - (j.embedding <-> u.embedding) AS similarity
FROM jobs j
JOIN users u ON u.id = $1
WHERE j.embedding IS NOT NULL
  AND u.embedding IS NOT NULL
ORDER BY similarity DESC
LIMIT 200;
```

---

## Appendix B: Known Issues & TODOs

1. **Missing Import in skills-worker**
   - `context` package not imported but `context.Background()` is used
   - Will cause compilation error

2. **Hardcoded Test Data in user-worker**
   - Resume text is hardcoded
   - Needs integration with actual user input

3. **Limited Skill Dictionary**
   - Only 14 known skills
   - Should be expanded or use ML-based extraction

4. **Resume Parser Placeholder**
   - Currently only trims whitespace
   - Needs PDF/DOCX support

5. **No API Authentication**
   - `/matches` endpoint is open
   - Consider adding authentication

6. **Gin Framework Unused**
   - In dependencies but using standard library
   - Consider cleaning up

---

*End of Architecture Documentation*
