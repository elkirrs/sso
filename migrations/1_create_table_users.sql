-- +goose Up

CREATE TABLE IF NOT EXISTS users
(
    id                INTEGER PRIMARY KEY,
    uuid              UUID NOT NULL UNIQUE,
    name              TEXT NOT NULL,
    email             TEXT NOT NULL UNIQUE,
    email_verified_at INT     DEFAULT 0,
    password          TEXT NOT NULL,
    remember_token    TEXT    DEFAULT NULL,
    is_active         INT DEFAULT 0,
    created_at        INT     DEFAULT 0,
    updated_at        INT     DEFAULT 0
);

CREATE SEQUENCE IF NOT EXISTS users_id_seq;
ALTER TABLE users
    ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_uuid ON users (uuid);

-- +goose Down

DROP TABLE IF EXISTS users;
DROP SEQUENCE IF EXISTS users_id_seq;