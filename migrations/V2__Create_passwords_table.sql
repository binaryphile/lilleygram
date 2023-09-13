CREATE TABLE passwords
(
    pw_id      INTEGER NOT NULL PRIMARY KEY,
    salt       TEXT    NOT NULL UNIQUE,
    argon2     TEXT    NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    expiry     INTEGER NOT NULL DEFAULT 0,
    user_id    INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);