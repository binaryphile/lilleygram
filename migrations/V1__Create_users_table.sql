CREATE TABLE users
(
    id         INTEGER NOT NULL PRIMARY KEY,
    avatar     TEXT    NOT NULL,
    expire_at  INTEGER NOT NULL DEFAULT 0,
    first_name TEXT    NOT NULL,
    last_name  TEXT    NOT NULL,
    last_seen  INTEGER NOT NULL DEFAULT 0,
    user_name  TEXT    NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
