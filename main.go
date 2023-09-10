package main

import (
	"context"
	"fmt"
	"github.com/binaryphile/lilleygram/must/osmust"
	"github.com/binaryphile/lilleygram/must/tlsmust"
	"log"

	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
)

func main() {
	// Create the handlers for a domain

	okHandler := gemini.HandlerFunc(func(w gemini.ResponseWriter, r *gemini.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("couldn't write response: %s", err)
		}
	})

	helloHandler := gemini.HandlerFunc(func(w gemini.ResponseWriter, r *gemini.Request) {
		_, err := w.Write([]byte("# Hello, user!\n"))
		if err != nil {
			log.Printf("couldn't write response: %s", err)
		}

		if r.Certificate.ID == "" {
			_, err = w.Write([]byte("You're not authenticated"))
			if err != nil {
				log.Printf("couldn't write response: %s", err)
			}

			return
		}

		_, err = w.Write([]byte(fmt.Sprintf("Certificate: %v\n", r.Certificate.ID)))
		if err != nil {
			log.Printf("couldn't write response: %s", err)
		}
	})

	// Create a router for gemini://host/require_cert and gemini://host/public

	routerA := mux.NewMux()

	// Let's make /require_cert require the client to be authenticated.
	routerA.AddRoute("/require_cert", gemini.RequireCertificateHandler(helloHandler, nil))

	routerA.AddRoute("/public", okHandler)

	// Set up the domain handlers.

	ctx := context.Background()

	certificate := tlsmust.LoadX509KeyPair(osmust.Getenv("GMNI_X509_CERT_FILE"), osmust.Getenv("GMNI_X509_KEY_FILE"))

	a := gemini.NewDomainHandler("localhost", certificate, routerA)

	// Start the server
	err := gemini.ListenAndServe(ctx, ":1965", a)
	if err != nil {
		log.Fatal("error:", err)
	}
}
