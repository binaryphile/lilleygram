package controller

import (
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	"github.com/binaryphile/lilleygram/middleware"
	. "github.com/binaryphile/lilleygram/must"
	"github.com/binaryphile/lilleygram/opt"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/slice"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"net/url"
	"path/filepath"
	"text/template"
)

type GramController struct {
	getTemplate *Template
	handler     *Mux
	repo        sqlrepo.GramRepo
}

func NewGramController(repo sqlrepo.GramRepo) GramController {
	c := GramController{
		repo: repo,
	}

	baseTemplates := []string{
		"view/layout/base.tmpl",
		"view/partial/nav.tmpl",
		"view/partial/footer.tmpl",
	}

	funcs := template.FuncMap{
		"incr": func(index int) int {
			return index + 1
		},
	}

	fileName := "view/home.get.tmpl"
	templates := append([]string{fileName}, baseTemplates...)
	c.getTemplate = Must(template.New(filepath.Base(fileName)).Funcs(funcs).ParseFiles(templates...))

	return c
}

func (c GramController) Add(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = helper.InputPrompt(writer, "Compose your gram of up to 500 characters:")
		return
	}

	user, _ := middleware.UserFromRequest(request)

	gram, err := url.QueryUnescape(request.URL.RawQuery)

	_, err = c.repo.Add(user.UserID, gram)
	if err != nil {
		return
	}

	_, err = writer.Write([]byte(Heredoc(`
		Your gram has been posted!
		=> / Go Home
	`)))
}

func (c GramController) Get(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.UserFromRequest(request)

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

	err = c.getTemplate.Execute(writer, data)
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
		"/":          HandlerFunc(c.Get),
		"/grams/add": HandlerFunc(c.Add),
	}
}

func (c GramController) ServeGemini(writer ResponseWriter, request *Request) {
	c.handler = opt.OfPointer(c.handler).Or(c.Handler())

	c.handler.ServeGemini(writer, request)
}
