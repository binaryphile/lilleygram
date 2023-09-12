CREATE TABLE emails
(
    email_id INTEGER NOT NULL PRIMARY KEY,
    address  TEXT    NOT NULL UNIQUE,
    user_id  INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);