CREATE TABLE follows
(
    user_id    INTEGER NOT NULL,
    follow_id  INTEGER NOT NULL,
    expire_at  INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (user_id, follow_id),
    FOREIGN KEY (follow_id) REFERENCES users (id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);
