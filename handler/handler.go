package handler

import (
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware"
	"log"
	"text/template"
)

func FileHandler(fileNames ...string) Handler {
	tmpl, err := template.ParseFiles(fileNames...)
	if err != nil {
		log.Panic(err)
	}

	return HandlerFunc(func(writer ResponseWriter, request *Request) {
		user, _ := UserFromRequest(request)

		err := tmpl.Execute(writer, user)
		if err != nil {
			helper.InternalServerError(writer, err)
			return
		}
	})
}
