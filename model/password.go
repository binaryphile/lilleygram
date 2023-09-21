package model

import (
	"encoding/base64"
	"github.com/binaryphile/lilleygram/hash"
)

type Password struct {
	UserID    uint64 `db:"user_id"`
	Argon2    string `db:"argon2"`
	Salt      string `db:"salt"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func NewPassword(password string) Password {
	salt := hash.GenerateSalt()

	return Password{
		Argon2: hash.HashPassword(password, salt),
		Salt:   base64.RawStdEncoding.EncodeToString(salt),
	}
}
