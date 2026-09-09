-- name: ListJobs :many
SELECT *
FROM jobs;

-- name: FindJobByID :one
SELECT * 
FROM jobs 
WHERE id = $1;

