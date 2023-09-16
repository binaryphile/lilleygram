CREATE TABLE emails
(
    user_id INTEGER NOT NULL PRIMARY KEY,
    address TEXT    NOT NULL,
    created_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);