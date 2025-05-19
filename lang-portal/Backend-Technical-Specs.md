# Backend Server Technical Specs
## Business Goal 
A language learning school wants to build a prototype of learning portal which will act as three things:
- Inventory of possible vocabulary that can be learned
- Act as a Learning record store (LRS), providing correct and wrong score on practice vocabulary
- A unified launchpad to launch different learning apps

You have been tasked with creating the backend API of the application.

## Technical Requirements

- The backend will be built using Go
- The database will be SQLite3
- The API will be built using Gin
- The API will always return JSON
- There will be no authentication or authorization
- Everything will be treated as a single user

## Database Schema

Our database will be a single sqlite3 database called `words.db` 
that will be in the root og the project folder of `backend_go`.

## Directory Structure

```text 
backend_go/
├── words.db            # SQLite database file
├── cmd/
│   └── server/         # Main application entry
├── internal/
│   ├── api/            # API handlers
│   │   ├── handlers/   # Individual endpoint handlers
│   │   │   ├── dashboard.go # Dashboard endpoints
│   │   │   ├── groups.go    # Group endpoints
│   │   │   ├── settings.go  # Settings endpoints
│   │   │   ├── study.go     # Study activity/session endpoints
│   │   │   └── words.go     # Word endpoints
│   │   ├── middleware/   # HTTP middleware
│   │   │   └── pagination.go
│   │   └── router.go      # Gin router setup
│   │
│   ├── models/         # Database models/entities
│   │   ├── word.go
│   │   ├── group.go
│   │   ├── study.go
│   │   └── review.go
│   │
│   ├── repository/     # Database operations
│   │   ├── word.go
│   │   ├── group.go
│   │   ├── study.go
│   │   └── review.go
│   │
│   └── service/        # Business logic layer
│       ├── dashboard.go
│       ├── group.go
│       ├── study.go
│       └── word.go
├── db/                 # Database related files
│   ├── migrations/     # SQL migration files
│   │   ├── 00001_init.sql
│   │   └── 00002_create_words_table.sql
│   └── seeds/         # JSON seed data
│       └── basic_verbs.json
├── magefile.go         # Task runner
├── go.mod
└── go.sum
```

We have the following tables:

- words - stored vocabulary words
    - id: integer
    - japanese: string
    - romaji: string
    - english: string
    - parts: json
    - created_at: timestamp

- word_groups - join table for words and groups 
{many-to-many}
    - id: integer
    - word_id: integer
    - group_id: integer
    - created_at: timestamp

- groups - thematic groups of words
    - id: integer
    - name: string
    - created_at: timestamp

- study_sessions - records of study session grouping word_review_items
    - id: integer
    - group_id: integer
    - created_at: timestamp
    - study_activity_id: integer

- study_activities - a specific study activity, linking a study session to a group
    - id: integer
    - study_session_id: integer
    - group_id: integer
    - created_at: timestamp

- word_review_items - a record of word practice, determining if the word was correct or not
    - word_id: integer
    - study_session_id: integer
    - correct: boolean
    - created_at: timestamp

### API Endpoints

- GET /api/dashboard/last_study_session
- GET /api/dashboard/study_progress
- GET /api/dashboard/quick_stats
- GET /api/study_activities
- GET /api/study_activities/:id
- GET /api/study_activities/:id/study_sessions
- POST /api/study_activities
    - required params: group_id, study_activity_id
- GET /api/words
    - pagination with 100 items per page
- GET /api/words/:id
- GET /api/groups
    - pagination with 100 items per page
- GET /api/groups/:id
- GET /api/groups/:id/words
- GET /api/groups/:id/study_sessions
- GET /api/study_sessions
    - pagination with 100 items per page
- GET /api/study_sessions/:id
- GET /api/study_sessions/:id/words
- POST /api/settings/theme
- POST /api/settings/reset_history
- POST /api/settings/full_reset
- POST /api/study_sessions/:id/words/:word_id/review
    - required params: correct

