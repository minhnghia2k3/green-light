CREATE TABLE IF NOT EXISTS users
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    name            text                        NOT NULL,
    email           citext UNIQUE               NOT NULL, -- case insensitive text -> no same email even different case
    hashed_password bytea                       NOT NULL, -- binary string
    activated       bool                        NOT NULL,
    version         integer                     NOT NULL DEFAULT 1
)
