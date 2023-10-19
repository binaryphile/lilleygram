CREATE TABLE sparkles
(
    id         INTEGER NOT NULL PRIMARY KEY,
    expire_at  INTEGER NOT NULL DEFAULT 0,
    gram_id    INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (gram_id) REFERENCES grams (id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);
