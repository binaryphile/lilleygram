package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/a-h/gemini"
	"github.com/binaryphile/lilleygram/controller"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/sqlrepo"
	. "github.com/doug-martin/goqu/v9"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	// open the database
	sqlDB, closeDB := openSQL()
	defer closeDB()

	db := New("sqlite3", sqlDB)

	// create controllers

	userRepo := sqlrepo.NewUserRepo(db, unixNow)

	userController := controller.NewUserController(userRepo)

	userRouter := ExtendRouter(userController.Router(), WithAuthentication(userRepo, newCertAuthorizer(db)))

	homeController := controller.NewHomeController()

	homeRouter := ExtendRouter(homeController.Router(), WithOptionalAuthentication(userRepo))

	// set up the domain handler

	root := mountRouters(map[string]Handler{
		"/":      homeRouter,
		"/users": userRouter,
	})

	certificate := tlsmust.LoadX509KeyPair(osmust.Getenv("LGRAM_X509_CERT_FILE"), osmust.Getenv("LGRAM_X509_KEY_FILE"))

	address := opt.Getenv("LGRAM_SERVER_ADDRESS").Or("g.lilleygram.com:1965")

	host, _, _ := strings.Cut(address, ":")

	domainHandler := gemini.NewDomainHandler(host, certificate, root)

	// Start the server
	err := gemini.ListenAndServe(context.Background(), address, domainHandler)
	if err != nil {
		log.Panic(err)
	}
}

func mountRouters(handlers map[string]Handler) HandlerFunc {
	routes := make(map[string]Handler)

	for pattern, handler := range handlers {
		pattern = strings.TrimPrefix(pattern, "/")

		routes[pattern] = handler
	}

	return func(writer ResponseWriter, request *Request) {
		path := strings.TrimPrefix(request.URL.Path, "/")

		first, _, _ := strings.Cut(path, "/")

		if handler, ok := routes[first]; ok {
			handler.ServeGemini(writer, request)
		} else {
			gemini.NotFound(writer, request)
		}
	}
}

func newCertAuthorizer(db *Database) func(_, _ string) bool {
	return func(certID, _ string) bool {
		hash := sha256.Sum256([]byte(certID))

		certSHA256 := hex.EncodeToString(hash[:])

		var count int

		_, err := db.From("users").
			Select(
				COUNT("*"),
			).
			Join(T("certificates"), On(
				Ex{"users.user_id": I("certificates.user_id")},
			)).
			Where(Ex{"cert_sha256": certSHA256}).
			ScanVal(&count)
		if err != nil {
			return false
		}

		return count > 0
	}
}

func openSQL() (db *sql.DB, cleanup func()) {
	db, err := sql.Open("sqlite", osmust.Getenv("LGRAM_SQLITE_FILE"))
	if err != nil {
		log.Fatalf("couldn't open sql db: %s", err)
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
