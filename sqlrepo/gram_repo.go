package sqlrepo

import (
	"github.com/binaryphile/lilleygram/model"
	. "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

type (
	GramRepo struct {
		db  *Database
		now func() int64
		tx  *TxDatabase
	}

	Fromer interface {
		From(...any) *SelectDataset
	}
)

func NewGramRepo(db *Database, now fnTime) GramRepo {
	return GramRepo{
		db:  db,
		now: now,
	}
}

func (r GramRepo) Add(userID uint64, gram string) (_ uint64, err error) {
	db := ifThenElse[Inserter](r.tx != nil, r.tx, r.db)

	query := db.
		Insert("grams").
		Rows(
			Record{"user_id": userID, "gram": gram},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return
	}

	gramID, err := result.LastInsertId()
	if err != nil {
		return
	}

	return uint64(gramID), nil
}

func (r GramRepo) List(userID uint64) (_ []model.Gram, err error) {
	db := ifThenElse[Fromer](r.tx != nil, r.tx, r.db)

	grams := make([]model.Gram, 0, 25)

	query := db.
		From("grams").
		Join(
			T("follows"),
			On(Ex{"follows.follow_id": I("grams.user_id")}),
		).
		Join(
			T("users"),
			On(Ex{"grams.user_id": I("users.id")}),
		).
		Where(
			Ex{"follows.user_id": userID},
		)

	err = query.ScanStructs(&grams)
	if err != nil {
		return
	}

	return grams, nil
}

func (r GramRepo) Commit() (_ error) {
	if r.tx != nil {
		return r.tx.Commit()
	}

	return
}

func (r GramRepo) Rollback() (_ error) {
	if r.tx != nil {
		return r.tx.Rollback()
	}

	return
}

func (r GramRepo) Begin() (_ GramRepo, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}

	return GramRepo{
		db:  r.db,
		now: r.now,
		tx:  tx,
	}, nil
}

// WithTx starts a new transaction and executes it in Wrap method
func (r GramRepo) WithTx(fn func(GramRepo) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	repo := GramRepo{
		db:  r.db,
		now: r.now,
		tx:  tx,
	}

	return tx.Wrap(
		func() error {
			return fn(repo)
		},
	)
}
