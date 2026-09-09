-- +goose Up

-- +goose StatementBegin
CREATE TABLE job_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL UNIQUE,
    output TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key linking job_id to jobs(id).
    CONSTRAINT fk_jobs_result
        FOREIGN KEY (job_id)
        REFERENCES jobs(id)
        ON DELETE CASCADE
);

-- Create an index on job_id.
CREATE INDEX idx_job_results_job_id ON job_results(job_id);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS job_results CASCADE;

-- +goose StatementEnd
