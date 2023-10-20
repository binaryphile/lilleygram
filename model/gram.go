package model

type Gram struct {
	ID              uint64 `db:"id"`
	AuthorAvatar    string `db:"avatar"`
	AuthorID        uint64 `db:"user_id"`
	AuthorUserName  string `db:"user_name"`
	Body            string `db:"body"`
	ExpireAt        int64  `db:"expire_at"`
	FollowingAuthor bool   `db:"following_author"`
	LikedByFollowed bool   `db:"liked_by_followed"`
	Sparkles        int    `db:"sparkles"`
	CreatedAt       int64  `db:"created_at"`
	UpdatedAt       int64  `db:"updated_at"`
}
