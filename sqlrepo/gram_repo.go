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

	Selector interface {
		Select(...any) *SelectDataset
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
	db := ifThenElse[Selector](r.tx != nil, r.tx, r.db)

	grams := make([]model.Gram, 0, 25)

	query := db.
		Select(
			"g.id",
			"u.avatar",
			"g.expire_at",
			"g.gram",
			COUNT("s.id").As("sparkles"),
			"g.user_id",
			"u.user_name",
			"g.created_at",
			"g.updated_at",
		).
		From(T("grams").As("g")).
		Join(
			T("users").As("u"),
			On(Ex{"g.user_id": I("u.id")}),
		).
		LeftJoin(
			T("sparkles").As("s"),
			On(Ex{"g.id": I("s.gram_id")}),
		).
		Where(
			Ex{"g.user_id": db.Select("followed_id").
				From("follows").
				Where(Ex{"follower_id": userID})},
		).
		GroupBy("g.id").
		Order(I("g.updated_at").Desc())

	err = query.ScanStructs(&grams)
	if err != nil {
		return
	}

	return grams, nil
}

func (r GramRepo) Sparkle(gramID, userID uint64) (_ uint64, err error) {
	db := ifThenElse[Inserter](r.tx != nil, r.tx, r.db)

	query := db.
		Insert("sparkles").
		Rows(
			Record{"gram_id": gramID, "user_id": userID},
		)

	result, err := query.Executor().Exec()
	if err != nil {
		return
	}

	sparkleID, err := result.LastInsertId()
	if err != nil {
		return
	}

	return uint64(sparkleID), nil
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
