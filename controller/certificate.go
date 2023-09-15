package controller

import (
	"encoding/hex"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/sql"
	"log"
	"text/template"
)

type (
	CertificateController struct {
		handlers map[string]Handler
		repo     sql.CertificateRepo
	}
)

func NewCertificateController(
	repo sql.CertificateRepo, specific map[string][]Middleware, general ...Middleware,
) CertificateController {
	c := CertificateController{
		handlers: make(map[string]Handler),
		repo:     repo,
	}

	methods := map[string]Handler{
		"add":  HandlerFunc(c.Add),
		"list": HandlerFunc(c.List),
	}

	for key, method := range methods {
		s := opt.OfIndex(specific, key)

		c.handlers[key] = ExtendHandler(method, s.OrZero()...)

		c.handlers[key] = ExtendHandler(c.handlers[key], general...)
	}

	return c
}

func (c CertificateController) Add(writer ResponseWriter, request *Request) {
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

	_, err := c.repo.Add(sha256, 0, user.ID)
	if err != nil {
		log.Println(err)
	}
}

func (c CertificateController) Handlers() map[string]Handler {
	return c.handlers
}

func (c CertificateController) List(writer ResponseWriter, request *Request) {
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

	certificates, err := c.repo.ListByUser(user.ID)
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
