package controller

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
)

func (c Controller) Router() *mux.Mux {
	router := mux.NewMux()

	router.AddRoute("/", gemini.HandlerFunc(c.handlers["home"]))

	router.AddRoute("/users", gemini.HandlerFunc(c.handlers["users"]))

	router.AddRoute("/users/{userid}/certificates", gemini.HandlerFunc(c.handlers["certificates"]))

	return router
}
