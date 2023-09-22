package model

import (
	"github.com/argoproj/pkg/humanize"
	"time"
)

type Certificate struct {
	CreatedAt int64  `db:"created_at"`
	ExpireAt  int64  `db:"expire_at"`
	SHA256    string `db:"cert_sha256"`
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

	if time.Since(unix) > 168*time.Hour {
		return unix.Format("02 Jan 2006 15:04")
	}

	return humanize.RelativeDuration(time.Now(), unix)
}
