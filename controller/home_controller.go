package controller

import (
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"log"
	"text/template"
)

type (
	HomeController struct {
		Get Handler
	}
)

func NewHomeController(middlewares ...map[string][]Middleware) HomeController {
	authTmpls, err := template.ParseFiles(
		"view/home.page.tmpl",
		"view/base.layout.tmpl",
		"view/footer.partial.tmpl",
	)
	if err != nil {
		log.Panic(err)
	}

	unauthTmpls, err := template.ParseFiles(
		"view/home.page.user.tmpl",
		"view/base.layout.tmpl",
		"view/footer.partial.tmpl",
	)
	if err != nil {
		log.Panic(err)
	}

	get := func(writer ResponseWriter, request *Request) {
		if user, ok := UserFromContext(request.Context); ok {
			err = authTmpls.Execute(writer, user)
			if err != nil {
				log.Panicf("couldn't execute template: %s", err)
			}
			return
		}

		err = unauthTmpls.Execute(writer, nil)
		if err != nil {
			log.Panic(err)
		}
	}

	methods := []string{
		"get",
	}

	for _, method := range methods {
		switch method {
		case "get":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[method]; ok {
					get = ExtendFnHandler(get, extensions...)
				}
			}
		}
	}

	return HomeController{
		Get: HandlerFunc(get),
	}
}

func (c HomeController) Router() *mux.Mux {
	router := mux.NewMux()

	routes := map[string]Handler{
		"/": c.Get,
	}

	for pattern, handler := range routes {
		router.AddRoute(pattern, handler)
	}

	return router
}
