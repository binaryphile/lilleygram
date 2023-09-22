package controller

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/hash"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"strconv"
	"text/template"
)

type (
	UnauthorizedController struct {
		handler   *Mux
		repo      sqlrepo.UserRepo
		templates map[string]*Template
	}
)

func NewUnauthorizedController(repo sqlrepo.UserRepo) UnauthorizedController {
	c := UnauthorizedController{
		repo:      repo,
		templates: make(map[string]*Template),
	}

	baseTemplates := []string{
		"view/unauthorized/partial/nav.tmpl",
		"view/layout/base.tmpl",
		"view/partial/footer.tmpl",
	}

	{
		templates := append([]string{"view/unauthorized/home.get.tmpl"}, baseTemplates...)

		tmpl, err := template.ParseFiles(templates...)
		if err != nil {
			log.Panic(err)
		}

		c.templates["get"] = tmpl
	}

	{
		templates := append([]string{"view/unauthorized/register.tmpl"}, baseTemplates...)

		tmpl, err := template.ParseFiles(templates...)
		if err != nil {
			log.Panic(err)
		}

		c.templates["register"] = tmpl
	}

	return c
}

func (c UnauthorizedController) UserNameCheck(writer ResponseWriter, request *Request) {
	if request.URL.RawQuery == "" {
		err := helper.InputPrompt(writer, "Provide your account's username:")
		if err != nil {
			helper.InternalServerError(writer, err)
			return
		}

		return
	}

	userName := request.URL.RawQuery

	user, found, err := c.repo.GetByUserName(userName)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}

	if !found {
		_, err = writer.Write([]byte(Heredoc(`
			That user was not found.  If you feel this result was in error, click the link below to try again.
			=> /register/username/check Resubmit User
		`)))
		if err != nil {
			helper.InternalServerError(writer, err)
			return
		}
	}

	err = writer.SetHeader(gemini.CodeRedirect, fmt.Sprintf("/register/%d/certificate/add", user.ID))
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UnauthorizedController) Get(writer ResponseWriter, _ *Request) {
	err := c.templates["get"].Execute(writer, nil)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UnauthorizedController) Handler(routes ...map[string]FnHandler) *Mux {
	var handlers map[string]FnHandler

	if len(routes) > 0 {
		handlers = routes[0]
	} else {
		handlers = c.Routes()
	}

	router := mux.NewMux()

	for pattern, handler := range handlers {
		router.AddRoute(pattern, HandlerFunc(handler))
	}

	return router
}

func (c UnauthorizedController) CertificateAdd(writer ResponseWriter, request *Request) {
	if request.URL.RawQuery == "" {
		err := helper.InputSensitive(writer, "Password:")
		if err != nil {
			return
		}

		return
	}

	rawPassword := request.URL.RawQuery

	route, ok := mux.GetMatchedRoute(request.Context)
	if !ok {
		helper.InternalServerError(writer, errors.New("no route match"))
		return
	}

	strID, ok := route.PathVars["userid"]
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("no userid")
		return
	}

	userID, err := strconv.Atoi(strID)
	if err != nil {
		gemini.BadRequest(writer, request)
		log.Print(err)
		return
	}

	password, found, err := c.repo.PasswordGet(uint64(userID))
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}

	salt, err := base64.RawStdEncoding.DecodeString(password.Salt)
	if err != nil {
		helper.InternalServerError(writer, err)
	}

	if hash.ComparePasswords(rawPassword, salt, password.Argon2) && found { // order matters for security
		certID := request.Certificate.ID

		if certID == "" {
			_, err := writer.Write([]byte(Heredoc(`
				Error: no certificate supplied.  Please enable your certificate and try again.
				=> . Try Again
			`)))
			if err != nil {
				helper.InternalServerError(writer, err)
				return
			}

			return
		}

		idHash := sha256.Sum256([]byte(certID))

		certSHA256 := hex.EncodeToString(idHash[:])

		err = c.repo.CertificateAdd(certSHA256, 0, uint64(userID))
		if err != nil {
			helper.InternalServerError(writer, err)
			return
		}

		_, err = writer.Write([]byte(Heredoc(`
			Certificate added successfully.
			=> / Return home
		`)))
		if err != nil {
			helper.InternalServerError(writer, err)
			return
		}

		return
	}

	_, err = writer.Write([]byte(Heredoc(`
		Either the username or password were incorrect.
		=> /register/username/check Try Again
	`)))
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UnauthorizedController) Register(writer ResponseWriter, request *Request) {
	err := c.templates["register"].Execute(writer, request)
	if err != nil {
		helper.InternalServerError(writer, err)
		return
	}
}

func (c UnauthorizedController) Routes() map[string]FnHandler {
	return map[string]FnHandler{
		"/":                                  c.Get,
		"/register":                          c.Register,
		"/register/username/check":           c.UserNameCheck,
		"/register/{userid}/certificate/add": c.CertificateAdd,
	}
}

func (c UnauthorizedController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}
