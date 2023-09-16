package controller

import (
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
)

func Router(
	certificateController CertificateController,
	homeController HomeController,
	userController UserController,
) *mux.Mux {
	router := mux.NewMux()

	fnHandlers := map[string]FnHandler{
		"/":                                homeController.Get,
		"/users":                           userController.List,
		"/users/{userID}":                  userController.Get,
		"/users/{userID}/password":         userController.PasswordGet,
		"/users/{userID}/password/add":     userController.PasswordAdd,
		"/users/{userID}/certificates":     certificateController.List,
		"/users/{userID}/certificates/add": certificateController.Add,
	}

	for pattern, fnHandler := range fnHandlers {
		router.AddRoute(pattern, HandlerFunc(fnHandler))
	}

	return router
}
