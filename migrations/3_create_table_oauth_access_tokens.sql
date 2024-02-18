-- +goose Up

CREATE TABLE IF NOT EXISTS oauth_access_tokens
(
    id         TEXT PRIMARY KEY,
    user_id    BIGINT           DEFAULT NULL,
    client_id  TEXT    NOT NULL,
    name       TEXT             DEFAULT NULL,
    scopes     TEXT    NOT NULL DEFAULT '[]',
    revoked    BOOLEAN NOT NULL DEFAULT false,
    created_at INT              DEFAULT 0,
    updated_at INT              DEFAULT 0,
    expires_at INT              DEFAULT 0
);

CREATE INDEX oauth_clients_user_id_index ON oauth_access_tokens (user_id);

-- +goose Down

DROP TABLE IF EXISTS oauth_access_tokens;
