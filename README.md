# Job-Queue
Software that completes user jobs using workers.

# Architecture

client --> API server --> database <-- worker service

The client makes a request regarding a job. The database is updated and the API sends its response. The worker service updates its jobs queue, assigns it to workers, and marks the job as completed in the database.

## Database

```mermaid
erDiagram
  USERS ||--o{ JOBS : creates
  USERS {
      uuid id PK
      string first_name
      string last_name
      string username "Unique"
      string email "Unique"
      string password
      timestamp updated_at
      timestamp created_at
  }
  JOBS {
      uuid id PK
      uuid author_id FK
      enum language "['javascript', 'python']"
      string dependencies
      string code
      enum status "['idle', 'running', 'completed']"
      timestamp updated_at
      timestamp created_at
  }
```
