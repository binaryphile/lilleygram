package model

type Gram struct {
	ID             uint64 `db:"id"`
	AuthorAvatar   string `db:"avatar"`
	AuthorID       uint64 `db:"user_id"`
	AuthorUserName string `db:"user_name"`
	Body           string `db:"body"`
	ExpireAt       int64  `db:"expire_at"`
	LikedByTracked bool   `db:"liked_by_tracked"`
	Sparkles       int    `db:"sparkles"`
	TrackingAuthor bool   `db:"tracking_author"`
	CreatedAt      int64  `db:"created_at"`
	UpdatedAt      int64  `db:"updated_at"`
}
