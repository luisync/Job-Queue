-- +goose Up

-- +goose StatementBegin

-- Create ENUM types for the language the job was coded in and its status.
CREATE TYPE job_language AS ENUM ('javascript', 'python');
CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');

-- Create the jobs table
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    language job_language NOT NULL,
    dependencies TEXT NOT NULL DEFAULT '',
    function TEXT NOT NULL,
    status job_status NOT NULL DEFAULT 'pending',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key linking creator_id to users(id).
    CONSTRAINT fk_jobs_creator
        FOREIGN KEY (creator_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create an index on creator_id.
CREATE INDEX idx_jobs_creator_id ON jobs(creator_id);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS jobs CASCADE;
DROP TYPE IF EXISTS job_language;
DROP TYPE IF EXISTS job_status;

-- +goose StatementEnd