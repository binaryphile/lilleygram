package model

type Gram struct {
	ID        uint64 `db:"grams.id"`
	Avatar    string `db:"users.avatar"`
	ExpireAt  int64  `db:"grams.expire_at"`
	Gram      string `db:"grams.gram"`
	UserID    uint64 `db:"grams.user_id"`
	UserName  string `db:"users.user_name"`
	CreatedAt int64  `db:"grams.created_at"`
	UpdatedAt int64  `db:"grams.updated_at"`
}
