CREATE TABLE IF NOT EXISTS tokens
(
    hash    bytea PRIMARY KEY,
    user_id bigint                      NOT NULL REFERENCES USERS ON DELETE CASCADE, -- DELETE all tokens when parent user is deleted
    expiry  timestamp(0) with time zone NOT NULL,
    scope   text                        NOT NULL
)