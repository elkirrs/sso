-- +goose Up

CREATE TABLE IF NOT EXISTS oauth_clients
(
    id                     UUID PRIMARY KEY,
    user_id                BIGINT DEFAULT 0,
    name                   TEXT    NOT NULL,
    secret                 TEXT    NOT NULL,
    provider               TEXT    NOT NULL,
    redirect               TEXT    NOT NULL,
    personal_access_client BOOLEAN NOT NULL,
    password_client        BOOLEAN NOT NULL,
    revoked                BOOLEAN NOT NULL,
    created_at             INT    DEFAULT 0,
    updated_at             INT    DEFAULT 0
);

CREATE UNIQUE INDEX oauth_clients_id_uniq_index ON oauth_clients (id);

-- +goose Down

DROP TABLE IF EXISTS oauth_clients;
