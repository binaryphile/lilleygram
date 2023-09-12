INSERT INTO users (first_name, last_name, user_name, avatar, start, stop)
VALUES ('Ted', 'Lilley', 'KingDad', '👑', unixepoch(), 0);

INSERT INTO emails (address, user_id)
VALUES ('admin@lilleygram.com', 1);

INSERT INTO certificates (cert_sha256, user_id)
VALUES ('2ad6b0ba5c9c4fc34ce5bf67e803373169465f7f0209e87cd5f5586384175f39', 1);
