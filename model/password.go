package model

import (
	"encoding/base64"
	"github.com/binaryphile/lilleygram/hash"
	"unicode"
)

type Password struct {
	UserID    uint64 `db:"user_id"`
	Argon2    string `db:"argon2"`
	Salt      string `db:"salt"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func NewPassword(password string) (_ Password, length, upper, lower, digit, special bool) {
	length, upper, lower, digit, special = isPasswordComplex(password)
	if !(length && upper && lower && digit && special) {
		return
	}

	salt := hash.GenerateSalt()

	return Password{
		Argon2: hash.HashPassword(password, salt),
		Salt:   base64.RawStdEncoding.EncodeToString(salt),
	}, true, true, true, true, true
}

func isPasswordComplex(password string) (length, upper, lower, digit, special bool) {
	length = len(password) >= 8

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			special = true
		}
	}

	return
}
