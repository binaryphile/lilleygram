package controller

import (
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/opt"
	"log"
	"text/template"
)

type (
	HomeController struct {
		handlers map[string]Handler
	}
)

func NewHomeController(specific map[string][]Middleware, general ...Middleware) HomeController {
	c := HomeController{
		handlers: make(map[string]Handler),
	}

	handlers := map[string]Handler{
		"get": HandlerFunc(c.Get),
	}

	for key, handler := range handlers {
		s := opt.OfIndex(specific, key)

		c.handlers[key] = ExtendHandler(handler, s.OrZero()...)

		c.handlers[key] = ExtendHandler(c.handlers[key], general...)
	}

	return c
}

func (c HomeController) Get(writer ResponseWriter, request *Request) {
	user := opt.Of(UserFromContext(request.Context))

	ts, err := template.ParseFiles(
		opt.OkOrNot(user, "gemtext/home.page.user.tmpl", "gemtext/home.page.tmpl"),
		"gemtext/base.layout.tmpl",
		"gemtext/footer.partial.tmpl",
	)
	if err != nil {
		log.Println(err.Error())

		err := writer.SetHeader(gemini.CodeCGIError, "internal error")
		if err != nil {
			log.Println(err.Error())
			panic(err)
		}
	}

	err = ts.Execute(writer, user.OrZero())
	if err != nil {
		log.Panicf("couldn't execute template: %s", err)
	}
}
