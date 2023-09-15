package controller

import (
	"github.com/a-h/gemini/mux"
)

func Router(
	certificateController CertificateController,
	homeController HomeController,
	userController UserController,
) *mux.Mux {
	router := mux.NewMux()

	router.AddRoute("/", homeController.handlers["get"])

	router.AddRoute("/users", userController.handlers["list"])

	router.AddRoute("/users/{userID}", userController.handlers["get"])

	router.AddRoute("/users/{userID}/certificates", certificateController.handlers["list"])

	router.AddRoute("/users/{userID}/certificates/add", certificateController.handlers["add"])

	return router
}
