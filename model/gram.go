package model

type Gram struct {
	ID        uint64 `db:"id"`
	Avatar    string `db:"avatar"`
	Body      string `db:"body"`
	ExpireAt  int64  `db:"expire_at"`
	Sparkles  int    `db:"sparkles"`
	UserID    uint64 `db:"user_id"`
	UserName  string `db:"user_name"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}
