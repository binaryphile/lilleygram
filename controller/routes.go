package controller

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
)

func (x Controller) Router() *mux.Mux {
	router := mux.NewMux()

	router.AddRoute("/", gemini.HandlerFunc(x.handlers["home"]))

	router.AddRoute("/users", gemini.HandlerFunc(x.handlers["users"]))

	return router
}
