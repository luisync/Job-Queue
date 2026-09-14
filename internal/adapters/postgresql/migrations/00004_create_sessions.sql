-- +goose Up

-- +goose StatementBegin
CREATE TABLE sessions (
    id TEXT PRIMARY KEY NOT NULL,
    user_email TEXT NOT NULL,
    refresh_token VARCHAR(512) NOT NULL,
    is_revoked bool NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Foreign key linking user_email to users(email).
    CONSTRAINT fk_user_email
        FOREIGN KEY (user_email)
        REFERENCES users(email)
        ON DELETE CASCADE
);

-- Create an index on user_email and refresh_tokens.
CREATE INDEX idx_sessions_user_email ON sessions(user_email);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS sessions CASCADE;

-- +goose StatementEnd
