package sqlrepo

import (
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/pb"
	"github.com/binaryphile/lilleygram/slice"
	"github.com/binaryphile/lilleygram/sqlrepo/defaults"
	. "github.com/binaryphile/lilleygram/sqlrepo/shortcuts"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
)

type GramRepo struct {
	DB  DB
	now func() int64
}

func NewGramRepo(db *Database, now fnTime) GramRepo {
	return GramRepo{
		DB:  db,
		now: now,
	}
}

func (r GramRepo) Add(userID uint64, body string, tags ...string) (_ uint64, err error) {
	var gramID int64

	err = r.WithTx(func(r GramRepo) (err error) {
		query := r.DB.Insert("grams").Rows(Record{"user_id": userID, "body": body})

		result, err := query.Executor().Exec()
		if err != nil {
			return
		}

		gramID, err = result.LastInsertId()
		if err != nil {
			return
		}

		if len(tags) > 0 {
			toRecord := func(tag string) any {
				return Record{"gram_id": gramID, "body": tag}
			}

			records := slice.Map(toRecord, tags)

			query = r.DB.Insert("tags").Rows(records...)

			_, err = query.Executor().Exec()
		}

		return
	})
	if err != nil {
		return
	}

	return uint64(gramID), nil
}

func (r GramRepo) List(userID uint64, pageToken string, pageSizes ...uint32) (_ []model.Gram, _ string, err error) {
	token := pb.TimePageTokenOfNonZero(pageToken)

	pageSize := uint(token.Apply(pb.PageSize).Or(opt.OfFirst(pageSizes).Or(defaults.PageSize)))

	grams := make([]model.Gram, 0, pageSize+1)

	isTracking := r.DB.
		From("tracks").
		Select(V(1).As("tracking")).
		Where(Ex{"tracker_id": userID, "tracked_id": I("g.user_id")}).
		Limit(1)

	query := r.DB.
		From(T("grams").As("g")).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
		LeftJoin(T("tracks").As("t"), On(Ex{"u.id": I("t.tracked_id")})).
		LeftJoin(T("sparkles").As("s"), On(Ex{"g.id": I("s.gram_id")})).
		Where(ExOr{
			"t.tracker_id": userID,
			"u.id":         userID,
		}).
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
			COALESCE(isTracking, 0).As("tracking_author"),
		).
		GroupBy(I("g.id")).
		Order(I("g.updated_at").Desc(), I("g.id").Desc()).
		Limit(pageSize + 1)

	if token, ok := token.Unpack(); ok {
		query = query.Where(Ex{
			"g.updated_at": Op{"lte": token.UpdatedAt},
			"g.id":         Op{"lt": token.Id},
		})
	}

	err = query.ScanStructs(&grams)
	if err != nil {
		return
	}

	length := uint(len(grams))

	nextPageToken := ""

	if length > pageSize {
		grams = grams[:length-1]

		g := grams[length-2]

		nextPageToken = pb.NewTimePageToken(g.UpdatedAt, g.ID, uint32(pageSize)).Marshal()
	}

	return grams, nextPageToken, nil
}

func (r GramRepo) ListByTag(
	userID uint64,
	tag string,
	pageToken string,
	includeOwn bool,
	pageSizes ...uint32,
) (_ []model.Gram, _ string, err error) {
	pageSize := uint(pb.TimePageTokenOfNonZero(pageToken).Apply(pb.PageSize).Or(opt.OfFirst(pageSizes).Or(defaults.PageSize)))

	grams := make([]model.Gram, 0, pageSize+1)

	isTracking := r.DB.
		From("tracks").
		Select(V(1).As("tracking")).
		Where(Ex{"tracker_id": userID, "tracked_id": I("g.user_id")}).
		Limit(1)

	query := r.DB.
		From(T("grams").As("g")).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
		Join(T("tags").As("t"), On(Ex{"g.id": I("t.gram_id")})).
		LeftJoin(T("sparkles").As("s"), On(Ex{"g.id": I("s.gram_id")})).
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
			COALESCE(isTracking, 0).As("tracking_author"),
		).
		Where(Ex{"t.body": tag}).
		GroupBy(I("g.id")).
		Order(I("g.updated_at").Desc(), I("g.id").Desc()).
		Limit(pageSize + 1)

	if !includeOwn {
		query = query.Where(Ex{
			"u.id": Op{"neq": userID},
		})
	}
	err = query.ScanStructs(&grams)
	if err != nil {
		return
	}

	length := uint(len(grams))

	nextPageToken := ""

	if length > pageSize {
		grams = grams[:length-1]

		g := grams[length-2]

		nextPageToken = pb.NewTimePageToken(g.UpdatedAt, g.ID, uint32(pageSize)).Marshal()
	}

	return grams, nextPageToken, nil
}

func (r GramRepo) ListPublic(userID uint64, pageToken string, pageSizes ...uint32) (_ []model.Gram, _ string, err error) {
	token := pb.TimePageTokenOfNonZero(pageToken)

	pageSize := uint(token.Apply(pb.PageSize).Or(opt.OfFirst(pageSizes).Or(defaults.PageSize)))

	grams := make([]model.Gram, 0, pageSize+1)

	isTracking := r.DB.
		From("tracks").
		Select(V(1).As("tracking")).
		Where(Ex{"tracker_id": userID, "tracked_id": I("g.user_id")}).
		Limit(1)

	query := r.DB.
		From(T("grams").As("g")).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
		Join(T("tags").As("t"), On(Ex{"g.id": I("t.gram_id")})).
		LeftJoin(T("sparkles").As("s"), On(Ex{"g.id": I("s.gram_id")})).
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
			COALESCE(isTracking, 0).As("tracking_author"),
		).
		Where(Ex{"g.user_id": Op{"neq": userID}}).
		GroupBy(I("g.id")).
		Order(I("g.updated_at").Desc(), I("g.id").Desc()).
		Limit(pageSize + 1)

	err = query.ScanStructs(&grams)
	if err != nil {
		return
	}

	length := uint(len(grams))

	nextPageToken := ""

	if length > pageSize {
		grams = grams[:length-1]

		g := grams[length-2]

		nextPageToken = pb.NewTimePageToken(g.UpdatedAt, g.ID, uint32(pageSize)).Marshal()
	}

	return grams, nextPageToken, nil
}

func (r GramRepo) Sparkle(gramID, userID uint64) (_ uint64, err error) {
	query := r.DB.Insert("sparkles").Rows(Record{"gram_id": gramID, "user_id": userID})

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
	db := r.DB.(Beginner)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	repo := GramRepo{
		DB:  tx,
		now: r.now,
	}

	return tx.Wrap(
		func() error {
			return fn(repo)
		},
	)
}
