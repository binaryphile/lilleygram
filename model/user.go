package model

type User struct {
	ID        uint64 `db:"users.id"`
	Avatar    string `db:"users.avatar"`
	ExpireAt  int64  `db:"users.expire_at"`
	FirstName string `db:"users.first_name"`
	LastName  string `db:"users.last_name"`
	LastSeen  int64  `db:"users.last_seen"`
	UserName  string `db:"users.user_name"`
	CreatedAt int64  `db:"users.created_at"`
	UpdatedAt int64  `db:"users.updated_at"`
}
