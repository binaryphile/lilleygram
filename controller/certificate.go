package controller

import (
	"encoding/hex"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"text/template"
)

type (
	CertificateController struct {
		add  FnHandler
		list FnHandler
	}
)

func NewCertificateController(repo sqlrepo.CertificateRepo, middlewares ...Middleware) CertificateController {
	addWith := func(repo sqlrepo.CertificateRepo) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			if request.URL.RawQuery == "" {
				err := writer.SetHeader(gemini.CodeInput, "Certificate's SHA256:")
				if err != nil {
					log.Println(err)
				}
				return
			}

			user, _ := UserFromContext(request.Context)

			sha256 := request.URL.RawQuery

			bad := len(sha256) != 64

			if !bad {
				_, err := hex.DecodeString(sha256)
				bad = err != nil
			}

			if bad {
				panic(bad)
			}

			_, err := repo.Add(sha256, 0, user.ID)
			if err != nil {
				log.Println(err)
			}
		}
	}

	listWith := func(repo sqlrepo.CertificateRepo) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			user, _ := UserFromContext(request.Context)

			ts, err := template.ParseFiles(
				"gemtext/certificate.index.tmpl",
				"gemtext/base.layout.tmpl",
				"gemtext/footer.partial.tmpl",
			)
			if err != nil {
				log.Println(err)

				err := writer.SetHeader(gemini.CodeCGIError, "internal error")
				if err != nil {
					log.Println(err)
					panic(err)
				}

				return
			}

			certificates, err := repo.ListByUser(user.ID)
			if err != nil {
				log.Println(err)
				return
			}

			type User struct {
				Avatar   string
				ID       uint64
				UserName string
			}

			data := struct {
				Certificates []model.Certificate
				User         User
			}{
				Certificates: certificates,
				User:         user,
			}

			err = ts.Execute(writer, data)
			if err != nil {
				log.Panicf("couldn't execute template: %s", err)
			}
		}
	}

	return CertificateController{
		add:  ExtendFnHandler(addWith(repo), middlewares...),
		list: ExtendFnHandler(listWith(repo), middlewares...),
	}
}

func (c CertificateController) Add(writer ResponseWriter, request *Request) {
	c.add(writer, request)
}

func (c CertificateController) List(writer ResponseWriter, request *Request) {
	c.list(writer, request)
}
