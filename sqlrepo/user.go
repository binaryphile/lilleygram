package sqlrepo

import (
	"database/sql"
	"errors"
	"github.com/binaryphile/lilleygram/model"
	. "github.com/binaryphile/lilleygram/shortcuts"
)

type (
	UserRepo struct {
		db  *sql.DB
		now func() int64
	}

	fnTime = func() int64
)

func NewUserRepo(db *sql.DB, now fnTime) UserRepo {
	return UserRepo{
		db:  db,
		now: now,
	}
}

func (r UserRepo) Add(firstName, lastName, userName, avatar string) (zero uint64, err error) {
	stmt := Heredoc(`
		INSERT INTO users (avatar, first_name, last_name, user_name)
		VALUES ($1, $2, $3, $4)
	`)

	result, err := r.db.Exec(stmt, avatar, firstName, lastName, userName)
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

func (r UserRepo) Get(id uint64) (_ model.User, err error) {
	stmt := Heredoc(`
		SELECT users.user_id, avatar, first_name, last_name, user_name FROM users
		INNER JOIN certificates ON users.user_id = certificates.user_id
		WHERE users.user_id = $1
	`)

	var u model.User

	err = r.db.QueryRow(stmt, id).Scan(&u.ID, &u.Avatar, &u.FirstName, &u.LastName, &u.UserName)
	if err != nil {
		return
	}

	return u, nil
}

func (r UserRepo) GetByCertificate(certSHA256 string) (_ model.User, err error) {
	stmt := Heredoc(`
		SELECT users.user_id, first_name, last_name, user_name, avatar FROM users
		INNER JOIN certificates ON users.user_id = certificates.user_id
		WHERE cert_sha256 = $1
	`)

	var u model.User

	err = r.db.QueryRow(stmt, certSHA256).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.UserName, &u.Avatar)
	if err != nil {
		return
	}

	return u, nil
}

func (r UserRepo) PasswordAdd(userID uint64, password model.Password) error {
	stmt := Heredoc(`
		INSERT INTO passwords (user_id, argon2, salt) 
		VALUES ($1, $2, $3)
	`)

	_, err := r.db.Exec(stmt, userID, password.Argon2, password.Salt)

	return err
}

func (r UserRepo) PasswordGet(userID uint64) (_ model.Password, err error) {
	stmt := Heredoc(`
		SELECT argon2, salt, created_at, updated_at
		FROM passwords
		WHERE user_id = $1
	`)

	var p model.Password

	err = r.db.QueryRow(stmt, userID).Scan(&p.Argon2, &p.CreatedAt, &p.Salt)
	if err != nil {
		return
	}

	return p, nil
}

func (r UserRepo) PasswordUpdate(userID uint64, password model.Password) error {
	stmt := Heredoc(`
		UPDATE passwords
		SET argon2 = $1, salt = $2, updated_at = $3
		WHERE user_id = $4
	`)

	_, err := r.db.Exec(stmt, password.Argon2, password.Salt, r.now(), userID)

	return err
}
