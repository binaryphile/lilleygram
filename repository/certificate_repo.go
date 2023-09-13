package repository

import (
	"database/sql"
	. "github.com/MakeNowJust/heredoc/v2"
	"github.com/binaryphile/lilleygram/opt"
	"log"
	"time"
)

type (
	Certificate struct {
		CertSHA256 string
		CreatedAt  string
		Expiry     string
	}

	CertificateRepo struct {
		db *sql.DB
	}
)

func NewCertificateRepo(db *sql.DB) CertificateRepo {
	return CertificateRepo{
		db: db,
	}
}

func (r CertificateRepo) Add(sha256 string, expiry, userID uint64) (_ string, err error) {
	stmt := Doc(`
		INSERT INTO certificates (cert_sha256, expiry, user_id)
		VALUES (?, ?, ?)
	`)

	_, err = r.db.Exec(stmt, sha256, expiry, userID)
	if err != nil {
		return
	}

	return sha256, nil
}

func (r CertificateRepo) ListByUser(userID uint64) (_ []Certificate, err error) {
	stmt := Doc(`
		SELECT cert_sha256, created_at, expiry
		FROM certificates
		WHERE user_id = ?
	`)

	rows, err := r.db.Query(stmt, userID)
	if err != nil {
		return
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	var certificates []Certificate

	for rows.Next() {
		var certSHA256 string

		var createdAt, expiry uint64

		err = rows.Scan(&certSHA256, &createdAt, &expiry)
		if err != nil {
			return
		}

		toString := opt.Map(func(unixtime int64) string {
			return time.Unix(unixtime, 0).String()
		})

		certificates = append(certificates, Certificate{
			CertSHA256: certSHA256,
			CreatedAt:  time.Unix(int64(createdAt), 0).Format(time.RFC822),
			Expiry:     toString(opt.OfNonZero(int64(expiry))).Or("never"),
		})
	}

	return certificates, nil
}