## API Response Documentation

### Dashboard Endpoints

#### GET /api/dashboard/last_study_session
Returns information about the most recent study session.

##### JSON Response
```json
{
  "id": 123,
  "group_id": 456,
  "created_at": "2024-03-20T15:30:00Z",
  "study_activity_id": 789,
  "group_id": 456,
  "group_name": "Basic Greetings"
}
```

#### GET /api/dashboard/study_progress
Returns study progress statistics.
Please note that the frontend will determine progress bar based on total words 
and total available words.

##### JSON Response

```json
{
  "total_words_studied": 3,
  "total_available_words": 124,
}
```

#### GET /api/dashboard/quick_stats
Returns quick overview statistics.

##### JSON Response

```json
{
  "stats": {
    "success_rate": 80,
    "total_study_sessions": 4,
    "total_active_groups": 3,
    "study_streak_days": 1
  }
}
```

### Study Activities Endpoints

#### GET /api/study_activities
Returns list of study activities for the Study Activities Index page.

##### JSON Response

```json
{
  "study_activities": [
    {
      "id": 1,
      "name": "Vocabulary Quiz",
      "thumbnail_url": "/images/flashcards.png",
      "description": "Practice words with flashcards"
    }
  ]
}
```

#### GET /api/study_activities/:id
Returns detailed information for the Study Activity Show page.

##### JSON Response

```json
{
  "id": 1,
  "name": "Vocabulary Quiz",
  "thumbnail_url": "/images/flashcards.png",
  "description": "Practice words with flashcards",
  "available_groups": [
    {
      "id": 1,
      "name": "Basic Verbs"
    }
  ]
}
```

### GET /api/study_activities/:id/study_sessions
Returns study sessions for the Study Activity Show page.
  - pagination with 20 items per page

##### JSON Response

```json
{
  "items": [
    {
      "id": 123,
      "activity_name": "Flashcards",
      "group_name": "Basic Verbs",
      "start_time": "2024-03-20T15:30:00Z",
      "end_time": "2024-03-20T15:45:00Z",
      "review_items_count": 20
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items":100,
    "items_per_page": 20
  }
}
```

#### POST /api/study_activities
Creates a new study session for the Study Activities Launch page.

##### Request Params
- group_id: integer
- study_activity_id: integer

##### JSON Response

```json
{
    "id": 123,
    "group_id": 456
}
```

#### Words Endpoints

#### GET /api/words
Returns paginated list of words for the Words Index page.
- pagination with 100 items per page

##### JSON Response

```json
{
  "items": [
    {
      "id": 1,
      "japanese": "食べる",
      "romaji": "taberu",
      "english": "to eat",
      "correct_count": 12,
      "wrong_count": 3
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 500,
    "items_per_page": 100
  }
}
```

#### GET /api/words/:id
Returns detailed word information for the Word Show page.

##### JSON Response

```json
{
  "word": {
    "id": 1,
    "japanese": "食べる",
    "romaji": "taberu",
    "english": "to eat",
    "study_stats": {
      "correct_count": 12,
      "wrong_count": 3
    },
    "groups": [
      {
        "id": 1,
        "name": "Basic Verbs"
      }
    ]
  }
}
```

### Groups Endpoints

#### GET /api/groups
Returns paginated list of groups for the Word Groups Index page.
- pagination with 100 items per page

##### JSON Response

```json
{
  "items": [
    {
      "id": 1,
      "name": "Basic Verbs",
      "word_count": 50
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_items": 250,
    "items_per_page": 100
  }
}
```

#### GET /api/groups/:id
Returns group details for the Group Show page.

##### JSON Response
```json
{
  "id": 1,
  "name": "Basic Verbs",
  "word_count": 50
}
```

#### GET /api/groups/:id/words
Returns words in the group for the Group Show page.
- pagination with 100 items per page

