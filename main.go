package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/a-h/gemini"
	"github.com/binaryphile/lilleygram/controller"
	. "github.com/binaryphile/lilleygram/extensions"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	// open the database

	db, err := sql.Open("sqlite", osmust.Getenv("LGRAM_SQLITE_FILE"))
	if err != nil {
		log.Fatalf("couldn't open sql db: %s", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("couldn't close database file: %s", err)
		}
	}()

	// create a controller

	siteController := controller.New(db, map[string][]FnHandlerExtension{
		"certificates": {WithAuthentication(db, authorizer(db))},
		"home":         {WithOptionalAuthentication(db)},
		"users":        {WithAuthentication(db, authorizer(db))},
	})

	// set up the domain handler

	ctx := context.Background()

	certificate := tlsmust.LoadX509KeyPair(osmust.Getenv("LGRAM_X509_CERT_FILE"), osmust.Getenv("LGRAM_X509_KEY_FILE"))

	router := siteController.Router()

	domainHandler := gemini.NewDomainHandler(osmust.Getenv("LGRAM_SERVER_NAME"), certificate, router)

	// Start the server
	err = gemini.ListenAndServe(ctx, ":1965", domainHandler)
	if err != nil {
		log.Fatal("error:", err)
	}
}

func authorizer(db *sql.DB) func(certID string, _ string) bool {
	return func(certID, _ string) bool {
		h := sha256.New()

		_, err := h.Write([]byte(certID))
		if err != nil {
			log.Panicf("couldn't hash certificate ID: %s", err)
		}

		hash := sha256.Sum256([]byte(certID))

		hexHash := hex.EncodeToString(hash[:])

		var count int

		err = db.QueryRow(heredoc.Doc(`
				SELECT count(*) FROM users
				INNER JOIN certificates ON users.user_id = certificates.user_id
				WHERE cert_sha256 = $1
			`), hexHash).Scan(&count)
		if err != nil {
			log.Panicf("couldn't query users: %s", err)
		}

		return count > 0
	}
}
