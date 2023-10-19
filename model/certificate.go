package model

type Certificate struct {
	SHA256    string `db:"cert_sha256"`
	ExpireAt  int64  `db:"expire_at"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func (c Certificate) GetCreatedAt() string {
	return HumanTime(c.CreatedAt)
}

func (c Certificate) GetExpireAt() string {
	if c.ExpireAt == 0 {
		return "never"
	}

	return HumanTime(c.ExpireAt)
}

func (c Certificate) GetSHA256() string {
	return c.SHA256
}

func (c Certificate) GetUpdatedAt() string {
	return HumanTime(c.UpdatedAt)
}
