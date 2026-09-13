-- name: ListJobs :many
SELECT *
FROM jobs;

-- name: FindJobByID :one
SELECT * 
FROM jobs 
WHERE id = $1;

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
