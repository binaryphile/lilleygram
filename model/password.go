package model

type Password struct {
	Argon2    string
	Salt      string
	CreatedAt int64
	UpdatedAt int64
}
