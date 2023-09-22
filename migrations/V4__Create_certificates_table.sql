CREATE TABLE certificates
(
    cert_sha256 TEXT    NOT NULL PRIMARY KEY,
    expire_at   INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    user_id     INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);