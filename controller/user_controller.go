package controller

import (
	"encoding/hex"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	"github.com/argoproj/pkg/humanize"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/model"
	. "github.com/binaryphile/lilleygram/must"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

type (
	UserController struct {
		handler   *Mux
		repo      sqlrepo.UserRepo
		templates map[string]*Template
	}
)

func NewUserController(repo sqlrepo.UserRepo) UserController {
	c := UserController{
		repo:      repo,
		templates: make(map[string]*Template),
	}

	baseTemplates := []string{
		"view/layout/base.tmpl",
		"view/partial/nav.tmpl",
		"view/partial/footer.tmpl",
	}

	fileNames := map[string]string{
		"certificateList": "view/certificate.list.tmpl",
		"passwordGet":     "view/password.get.tmpl",
		"profileGet":      "view/profile.get.tmpl",
	}

	funcs := template.FuncMap{
		"incr": func(index int) int {
			return index + 1
		},
	}

	for method, fileName := range fileNames {
		templates := append([]string{fileName}, baseTemplates...)

		c.templates[method] = Must(template.New(filepath.Base(fileName)).Funcs(funcs).ParseFiles(templates...))
	}

	return c
}

func (c UserController) CertificateAdd(writer ResponseWriter, request *Request) {
	if request.URL.RawQuery == "" {
		err := writer.SetHeader(gemini.CodeInput, "Certificate's SHA256:")
		if err != nil {
			helper.InternalServerError(writer, err)
			return
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
		gemini.BadRequest(writer, request)
		log.Print("bad sha256")
		return
	}

	err := c.repo.CertificateAdd(sha256, 0, user.UserID)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) CertificateList(writer ResponseWriter, request *Request) {
	user, _ := UserFromContext(request.Context)

	certificates, err := c.repo.CertificateListByUser(user.UserID)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}

	data := struct {
		Certificates []model.Certificate
		User         helper.User
	}{
		Certificates: certificates,
		User:         user,
	}

	err = c.templates["certificateList"].Execute(writer, data)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) Get(writer ResponseWriter, request *Request) {
	user, _ := UserFromContext(request.Context)

	u, found, err := c.repo.Get(user.UserID)
	if err != nil || !found {
		helper.InternalServerError(writer, err)
		return
	}

	_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) Handler(routes ...map[string]FnHandler) *Mux {
	defaultRoutes := map[string]FnHandler{
		"/users":                           c.List,
		"/users/{userID}":                  c.Get,
		"/users/{userID}/certificates":     c.CertificateList,
		"/users/{userID}/certificates/add": c.CertificateAdd,
		"/users/{userID}/password":         c.PasswordGet,
		"/users/{userID}/password/set":     c.PasswordSet,
		"/users/{userID}/profile":          c.ProfileGet,
	}

	var handlers map[string]FnHandler

	if len(routes) > 0 {
		handlers = routes[0]
	} else {
		handlers = defaultRoutes
	}

	router := mux.NewMux()

	for pattern, handler := range handlers {
		router.AddRoute(pattern, HandlerFunc(handler))
	}

	return router
}

func (c UserController) List(writer ResponseWriter, request *Request) {
	user, _ := UserFromContext(request.Context)

	u, found, err := c.repo.Get(user.UserID)
	if err != nil || !found {
		helper.InternalServerError(writer, err)
		return
	}

	_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) PasswordGet(writer ResponseWriter, request *Request) {
	user, _ := UserFromContext(request.Context)

	password, _, err := c.repo.PasswordGet(user.UserID)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}

	data := struct {
		Password model.Password
		User     helper.User
	}{
		Password: password,
		User:     user,
	}

	err = c.templates["passwordGet"].Execute(writer, data)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) PasswordSet(writer ResponseWriter, request *Request) {
	if request.URL.RawQuery == "" {
		err := helper.InputSensitive(writer, "New Password:")
		if err != nil {
			log.Print(err)
		}
		return
	}

	user, _ := UserFromContext(request.Context)

	password := request.URL.RawQuery

	p := model.NewPassword(password)

	err := c.repo.PasswordSet(user.UserID, p)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}

	err = writer.SetHeader(gemini.CodeRedirect, ".")
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) ProfileGet(writer ResponseWriter, request *Request) {
	route, ok := mux.GetMatchedRoute(request.Context)
	if !ok {
		gemini.BadRequest(writer, request)
		return
	}

	s, ok := route.PathVars["userID"]
	if !ok {
		gemini.BadRequest(writer, request)
		return
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		gemini.BadRequest(writer, request)
		return
	}

	userID := uint64(i)

	u, _ := UserFromContext(request.Context)

	p, cs, found, err := c.repo.ProfileGet(userID)
	if err != nil || !found {
		helper.InternalServerError(writer, fmt.Errorf("profile not found: %w", err))
		return
	}

	certificates := make([]helper.Certificate, len(cs))

	for i, certificate := range cs {
		certificates[i] = helper.Certificate{
			CreatedAt: humanTime(certificate.CreatedAt),
			ExpireAt:  humanTime(certificate.ExpireAt),
		}
	}

	profile := helper.Profile{
		Avatar:        p.Avatar,
		Certificates:  certificates,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		LastSeen:      humanTime(p.LastSeen),
		Me:            userID == u.UserID,
		PasswordFound: p.Password.Valid,
		UserID:        fmt.Sprintf("%d", userID),
		UserName:      p.UserName,
		CreatedAt:     humanTime(p.CreatedAt),
	}

	err = c.templates["profileGet"].Execute(writer, profile)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}

func humanTime(unixTime int64) string {
	if unixTime == 0 {
		return ""
	}

	unix := time.Unix(unixTime, 0)

	if time.Since(unix) > 48*time.Hour {
		return unix.Format("02 Jan 2006 15:04")
	}

	return humanize.RelativeDuration(unix, time.Now().UTC())
}
