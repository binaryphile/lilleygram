CREATE TABLE certificates
(
    cert_sha256 TEXT    NOT NULL PRIMARY KEY,
    created_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    expiry      INTEGER NOT NULL DEFAULT 0,
    user_id     INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);