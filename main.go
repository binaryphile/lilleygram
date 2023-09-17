package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	"github.com/binaryphile/lilleygram/controller"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"github.com/binaryphile/lilleygram/opt"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	// open the database

	db, closeDB := openSQL()
	defer closeDB()

	// create controllers

	userRepo := sqlrepo.NewUserRepo(db, unixNow)

	userController := controller.NewUserController(userRepo)

	userRouter := ExtendRouter(userController.Router(), WithAuthentication(userRepo, certAuthorizerWith(db)))

	homeController := controller.NewHomeController()

	homeRouter := ExtendRouter(homeController.Router(), WithOptionalAuthentication(userRepo))

	// set up the domain handler

	routes := map[string]gemini.Handler{
		"/":      homeRouter,
		"/users": userRouter,
	}

	router := mux.NewMux()

	for pattern, handler := range routes {
		router.AddRoute(pattern, handler)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	certificate := tlsmust.LoadX509KeyPair(osmust.Getenv("LGRAM_X509_CERT_FILE"), osmust.Getenv("LGRAM_X509_KEY_FILE"))

	domainHandler := gemini.NewDomainHandler(osmust.Getenv("LGRAM_SERVER_NAME"), certificate, router)

	// handle shutdown signals

	go cancelOnSignal(cancel, os.Interrupt, syscall.SIGTERM)

	// Start the server
	err := gemini.ListenAndServe(ctx, opt.Getenv("LGRAM_SERVER_ADDRESS").Or(":1965"), domainHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func cancelOnSignal(cancel context.CancelFunc, signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	<-c
	cancel()
}

func certAuthorizerWith(db *sql.DB) func(_, _ string) bool {
	return func(certID, _ string) bool {
		hash := sha256.Sum256([]byte(certID))

		certSHA256 := hex.EncodeToString(hash[:])

		var count int

		err := db.QueryRow(Heredoc(`
				SELECT count(*) FROM users
				INNER JOIN certificates ON users.user_id = certificates.user_id
				WHERE cert_sha256 = $1
			`), certSHA256).Scan(&count)
		if err != nil {
			log.Panic(err)
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
