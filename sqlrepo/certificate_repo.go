package sqlrepo

import (
	"database/sql"
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/opt"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"log"
	"time"
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

func (r CertificateRepo) Add(sha256 string, expiry int64, userID uint64) (_ string, err error) {
	stmt := Heredoc(`
		INSERT INTO certificates (cert_sha256, expiry, user_id)
		VALUES ($1, $2, $3)
	`)

	_, err = r.db.Exec(stmt, sha256, expiry, userID)
	if err != nil {
		return
	}

	return sha256, nil
}

func (r CertificateRepo) ListByUser(userID uint64) (_ []model.Certificate, err error) {
	stmt := Heredoc(`
		SELECT cert_sha256, created_at, expiry
		FROM certificates
		WHERE user_id = $1
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

	var certificates []model.Certificate

	for rows.Next() {
		var certSHA256 string

		var createdAt, exp int64

		err = rows.Scan(&certSHA256, &createdAt, &exp)
		if err != nil {
			return
		}

		expiry := opt.OfNonZero(exp)

		toString := opt.Map(func(unixtime int64) string {
			return time.Unix(unixtime, 0).Format(time.RFC1123)
		})

		certificates = append(certificates, model.Certificate{
			CertSHA256: certSHA256,
			CreatedAt:  time.Unix(createdAt, 0).Format(time.RFC1123),
			Expiry:     toString(expiry).Or("never"),
		})
	}

	return certificates, nil
}
