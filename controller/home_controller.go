package controller

import (
	"errors"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware"
	"log"
	"text/template"
)

type (
	HomeController struct {
		handler   *Mux
		templates map[string]*Template
	}
)

func NewHomeController() HomeController {
	c := HomeController{
		templates: make(map[string]*Template),
	}

	baseTemplates := []string{
		"view/partial/nav.tmpl",
		"view/layout/base.tmpl",
		"view/partial/footer.tmpl",
	}

	{
		templates := append([]string{"view/home.get.tmpl"}, baseTemplates...)

		tmpl, err := template.ParseFiles(templates...)
		if err != nil {
			log.Panic(err)
		}

		c.templates["get"] = tmpl
	}

	return c
}

func (c HomeController) Get(writer ResponseWriter, request *Request) {
	user, ok := UserFromContext(request.Context)
	if !ok {
		helper.InternalServerError(writer, errors.New("user should exist"))
		return
	}

	err := c.templates["get"].Execute(writer, user)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c HomeController) Handler(routes ...map[string]FnHandler) *Mux {
	defaultRoutes := map[string]FnHandler{
		"/": c.Get,
	}

	var handlers map[string]FnHandler

	if len(routes) > 0 {
		handlers = routes[0]
	} else {
		handlers = defaultRoutes
	}

	router := mux.NewMux()

	for pattern, handler := range handlers {
		router.AddRoute(pattern, HandlerFunc(handler))
	}

	return router
}

func (c HomeController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}
