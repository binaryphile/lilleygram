package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/a-h/gemini"
	"github.com/binaryphile/lilleygram/controller"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	// open the database

	db, closeDB := openSQL()
	defer closeDB()

	// create controllers

	certAuthorizer := newCertAuthorizer(db)

	userRepo := sqlrepo.NewUserRepo(db)

	withAuthentication := WithAuthentication(userRepo, certAuthorizer)

	userController := controller.NewUserController(userRepo, withAuthentication)

	certificateRepo := sqlrepo.NewCertificateRepo(db)

	certificateController := controller.NewCertificateController(certificateRepo, withAuthentication)

	homeController := controller.NewHomeController(WithOptionalAuthentication(userRepo))

	// set up the domain handler

	ctx := context.Background()

	certificate := tlsmust.LoadX509KeyPair(osmust.Getenv("LGRAM_X509_CERT_FILE"), osmust.Getenv("LGRAM_X509_KEY_FILE"))

	router := controller.Router(certificateController, homeController, userController)

	domainHandler := gemini.NewDomainHandler(osmust.Getenv("LGRAM_SERVER_NAME"), certificate, router)

	// Start the server
	err := gemini.ListenAndServe(ctx, ":1965", domainHandler)
	if err != nil {
		log.Fatal("error:", err)
	}
}

func newCertAuthorizer(db *sql.DB) func(_, _ string) bool {
	return func(certID, _ string) bool {
		hash := sha256.Sum256([]byte(certID))

		certSHA256 := hex.EncodeToString(hash[:])

		var count int

		err := db.QueryRow(heredoc.Doc(`
				SELECT count(*) FROM users
				INNER JOIN certificates ON users.user_id = certificates.user_id
				WHERE cert_sha256 = $1
			`), certSHA256).Scan(&count)
		if err != nil {
			log.Panicf("couldn't query users: %s", err)
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
