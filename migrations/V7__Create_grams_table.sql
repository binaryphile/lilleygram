CREATE TABLE grams
(
    id         INTEGER NOT NULL PRIMARY KEY,
    expire_at  INTEGER NOT NULL DEFAULT 0,
    gram       TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    user_id    INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
);
