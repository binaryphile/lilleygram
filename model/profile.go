package model

import (
	"database/sql"
)

type (
	Profile struct {
		Avatar    string         `db:"users.avatar"`
		FirstName string         `db:"users.first_name"`
		LastName  string         `db:"users.last_name"`
		LastSeen  int64          `db:"users.last_seen"`
		Password  sql.NullString `db:"passwords.argon2"`
		UserID    uint64         `db:"users.id"`
		UserName  string         `db:"users.user_name"`
		CreatedAt int64          `db:"users.created_at"`
	}
)
