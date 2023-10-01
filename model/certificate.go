package model

import (
	"github.com/dustin/go-humanize"
	"time"
)

type Certificate struct {
	SHA256    string `db:"cert_sha256"`
	ExpireAt  int64  `db:"expire_at"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func (c Certificate) GetCreatedAt() string {
	return humanTime(c.CreatedAt)
}

func (c Certificate) GetExpireAt() string {
	if c.ExpireAt == 0 {
		return "never"
	}

	return humanTime(c.ExpireAt)
}

func (c Certificate) GetSHA256() string {
	return c.SHA256
}

func (c Certificate) GetUpdatedAt() string {
	return humanTime(c.UpdatedAt)
}

func humanTime(unixTime int64) string {
	unix := time.Unix(unixTime, 0)

	if time.Since(unix) > 48*time.Hour {
		return unix.Format("02 Jan 2006 03:04PM")
	}

	return humanize.Time(unix)
}
