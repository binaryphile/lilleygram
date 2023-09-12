package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/binaryphile/lilleygram/controller"
	. "github.com/binaryphile/lilleygram/extensions"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"log"

	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"

	_ "modernc.org/sqlite"
)

func main() {
	// open the database

	db, err := sql.Open("sqlite", osmust.Getenv("GMNI_SQLITE_FILE"))
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

	siteController := controller.New(db)

	// Extend the controller Home method with authentication
	homeHandler := gemini.HandlerFunc(ExtendFnHandler(
		siteController.Home,
		WithOptionalAuthentication(db),
	))

	userHandler := gemini.HandlerFunc(ExtendFnHandler(
		siteController.Users,
		WithAuthentication(db, func(certID, _ string) bool {
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
		}),
	))

	// create a router

	router := mux.NewMux()

	router.AddRoute("/", homeHandler)

	router.AddRoute("/users", userHandler)

	// set up the domain handler

	ctx := context.Background()

	certificate := tlsmust.LoadX509KeyPair(osmust.Getenv("GMNI_X509_CERT_FILE"), osmust.Getenv("GMNI_X509_KEY_FILE"))

	domainHandler := gemini.NewDomainHandler(osmust.Getenv("GMNI_SERVER_NAME"), certificate, router)

	// Start the server
	err = gemini.ListenAndServe(ctx, ":1965", domainHandler)
	if err != nil {
		log.Fatal("error:", err)
	}
}
