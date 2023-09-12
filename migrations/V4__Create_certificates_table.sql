CREATE TABLE certificates
(
    cert_id     INTEGER NOT NULL PRIMARY KEY,
    cert_sha256 TEXT    NOT NULL UNIQUE,
    user_id     INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);