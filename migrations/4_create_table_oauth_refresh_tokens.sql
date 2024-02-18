-- +goose Up

CREATE TABLE IF NOT EXISTS oauth_refresh_tokens
(
    id              TEXT PRIMARY KEY,
    access_token_id TEXT    NOT NULL,
    revoked         BOOLEAN NOT NULL,
    expires_at      INT DEFAULT 0
);

CREATE INDEX oauth_refresh_tokens_access_token_id_index ON oauth_refresh_tokens (access_token_id);

-- +goose Down

DROP TABLE IF EXISTS oauth_refresh_tokens;
