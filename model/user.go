package model

type User struct {
	ID        uint64 `db:"user_id"`
	Avatar    string `db:"avatar"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	UserName  string `db:"user_name"`
}
