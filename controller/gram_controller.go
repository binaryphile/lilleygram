package controller

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/gmni"
	"github.com/binaryphile/lilleygram/helper"
	"github.com/binaryphile/lilleygram/middleware"
	. "github.com/binaryphile/lilleygram/must"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/slice"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"net/url"
	"path/filepath"
	"text/template"
)

type GramController struct {
	baseTemplateNames []string
	funcs             template.FuncMap
	listTemplate      *Template
	handler           *Mux
	repo              sqlrepo.GramRepo
}

func NewGramController(repo sqlrepo.GramRepo) GramController {
	c := GramController{
		baseTemplateNames: []string{
			"view/layout/base.tmpl",
			"view/partial/nav.tmpl",
			"view/partial/footer.tmpl",
		},
		funcs: template.FuncMap{
			"incr": func(index int) int {
				return index + 1
			},
		},
		repo: repo,
	}

	fileName := "view/timeline.tmpl"

	templates := append([]string{fileName}, c.baseTemplateNames...)

	c.listTemplate = Must(template.New(filepath.Base(fileName)).Funcs(c.funcs).ParseFiles(templates...))

	return c
}

func (c GramController) Add(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = gmni.InputPrompt(writer, "Compose your gram of up to 500 characters:")
		return
	}

	user, _ := middleware.CertUserFromRequest(request)

	gram, err := url.QueryUnescape(request.URL.RawQuery)

	_, err = c.repo.Add(user.UserID, gram)
	if err != nil {
		return
	}

	err = gmni.Redirect(writer, "/")
}

func (c GramController) List(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	grams, err := c.repo.List(user.UserID)
	if err != nil {
		return
	}

	data := struct {
		helper.User
		Grams []helper.Gram
	}{
		User:  user,
		Grams: slice.Map(helper.GramFromModel, grams),
	}

	if deployEnv, ok := middleware.DeployEnvFromRequest(request); ok && deployEnv == "local" {
		fileName := "view/timeline.tmpl"

		templates := append([]string{fileName}, c.baseTemplateNames...)

		c.listTemplate = Must(template.New(filepath.Base(fileName)).Funcs(c.funcs).ParseFiles(templates...))
	}

	err = c.listTemplate.Execute(writer, data)
}

func (c GramController) Handler(routes ...map[string]Handler) *Mux {
	handlers := opt.OfFirst(routes).Or(c.Routes())

	router := mux.NewMux()

	for pattern, h := range handlers {
		router.AddRoute(pattern, h)
	}

	return router
}

func (c GramController) Routes() map[string]Handler {
	return map[string]Handler{
		"/":                   HandlerFunc(c.List),
		"/grams/add":          HandlerFunc(c.Add),
		"/grams/{id}/sparkle": HandlerFunc(c.Sparkle),
	}
}

func (c GramController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}

func (c GramController) Sparkle(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	gramID, ok := middleware.Uint64FromRequest(request, "id")
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("no ID")
		return
	}

	_, err = c.repo.Sparkle(gramID, user.UserID)

	err = gmni.Redirect(writer, "/")
}
