CREATE TABLE emails
(
    address    TEXT    NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    user_id    INTEGER NOT NULL,
    PRIMARY KEY (user_id, address),
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);