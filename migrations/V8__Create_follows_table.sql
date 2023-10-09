CREATE TABLE follows
(
    id          INTEGER NOT NULL PRIMARY KEY,
    expire_at   INTEGER NOT NULL DEFAULT 0,
    followed_id INTEGER NOT NULL,
    follower_id INTEGER NOT NULL,
    created_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (followed_id) REFERENCES users (id),
    FOREIGN KEY (follower_id) REFERENCES users (id)
);
