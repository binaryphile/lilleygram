package sqlrepo

import (
	"database/sql"
)

type (
	CertificateRepo struct {
		db *sql.DB
	}
)

func NewCertificateRepo(db *sql.DB) CertificateRepo {
	return CertificateRepo{
		db: db,
	}
}
