package handler

import (
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/gmni"
	. "github.com/binaryphile/lilleygram/middleware"
	"log"
	"text/template"
)

func FileHandler(fileNames ...string) Handler {
	tmpl, err := template.ParseFiles(fileNames...)
	if err != nil {
		log.Panic(err)
	}

	return HandlerFunc(func(w ResponseWriter, request *Request) {
		user, _ := CertUserFromRequest(request)

		if deployEnv, ok := DeployEnvFromRequest(request); ok && deployEnv == "local" {
			var err error

			tmpl, err = template.ParseFiles(fileNames...)
			if err != nil {
				log.Panic(err)
			}
		}

		err := tmpl.Execute(w, user)
		if err != nil {
			gmni.InternalServerError(w, err)
			return
		}
	})
}
