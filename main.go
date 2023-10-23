package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/a-h/gemini"
	"github.com/binaryphile/lilleygram/controller"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/gmnifc"
	"github.com/binaryphile/lilleygram/handler"
	"github.com/binaryphile/lilleygram/helper"
	"github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"github.com/doug-martin/goqu/v9"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	// open the database
	sqlDB, closeDB := openSQL(osmust.Getenv("LGRAM_SQLITE_FILE"))
	defer closeDB()

	db := goqu.New("sqlite3", sqlDB)

	// create controllers/routers

	userRepo := sqlrepo.NewUserRepo(db, unixNow)

	certAuthorizer := newCertAuthorizer(userRepo)

	gramController := controller.NewGramController(sqlrepo.NewGramRepo(db, unixNow))

	authenticatedBaseTemplates := []string{
		"view/layout/base.tmpl",
		"view/partial/footer.tmpl",
		"view/partial/nav.tmpl",
	}

	// authenticatedHandler operates behind required authentication.
	// the logged-in user experience is here.
	authenticatedHandler := middleware.ExtendHandler(
		mountHandlers(map[string]Handler{
			"/":                gramController,
			"/discover":        gramController,
			"/getting-started": handler.FileHandler(append([]string{"view/unauthenticated/getting-started.tmpl"}, authenticatedBaseTemplates...)...),
			"/grams":           gramController,
			"/register":        handler.FileHandler(append([]string{"view/register.tmpl"}, authenticatedBaseTemplates...)...),
			"/tags":            gramController,
			"/users":           controller.NewUserController(userRepo),
		}),
		middleware.WithRequiredAuthentication(certAuthorizer),
	)

	// unauthenticatedController operates behind optional authentication.
	// the authenticatedHandler also relies on the optional authentication
	// to identify the certificate, so the two controllers are combined
	// before being extended with optional authentication.
	unauthenticatedController := controller.NewUnauthenticatedController(userRepo)

	rootHandler := middleware.ExtendHandler(
		loginHandler(authenticatedHandler, unauthenticatedController),
		middleware.WithOptionalAuthentication(certAuthorizer),
	)

	if deployEnv, ok := opt.Getenv("DEPLOY_ENV").Unpack(); ok && deployEnv == "local" {
		rootHandler = middleware.ExtendHandler(
			rootHandler,
			middleware.WithLocalDeployEnv,
		)
	}

	// set up the domain handler

	certificate := tlsmust.LoadX509KeyPair(
		opt.Getenv("LGRAM_X509_CERT_FILE").Or("server.crt"),
		opt.Getenv("LGRAM_X509_KEY_FILE").Or("server.key"),
	)

	address := opt.Getenv("LGRAM_SERVER_ADDRESS").Or("g.lilleygram.com")

	host, port, _ := strings.Cut(address, ":")

	domainHandler := gemini.NewDomainHandler(host, certificate, rootHandler)

	port = ":" + opt.OfNonZero(port).Or("1965")

	// Start the server
	err := gemini.ListenAndServe(context.Background(), port, domainHandler)
	if err != nil {
		log.Panic(err)
	}
}

func loginHandler(authenticatedHandler, unauthenticatedHandler Handler) HandlerFunc {
	return func(writer ResponseWriter, request *Request) {
		if _, ok := middleware.CertUserFromRequest(request); ok {
			authenticatedHandler.ServeGemini(writer, request)
			return
		}

		unauthenticatedHandler.ServeGemini(writer, request)
	}
}

func mountHandlers(handlers map[string]Handler) HandlerFunc {
	return func(writer ResponseWriter, request *Request) {
		path := strings.TrimPrefix(request.URL.Path, "/")

		first, _, _ := strings.Cut(path, "/")

		if h, ok := handlers["/"+first]; ok {
			h.ServeGemini(writer, request)
		} else {
			gmnifc.NotFound(writer, request)
		}
	}
}

func newCertAuthorizer(repo sqlrepo.UserRepo) middleware.FnAuthorize {
	return func(certID, _ string) (_ helper.User, ok bool) {
		hash := sha256.Sum256([]byte(certID))

		certSHA256 := hex.EncodeToString(hash[:])

		user, found, err := repo.GetByCertificate(certSHA256)
		if err != nil || !found {
			log.Print(err)
			return
		}

		err = repo.UpdateSeen(user.ID)
		if err != nil {
			log.Print(err)
		}

		return helper.User{
			Avatar:   user.Avatar,
			UserID:   user.ID,
			UserName: user.UserName,
		}, true
	}
}

func openSQL(fileName string) (db *sql.DB, cleanup func()) {
	db, err := sql.Open("sqlite", fileName)
	if err != nil {
		log.Fatalf("couldn't open sql db: %s", err)
	}

	var rank int

	query := goqu.New("sqlite", db).
		From("flyway_schema_history").
		Select(goqu.MAX("installed_rank"))

	found, err := query.ScanVal(&rank)
	if err != nil {
		panic(err)
	}

	if !found {
		panic("flyway schema version not found")
	}

	if rank != 10 {
		panic("database out of version")
	}

	return db, func() {
		err := db.Close()
		if err != nil {
			log.Printf("couldn't close database file: %s", err)
		}
	}
}

func unixNow() int64 {
	return time.Now().Unix()
}
