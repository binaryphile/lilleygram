package controller

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/gmnifc"
	"github.com/binaryphile/lilleygram/helper"
	"github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/slice"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"github.com/binaryphile/lilleygram/sqlrepo/defaults"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const (
	tagPattern = `(^|\s)(#[[:alpha:]]\w+[[:alnum:]]+\b)`
)

var (
	tagRegex = regexp.MustCompile(tagPattern)
)

type GramController struct {
	baseTemplateNames []string
	discoverTemplate  *Template
	funcs             template.FuncMap
	handler           *Mux
	listTemplate      *Template
	repo              sqlrepo.GramRepo
	tagGetTemplate    *Template
}

func NewGramController(repo sqlrepo.GramRepo) *GramController {
	c := &GramController{
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

	c.DiscoverRefresh()

	c.ListRefresh()

	c.TagGetRefresh()

	return c
}

func (c *GramController) Add(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = gmnifc.InputPrompt(writer, "Compose your gram of up to 500 characters:")
		return
	}

	user, _ := middleware.CertUserFromRequest(request)

	gram, err := url.QueryUnescape(request.URL.RawQuery)
	if err != nil {
		return
	}

	if strings.HasPrefix(gram, "#") {
		gram = " " + gram
	}

	matches := tagRegex.FindAllStringSubmatch(gram, -1)

	var tags []string

	for _, match := range matches {
		tag := match[2]

		if len(tag) <= 26 && !strings.Contains(tag, "__") {
			tags = append(tags, tag)
		}
	}

	_, err = c.repo.Add(user.UserID, gram, tags...)
	if err != nil {
		return
	}

	err = gmnifc.Redirect(writer, "/")
}

func (c *GramController) Discover(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	recentGrams, _, err := c.repo.ListPublic(user.UserID, "", defaults.SectionSize)
	if err != nil {
		return
	}

	recentIntros, _, err := c.repo.ListByTag(user.UserID, "#introduction", "", false, defaults.SectionSize)
	if err != nil {
		return
	}

	toHelperGram := helper.GramFromModel(user.UserID)

	data := struct {
		helper.User
		RecentIntros []helper.Gram
		RecentGrams  []helper.Gram
	}{
		RecentIntros: slice.Map(toHelperGram, recentIntros),
		RecentGrams:  slice.Map(toHelperGram, recentGrams),
		User:         user,
	}

	err = c.discoverTemplate.Execute(writer, data)
}

func (c *GramController) DiscoverRefresh() {
	fileName := "view/discover.tmpl"

	templates := append([]string{fileName}, c.baseTemplateNames...)

	c.discoverTemplate = Must(template.New(filepath.Base(fileName)).Funcs(c.funcs).ParseFiles(templates...))
}

func (c *GramController) Handler(routes ...map[string]Handler) *Mux {
	handlers := opt.OfFirst(routes).Or(c.Routes())

	router := mux.NewMux()

	for pattern, h := range handlers {
		router.AddRoute(pattern, h)
	}

	return router
}

func (c *GramController) List(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	pageToken := middleware.PageTokenFromRequest(request)

	grams, npt, err := c.repo.List(user.UserID, pageToken)
	if err != nil {
		return
	}

	data := struct {
		helper.User
		Grams         []helper.Gram
		NextPageToken string
	}{
		User:          user,
		Grams:         slice.Map(helper.GramFromModel(user.UserID), grams),
		NextPageToken: npt,
	}

	err = c.listTemplate.Execute(writer, data)
}

func (c *GramController) ListRefresh() {
	fileName := "view/timeline.tmpl"

	templates := append([]string{fileName}, c.baseTemplateNames...)

	c.listTemplate = Must(template.New(filepath.Base(fileName)).Funcs(c.funcs).ParseFiles(templates...))
}

func (c *GramController) Routes() map[string]Handler {
	return map[string]Handler{
		"/":                   ExtendHandler(HandlerFunc(c.List), WithRefresh(c.ListRefresh)),
		"/discover":           ExtendHandler(HandlerFunc(c.Discover), WithRefresh(c.DiscoverRefresh)),
		"/grams/add":          HandlerFunc(c.Add),
		"/grams/{id}/sparkle": HandlerFunc(c.Sparkle),
		"/tags/{tag}":         ExtendHandler(HandlerFunc(c.TagGet), WithRefresh(c.TagGetRefresh)),
	}
}

func (c *GramController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}

func (c *GramController) Sparkle(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	id, ok := middleware.Uint64FromRequest(request, "id")
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("no ID")
		return
	}

	_, err = c.repo.Sparkle(id, user.UserID)

	err = gmnifc.Redirect(writer, "/")
}

func (c *GramController) TagGet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	tag, ok := middleware.StrFromRequest(request, "tag")
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("missing tag")
		return
	}

	tag = "#" + tag

	pageToken := middleware.PageTokenFromRequest(request)

	grams, npt, err := c.repo.ListByTag(user.UserID, tag, pageToken, true)

	data := struct {
		helper.User
		Grams         []helper.Gram
		Hashtag       string
		NextPageToken string
	}{
		User:          user,
		Grams:         slice.Map(helper.GramFromModel(user.UserID), grams),
		Hashtag:       tag,
		NextPageToken: npt,
	}

	err = c.tagGetTemplate.Execute(writer, data)
}

func (c *GramController) TagGetRefresh() {
	fileName := "view/tag_get.tmpl"

	templates := append([]string{fileName}, c.baseTemplateNames...)

	c.tagGetTemplate = Must(template.New(filepath.Base(fileName)).Funcs(c.funcs).ParseFiles(templates...))
}
