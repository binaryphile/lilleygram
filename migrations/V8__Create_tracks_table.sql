CREATE TABLE tracks
(
    id         INTEGER NOT NULL PRIMARY KEY,
    expire_at  INTEGER NOT NULL DEFAULT 0,
    tracked_id INTEGER NOT NULL,
    tracker_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(tracked_id, tracker_id),
    FOREIGN KEY (tracked_id) REFERENCES users (id),
    FOREIGN KEY (tracker_id) REFERENCES users (id)
);
