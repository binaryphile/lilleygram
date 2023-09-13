CREATE TABLE users
(
    user_id    INTEGER NOT NULL PRIMARY KEY,
    first_name TEXT    NOT NULL,
    last_name  TEXT    NOT NULL,
    user_name  TEXT    NOT NULL,
    avatar     TEXT    NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    expiry     INTEGER NOT NULL DEFAULT 0
);