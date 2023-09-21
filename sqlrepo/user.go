package sqlrepo

import (
	"errors"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/view"
	. "github.com/doug-martin/goqu/v9"

	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

type (
	UserRepo struct {
		db  *Database
		now func() int64
	}

	fnTime = func() int64
)

func NewUserRepo(db *Database, now fnTime) UserRepo {
	return UserRepo{
		db:  db,
		now: now,
	}
}

func (r UserRepo) CertificateAdd(sha256 string, expireAt int64, userID uint64) (_ string, err error) {
	insert := r.db.From("certificates").Insert().Rows(
		Record{"cert_sha256": sha256, "expire_at": expireAt, "user_id": userID},
	).Executor()

	if _, err = insert.Exec(); err != nil {
		return
	}

	return sha256, nil
}

func (r UserRepo) CertificateListByUser(userID uint64) (_ []view.Certificate, err error) {
	certificates := make([]view.Certificate, 0)

	if err = r.db.From("certificates").Where(Ex{"user_id": userID}).ScanStructs(&certificates); err != nil {
		panic(err)
	}

	return certificates, nil
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

func (r UserRepo) Get(id uint64) (_ view.User, err error) {
	stmt := Heredoc(`
		SELECT users.user_id, avatar, first_name, last_name, user_name FROM users
		INNER JOIN certificates ON users.user_id = certificates.user_id
		WHERE users.user_id = $1
	`)

	var u view.User

	err = r.db.QueryRow(stmt, id).Scan(&u.ID, &u.Avatar, &u.FirstName, &u.LastName, &u.UserName)
	if err != nil {
		return
	}

	return u, nil
}

func (r UserRepo) GetByCertificate(certSHA256 string) (_ view.User, err error) {
	stmt := Heredoc(`
		SELECT users.user_id, first_name, last_name, user_name, avatar FROM users
		INNER JOIN certificates ON users.user_id = certificates.user_id
		WHERE cert_sha256 = $1
	`)

	var u view.User

	err = r.db.QueryRow(stmt, certSHA256).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.UserName, &u.Avatar)
	if err != nil {
		return
	}

	return u, nil
}

func (r UserRepo) PasswordAdd(userID uint64, password view.Password) error {
	update := r.db.Update("passwords").
		Where(Ex{"user_id": userID}).
		Set(
			Record{"argon2": password.Argon2, "salt": password.Salt, "updated_at": r.now()},
		).Executor()

	res, err := update.Exec()
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		insert := r.db.Insert("passwords").
			Rows(
				Record{"user_id": userID, "argon2": password.Argon2, "salt": password.Salt},
			).Executor()

		if _, err = insert.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r UserRepo) PasswordGet(userID uint64) (_ view.Password, err error) {
	stmt := Heredoc(`
		SELECT argon2, salt, created_at, updated_at
		FROM passwords
		WHERE user_id = $1
	`)

	var p view.Password

	err = r.db.QueryRow(stmt, userID).Scan(&p.Argon2, &p.CreatedAt, &p.Salt)
	if err != nil {
		return
	}

	return p, nil
}

func (r UserRepo) PasswordUpdate(userID uint64, password view.Password) error {
	stmt := Heredoc(`
		UPDATE passwords
		SET argon2 = $1, salt = $2, updated_at = $3
		WHERE user_id = $4
	`)

	_, err := r.db.Exec(stmt, password.Argon2, password.Salt, r.now(), userID)

	return err
}
