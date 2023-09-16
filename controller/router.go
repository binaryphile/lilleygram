package controller

import (
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
)

func Router(
	certificateController CertificateController,
	homeController HomeController,
	userController UserController,
) *mux.Mux {
	router := mux.NewMux()

	router.AddRoute("/", HandlerFunc(homeController.Get))

	router.AddRoute("/users", HandlerFunc(userController.List))

	router.AddRoute("/users/{userID}", HandlerFunc(userController.Get))

	router.AddRoute("/users/{userID}/certificates", HandlerFunc(certificateController.List))

	router.AddRoute("/users/{userID}/certificates/add", HandlerFunc(certificateController.Add))

	return router
}
