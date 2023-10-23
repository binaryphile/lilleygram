package sqlrepo

import (
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/pb"
	"github.com/binaryphile/lilleygram/slice"
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

func (r GramRepo) Discover(userID uint64, pageToken string, pageSizes ...uint32) (_ []model.Gram, _ string, err error) {
	optToken := opt.Of(pb.UnmarshalTimePageToken(pageToken), pageToken != "")

	pageSize := uint(opt.Apply((*pb.TimePageToken).GetPageSize, optToken).Or(opt.OfFirst(pageSizes).Or(25)))

	grams := make([]model.Gram, 0, pageSize+1)

	// This will return 1 row (of 1) if the gram's user is tracked by the current user, 0 otherwise
	isTracking := r.DB.
		From("tracks").
		Select(V(1).As("tracking")).
		Where(Ex{"tracker_id": userID, "tracked_id": I("g.user_id")}).
		Limit(1)

	trackedUsers := r.DB.
		From("tracks").
		Select("tracked_id").
		Where(Ex{"tracker_id": userID})

	gramsQuery := r.DB.
		From("grams").
		Select("id").
		Where(
			Or(
				Ex{"user_id": trackedUsers},
				Ex{"user_id": userID},
			),
		)

	sparklesQuery := r.DB.
		From("sparkles").
		Select("gram_id").
		Where(Ex{"user_id": trackedUsers})

	// Union the two queries
	unioned := gramsQuery.UnionAll(sparklesQuery)

	// Begin constructing the main query using the updated table names
	query := r.DB.
		From(T("combined_grams").As("cg")).
		With("combined_grams", unioned).
		Join(T("grams").As("g"), On(Ex{"cg.id": I("g.id")})).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
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

func (r GramRepo) List(userID uint64, pageToken string, pageSizes ...uint32) (_ []model.Gram, _ string, err error) {
	optToken := opt.Of(pb.UnmarshalTimePageToken(pageToken), pageToken != "")

	pageSize := uint(opt.Apply((*pb.TimePageToken).GetPageSize, optToken).Or(opt.OfFirst(pageSizes).Or(25)))

	grams := make([]model.Gram, 0, pageSize+1)

	// This will return 1 row (of 1) if the gram's user is tracked by the current user, 0 otherwise
	isTracking := r.DB.
		From("tracks").
		Select(V(1).As("tracking")).
		Where(Ex{"tracker_id": userID, "tracked_id": I("g.user_id")}).
		Limit(1)

	trackedUsers := r.DB.
		From("tracks").
		Select("tracked_id").
		Where(Ex{"tracker_id": userID})

	gramsQuery := r.DB.
		From("grams").
		Select("id").
		Where(
			Or(
				Ex{"user_id": trackedUsers},
				Ex{"user_id": userID},
			),
		)

	sparklesQuery := r.DB.
		From("sparkles").
		Select("gram_id").
		Where(Ex{"user_id": trackedUsers})

	// Union the two queries
	unioned := gramsQuery.UnionAll(sparklesQuery)

	// Begin constructing the main query using the updated table names
	query := r.DB.
		From(T("combined_grams").As("cg")).
		With("combined_grams", unioned).
		Join(T("grams").As("g"), On(Ex{"cg.id": I("g.id")})).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
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
		GroupBy(I("g.id")).
		Order(I("g.updated_at").Desc(), I("g.id").Desc()).
		Limit(pageSize + 1)

	if token, ok := optToken.Unpack(); ok {
		query = query.Where(
			Ex{
				"g.updated_at": Op{"lte": token.UpdatedAt},
				"g.id":         Op{"lt": token.Id},
			},
		)
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
	pageSizes ...uint32,
) (_ []model.Gram, _ string, err error) {
	optToken := opt.Of(pb.UnmarshalTimePageToken(pageToken), pageToken != "")

	pageSize := uint(opt.Apply((*pb.TimePageToken).GetPageSize, optToken).Or(opt.OfFirst(pageSizes).Or(25)))

	grams := make([]model.Gram, 0, pageSize+1)

	// This will return 1 row (of 1) if the gram's user is tracked by the current user, 0 rows otherwise
	isTracking := r.DB.
		From("tracks").
		Select(V(1).As("tracking")).
		Where(Ex{"tracker_id": userID, "tracked_id": I("g.user_id")}).
		Limit(1)

	gramsQuery := r.DB.
		From(T("grams").As("g")).
		Join(T("tags").As("t"), On(Ex{"g.id": I("t.gram_id")})).
		Select("id").
		Where(Ex{"t.body": tag})

	// Begin constructing the main query using the updated table names
	query := r.DB.
		From(T("tag_grams").As("cg")).
		With("tag_grams", gramsQuery).
		Join(T("grams").As("g"), On(Ex{"g.id": I("g.id")})).
		Join(T("users").As("u"), On(Ex{"g.user_id": I("u.id")})).
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
	db, ok := r.DB.(Beginner)
	if !ok {
		panic(unimplemented)
	}

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
