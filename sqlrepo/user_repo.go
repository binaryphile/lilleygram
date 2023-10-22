package sqlrepo

import (
	"database/sql"
	"errors"
	"github.com/binaryphile/lilleygram/model"
	. "github.com/binaryphile/lilleygram/shortcuts"
	. "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

const unimplemented = "unimplemented"

type (
	UserRepo struct {
		db  *Database
		now func() int64
		tx  *TxDatabase
	}

	DB interface {
		From(...any) *SelectDataset
		Insert(any) *InsertDataset
		Update(any) *UpdateDataset
	}

	fnTime = func() int64
)

func NewUserRepo(db *Database, now fnTime) UserRepo {
	return UserRepo{
		db:  db,
		now: now,
	}
}

func (r UserRepo) Add(firstName, lastName, userName, avatar string) (_ uint64, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Insert("users").
		Rows(
			Record{"avatar": avatar, "first_name": firstName, "last_name": lastName, "user_name": userName},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return
	}

	return uint64(userID), nil
}

func (r UserRepo) Begin() (_ UserRepo, err error) {
	if r.tx != nil {
		panic(unimplemented)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return
	}

	return UserRepo{
		db:  nil,
		now: r.now,
		tx:  tx,
	}, nil
}

func (r UserRepo) BeginTx(ctx Context, opts *sql.TxOptions) (_ UserRepo, err error) {
	if r.tx != nil {
		panic(unimplemented)
	}

	tx, err := r.db.BeginTx(ctx, opts)
	if err != nil {
		return
	}

	return UserRepo{
		db:  nil,
		now: r.now,
		tx:  tx,
	}, nil
}

func (r UserRepo) CertificateAdd(sha256 string, expireAt int64, userID uint64) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Insert("certificates").
		Rows(
			Record{"cert_sha256": sha256, "expire_at": expireAt, "user_id": userID},
		)

	_, err := query.Executor().Exec()

	return err
}

func (r UserRepo) CertificateListByUser(userID uint64) (_ []model.Certificate, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	certificates := make([]model.Certificate, 0)

	query := db.
		From("certificates").
		Where(Ex{"user_id": userID})

	err = query.ScanStructs(&certificates)
	if err != nil {
		return
	}

	return certificates, nil
}

func (r UserRepo) CodeGet(_ uint64) (_ string, found bool, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	var code string

	query := db.
		From("registration").
		Select("code").
		Limit(1)

	if found, err = query.ScanVal(&code); err != nil || !found {
		return
	}

	return code, true, nil
}

func (r UserRepo) Get(id uint64) (_ model.User, found bool, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	var u model.User

	query := db.
		From("users").
		Join(
			T("certificates"),
			On(Ex{"users.id": I("certificates.user_id")}),
		).
		Where(Ex{"users.id": id})

	if found, err = query.ScanStruct(&u); err != nil || !found {
		return
	}

	return u, true, nil
}

func (r UserRepo) GetByCertificate(certSHA256 string) (_ model.User, found bool, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		From("users").
		Join(
			T("certificates"), On(Ex{"users.id": I("certificates.user_id")}),
		).
		Where(
			Ex{"cert_sha256": certSHA256},
		)

	var u model.User

	if found, err = query.ScanStruct(&u); err != nil || !found {
		return
	}

	return u, true, nil
}

func (r UserRepo) GetByUserName(userName string) (_ model.User, found bool, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	var u model.User

	query := db.
		From("users").
		Where(
			Ex{"user_name": userName},
		)

	if found, err = query.ScanStruct(&u); err != nil || !found {
		return
	}

	return u, true, nil
}

func (r UserRepo) PasswordGet(userID uint64) (_ model.Password, found bool, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	var p model.Password

	query := db.
		From("passwords").
		Where(Ex{"user_id": userID})

	if found, err = query.ScanStruct(&p); err != nil || !found {
		return
	}

	p.UserID = userID

	return p, true, nil
}

func (r UserRepo) PasswordSet(userID uint64, password model.Password) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Update("passwords").
		Where(Ex{"user_id": userID}).
		Set(
			Record{"argon2": password.Argon2, "salt": password.Salt, "updated_at": r.now()},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		query := db.
			Insert("passwords").
			Rows(
				Record{"user_id": userID, "argon2": password.Argon2, "salt": password.Salt},
			)

		if _, err = query.Executor().Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r UserRepo) ProfileGet(userID uint64) (_ model.Profile, _ []model.Certificate, found bool, err error) {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	var profile model.Profile

	query := db.
		From("users").
		LeftOuterJoin(
			T("passwords"), On(Ex{"users.id": I("passwords.user_id")}),
		).
		Where(Ex{"id": userID})

	if found, err = query.ScanStruct(&profile); err != nil || !found {
		return
	}

	certificates, err := r.CertificateListByUser(userID)
	if err != nil {
		return
	}

	return profile, certificates, true, nil
}

func (r UserRepo) UpdateAvatar(userID uint64, avatar string) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Update("users").
		Where(Ex{"id": userID}).
		Set(
			Record{"avatar": avatar, "updated_at": r.now()},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (r UserRepo) UpdateFirstName(userID uint64, firstName string) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Update("users").
		Where(Ex{"id": userID}).
		Set(
			Record{"first_name": firstName, "updated_at": r.now()},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (r UserRepo) UpdateLastName(userID uint64, lastName string) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Update("users").
		Where(Ex{"id": userID}).
		Set(
			Record{"last_name": lastName, "updated_at": r.now()},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (r UserRepo) UpdateSeen(userID uint64) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Update("users").
		Where(Ex{"id": userID}).
		Set(
			Record{"last_seen": r.now()},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (r UserRepo) UpdateUserName(userID uint64, userName string) error {
	db := ifThenElse[DB](r.tx != nil, r.tx, r.db)

	query := db.
		Update("users").
		Where(Ex{"id": userID}).
		Set(
			Record{"user_name": userName, "updated_at": r.now()},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

// WithTx starts a new transaction and executes it in Wrap method
func (r UserRepo) WithTx(fn func(UserRepo) error) error {
	if r.tx != nil {
		panic(unimplemented)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	repo := UserRepo{
		db:  nil,
		now: r.now,
		tx:  tx,
	}

	return tx.Wrap(
		func() error {
			return fn(repo)
		},
	)
}
