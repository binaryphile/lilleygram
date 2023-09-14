package controller

import (
	"github.com/a-h/gemini/mux"
)

func Router(userController, certificateController, homeController Controller) *mux.Mux {
	router := mux.NewMux()

	router.AddRoute("/", homeController["get"])

	router.AddRoute("/users", userController["list"])

	router.AddRoute("/users/{userID}", userController["get"])

	router.AddRoute("/users/{userID}/certificates", certificateController["list"])

	router.AddRoute("/users/{userID}/certificates/add", certificateController["add"])

	return router
}
