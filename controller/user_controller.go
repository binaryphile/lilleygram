package controller

import (
	"encoding/hex"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"text/template"
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

	{
		tmpl, err := template.ParseFiles(append([]string{"view/certificate.list.tmpl"}, baseTemplates...)...)
		if err != nil {
			log.Panic(err)
		}

		c.templates["certificateList"] = tmpl
	}

	{
		tmpl, err := template.ParseFiles(append([]string{"view/password.get.tmpl"}, baseTemplates...)...)
		if err != nil {
			log.Panic(err)
		}

		c.templates["passwordGet"] = tmpl
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

	err := c.repo.CertificateAdd(sha256, 0, user.ID)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) CertificateList(writer ResponseWriter, request *Request) {
	user, _ := UserFromContext(request.Context)

	certificates, err := c.repo.CertificateListByUser(user.ID)
	if err != nil {
		helper.InternalServerError(writer, err)
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

	err = c.templates["certificateList"].Execute(writer, data)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UserController) Get(writer ResponseWriter, request *Request) {
	user, ok := UserFromContext(request.Context)
	if !ok {
		return
	}

	u, found, err := c.repo.Get(user.ID)
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

func (c UserController) List(writer ResponseWriter, request *Request) {
	user, ok := UserFromContext(request.Context)
	if !ok {
		return
	}

	u, found, err := c.repo.Get(user.ID)
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

	password, _, err := c.repo.PasswordGet(user.ID)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}

	type User struct {
		Avatar   string
		ID       uint64
		UserName string
	}

	data := struct {
		Password model.Password
		User     User
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

	err := c.repo.PasswordSet(user.ID, p)
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

func (c UserController) Handler(routes ...map[string]FnHandler) *Mux {
	defaultRoutes := map[string]FnHandler{
		"/users":                           c.List,
		"/users/{userID}":                  c.Get,
		"/users/{userID}/certificates":     c.CertificateList,
		"/users/{userID}/certificates/add": c.CertificateAdd,
		"/users/{userID}/password":         c.PasswordGet,
		"/users/{userID}/password/set":     c.PasswordSet,
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

func (c UserController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}
