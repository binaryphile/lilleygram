CREATE TABLE registration
(
    user_id    INTEGER NOT NULL PRIMARY KEY,
    code       TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (user_id) REFERENCES users (id)
);
