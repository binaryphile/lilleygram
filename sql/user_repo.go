package sql

import (
	"database/sql"
	"errors"
	. "github.com/MakeNowJust/heredoc/v2"
	. "github.com/binaryphile/lilleygram/model"
	. "github.com/binaryphile/lilleygram/shortcuts"
)

type (
	UserRepo struct {
		db *sql.DB
	}
)

func NewUserRepo(db *sql.DB) UserRepo {
	return UserRepo{
		db: db,
	}
}

func (r UserRepo) Add(firstName, lastName, userName, avatar string) (zero uint64, err error) {
	stmt := Doc(`
		INSERT INTO users (first_name, last_name, user_name, avatar)
		VALUES ($1, $2, $3, $4)
	`)

	result, err := r.db.Exec(stmt, firstName, lastName, userName, avatar)
	if err != nil {
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return
	}

	rowCount, err := result.RowsAffected()
	if err != nil {
		return
	}

	if rowCount != 1 {
		return zero, errors.New("unexpected row count on query result")
	}

	return uint64(userID), nil
}

func (r UserRepo) Get(id uint64) (_ User, err error) {
	var u User

	err = r.db.QueryRow(Heredoc(`
		SELECT users.user_id, first_name, last_name, user_name, avatar FROM users
		INNER JOIN certificates ON users.user_id = certificates.user_id
		WHERE users.user_id = $1
	`), id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.UserName, &u.Avatar)
	if err != nil {
		return
	}

	return u, nil
}

func (r UserRepo) GetByCertificate(certSHA256 string) (_ User, err error) {
	var u User

	err = r.db.QueryRow(Heredoc(`
		SELECT users.user_id, first_name, last_name, user_name, avatar FROM users
		INNER JOIN certificates ON users.user_id = certificates.user_id
		WHERE cert_sha256 = $1
	`), certSHA256).Scan(&u.ID, &u.FirstName, &u.LastName, &u.UserName, &u.Avatar)
	if err != nil {
		return
	}

	return u, nil
}
