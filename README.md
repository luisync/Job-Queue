# Job-Queue

An application that executes user jobs on an external server.

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
      string result
      enum status "['idle', 'running', 'completed']"
      timestamp updated_at
      timestamp created_at
  }
```

# Problem

Users want to minimize the utilisation of their machine to provide a more responsive experience to their end-users. But, system bottleneck analysis is costly and cumbersome for users and they would rather an alternative that is easier to implement and less expensive.

## How This Application Offers A Solution

This application tackles the problem by offering users a way to delegate jobs to an external machine through an API and later receive the result of its execution.
