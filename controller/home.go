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
		get FnHandler
	}
)

func NewHomeController(middlewares ...Middleware) HomeController {
	get := func(writer ResponseWriter, request *Request) {
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

	return HomeController{
		get: ExtendFnHandler(get, middlewares...),
	}
}

func (c HomeController) Get(writer ResponseWriter, request *Request) {
	c.get(writer, request)
}
