package handler

import (
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/gmni"
	"github.com/binaryphile/lilleygram/middleware"
	"log"
	"text/template"
)

func FileHandler(fileNames ...string) Handler {
	var tmpl *Template

	var err error

	refresh := func() {
		tmpl, err = template.ParseFiles(fileNames...)
		if err != nil {
			log.Panic(err)
		}
	}

	refresh()

	return HandlerFunc(func(w ResponseWriter, r *Request) {
		user, _ := middleware.CertUserFromRequest(r)

		LocalEnvFromRequest(r).AndDo(refresh)

		err := tmpl.Execute(w, user)
		if err != nil {
			gmni.InternalServerError(w, err)
			return
		}
	})
}
