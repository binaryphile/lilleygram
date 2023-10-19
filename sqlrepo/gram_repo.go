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

func (r GramRepo) Add(userID uint64, body string) (_ uint64, err error) {
	db := ifThenElse[Inserter](r.tx != nil, r.tx, r.db)

	query := db.
		Insert("grams").
		Rows(
			Record{"user_id": userID, "body": body},
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

	// Define the subquery for users you follow
	followedUsers := db.
		From("follows").
		Select("followed_id").
		Where(Ex{"follower_id": userID})

	gramsQuery := db.
		From("grams").
		Select("id").
		Where(
			Or(
				Ex{"user_id": followedUsers},
				Ex{"user_id": userID},
			),
		)

	sparklesQuery := db.
		From("sparkles").
		Select("gram_id").
		Where(Ex{"user_id": followedUsers})

	// Union the two queries
	unioned := gramsQuery.UnionAll(sparklesQuery)

	// Begin constructing the main query using the updated table names
	query := db.
		From(T("combined_grams").As("cg")).
		With("combined_grams", unioned).
		Join(T("grams").As("g"), On(Ex{"cg.id": I("g.id")})).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
		LeftJoin(T("sparkles").As("s"), On(Ex{"cg.id": I("s.gram_id")})).
		Select(
			"g.id",
			"g.user_id",
			"u.user_name",
			"u.avatar",
			"g.body",
			COUNT(I("s.id")).As("sparkles"),
			"g.expire_at",
			"g.created_at",
			"g.updated_at",
		).
		GroupBy(I("cg.id")).
		Order(I("g.created_at").Desc())

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
