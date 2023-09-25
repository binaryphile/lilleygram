package handler

import (
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware"
	"log"
	"text/template"
)

func UserNameCheck() Handler {
	tmpl, err := template.ParseFiles(
		"view/register.tmpl",
		"view/layout/base.tmpl",
		"view/partial/nav.tmpl",
	)
	if err != nil {
		log.Panic(err)
	}

	return HandlerFunc(func(writer ResponseWriter, request *Request) {
		user, _ := UserFromContext(request.Context)

		err := tmpl.Execute(writer, user)
		if err != nil {
			helper.InternalServerError(writer, err)
			return
		}
	})
}
