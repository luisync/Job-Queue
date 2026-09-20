-- Jobs

-- name: ListJobs :many
SELECT *
FROM jobs 
WHERE creator_id = $1;

-- name: FindJobByID :one
SELECT * 
FROM jobs 
WHERE 
id = $1
AND 
creator_id = $2;

-- name: CreateJob :one
INSERT INTO jobs (
    creator_id, 
    language, 
    dependencies, 
    function
)
VALUES (
    $1, 
    $2, 
    $3, 
    $4
) 
RETURNING *;

-- Users

-- name: FindUserByEmail :one
SELECT * 
FROM users 
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
    first_name, 
    last_name, 
    username, 
    password,
    email
)
VALUES (
    $1, 
    $2, 
    $3, 
    $4,
    $5
) 
RETURNING *;

-- Sessions

-- name: FindSessionByID :one
SELECT * 
FROM sessions 
WHERE id = $1;

-- name: FindSessionsByEmail :many
SELECT * 
FROM sessions 
WHERE user_email = $1;

-- name: RevokeSessions :many
UPDATE sessions 
SET is_revoked=true
WHERE user_email = $1
RETURNING *;

-- name: DeteleSessions :many
DELETE 
FROM sessions 
WHERE user_email = $1
RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (
    id,
    user_email, 
    refresh_token,
    is_revoked,
    expires_at
)
VALUES (
    $1, 
    $2, 
    $3,
    $4,
    $5
) 
RETURNING *;

-- Private functions for workers.

-- name: WorkerCreateJobResult :one
INSERT INTO job_results (
    job_id, 
    output
)
VALUES (
    $1, 
    $2
) 
RETURNING *;

-- name: WorkerFindJobByID :one
SELECT dependencies, language, function
FROM jobs 
WHERE id = $1;

-- name: WorkerUpdateStatus :one
UPDATE jobs 
SET status=$1
WHERE id = $2
RETURNING *;