##### JSON Response

```json
{
  "items": [
    {
      "id": 1,
      "japanese": "食べる",
      "romaji": "taberu",
      "english": "to eat",
      "correct_count": 12,
      "wrong_count": 3
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_items": 250,
    "items_per_page": 100
  }
}
```

#### GET /api/groups/:id/study_sessions
Returns study sessions for the group for the Group Show page.
- pagination with 100 items per page

##### JSON Response

```json
{
  "items": [
    {
      "id": 123,
      "activity_name": "Flashcards",
      "group_name": "Basic Verbs",
      "start_time": "2024-03-20T15:30:00Z",
      "end_time": "2024-03-20T15:45:00Z",
      "review_items_count": 20,
      "success_rate": 75
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_items": 250,
    "items_per_page": 100
  }
}
```

### Study Sessions Endpoints

#### GET /api/study_sessions
Returns paginated list of study sessions for the Study Sessions Index page.
- pagination with 100 items per page

##### JSON Response

```json
{
  "items": [
    {
      "id": 123,
      "activity_name": "Flashcards",
      "group_name": "Basic Verbs",
      "start_time": "2024-03-20T15:30:00Z",
      "end_time": "2024-03-20T15:45:00Z",
      "review_items_count": 20
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 450,
    "items_per_page": 100
  }
}
```

#### GET /api/study_sessions/:id
Returns study session details for the Study Session Show page.

##### JSON Response

```json
{
  "id": 123,
  "activity_name": "Flashcards",
  "group_name": "Basic Verbs",
  "start_time": "2024-03-20T15:30:00Z",
  "end_time": "2024-03-20T15:45:00Z",
  "review_items_count": 2
}
```

#### GET /api/study_sessions/:id/words
Returns words reviewed in the study session for the Study Session Show page.

##### JSON Response
```json
{
  "items": [
    {
      "id": 1,
      "japanese": "食べる",
      "romaji": "taberu",
      "english": "to eat",
      "correct": true,
      "reviewed_at": "2024-03-20T15:31:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_items": 250,
    "items_per_page": 100
  }
}
```

#### POST /api/study_sessions/:id/words/:word_id/review
Records a word review for a study session.

##### Request Params
- id: integer (study_session_id)
- word_id: integer
- correct: boolean

##### Request Payload

```json
{
  "correct": true
}
```

##### JSON Response

```json
{
  "success": true,
  "word_id": 1,
  "study_session_id": 123,
  "correct": true,
  "created_at": "2024-03-20T15:31:00Z"
}
```

### Settings Endpoints

#### POST /api/settings/theme
Updates the application theme.

##### Request Params
- theme: string

##### JSON Response

```json
{
  "settings": {
    "theme": "dark"
  }
}
```

#### POST /api/settings/reset_history
Resets study history while keeping words and groups.

##### JSON Response

```json
{
  "success": true,
  "message": "Study history has been reset successfully"
}
```

#### POST /api/settings/full_reset
Performs a complete reset of the application.

##### JSON Response

```json
{
  "success": true,
  "message": "Application has been fully reset"
}
```
## Task Runner Tasks

Mage is a task runner for Go. 
Lets list out possible tasks we need for our lang portal.

### Initialize Database

This task will initialize the sqlite3 database called `words.db` 
in the root of the project folder of `backend_go`.

### Migrate Database

- This task will run a series of migrations sql files on the database.
- Migrations will live in the `migrations` folder.
- The migration files will be run in order of their file name.
- The file names should look like this:

```sql
00001_init.sql
00002_create_words_table.sql
```

### Seed Database

This task will import json files and transform them into target data for our database.


- All seed files live in the `seeds` folder.
- All seed files should be loaded.
- In our task we should have a DSL to specify each seed file and its expected group word name.

```json
[
  {
    "kanji": "払う",
    "romaji": "harau",
    "english": "to pay"
  },
  ...
]
```