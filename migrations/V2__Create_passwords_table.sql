CREATE TABLE passwords
(
    user_id     INTEGER NOT NULL PRIMARY KEY ,
    argon2      TEXT    NOT NULL,
    salt        TEXT    NOT NULL UNIQUE,
    created_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);