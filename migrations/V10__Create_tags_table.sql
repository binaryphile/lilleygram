CREATE TABLE tags
(
    body       INTEGER NOT NULL COLLATE NOCASE,
    gram_id    INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (body, gram_id),
    FOREIGN KEY (gram_id) REFERENCES grams (id)
);
