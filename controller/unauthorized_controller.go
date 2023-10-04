package controller

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/handler"
	"github.com/binaryphile/lilleygram/hash"
	"github.com/binaryphile/lilleygram/helper"
	"github.com/binaryphile/lilleygram/middleware"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"strconv"
)

type (
	UnauthorizedController struct {
		handler *Mux
		repo    sqlrepo.UserRepo
	}
)

func NewUnauthorizedController(repo sqlrepo.UserRepo) UnauthorizedController {
	c := UnauthorizedController{
		repo: repo,
	}

	return c
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

	strID, ok := middleware.PathVarFromRequest("userID", request)
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("no userID")
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

func (c UnauthorizedController) CodeCheck(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = helper.InputSensitive(writer, "Registration code:")
		return
	}

	userCode := request.URL.RawQuery

	code, found, err := c.repo.CodeGet(0)
	if err != nil || !found || userCode != code {
		_, writeErr := writer.Write([]byte(Heredoc(`
			Registrations are not open at the moment.
			=> / Go Home
		`)))
		if writeErr != nil {
			err = writeErr
		}

		return
	}

	certID := request.Certificate.ID

	if certID == "" {
		_, writeErr := writer.Write([]byte(Heredoc(`
				Error: no certificate supplied.  Please enable your certificate and try again.
				=> . Try Again
			`)))
		if writeErr != nil {
			err = writeErr
		}

		return
	}

	idHash := sha256.Sum256([]byte(certID))

	certSHA256 := hex.EncodeToString(idHash[:])

	var userID uint64

	err = c.repo.WithTx(
		func(tx sqlrepo.UserRepo) (err error) {
			userID, err = tx.Add("", "", "", "")
			if err != nil {
				return
			}

			return tx.CertificateAdd(certSHA256, 0, userID)
		},
	)
	if err != nil {
		return
	}

	err = writer.SetHeader(gemini.CodeRedirect, fmt.Sprintf("/users/%d/username/set", userID))
	if err != nil {
		return
	}
}

func (c UnauthorizedController) Handler(routes ...map[string]Handler) *Mux {
	var handlers map[string]Handler

	if len(routes) > 0 {
		handlers = routes[0]
	} else {
		handlers = c.Routes()
	}

	router := mux.NewMux()

	for pattern, h := range handlers {
		router.AddRoute(pattern, h)
	}

	return router
}

func (c UnauthorizedController) Routes() map[string]Handler {
	baseTemplates := []string{
		"view/unauthorized/partial/nav.tmpl",
		"view/layout/base.tmpl",
		"view/partial/footer.tmpl",
	}

	getTemplates := append([]string{"view/unauthorized/home.get.tmpl"}, baseTemplates...)
	gettingStartedTemplates := append([]string{"view/unauthorized/getting-started.tmpl"}, baseTemplates...)
	registerTemplates := append([]string{"view/unauthorized/register.tmpl"}, baseTemplates...)

	return map[string]Handler{
		"/":                                  handler.FileHandler(getTemplates...),
		"/getting-started":                   handler.FileHandler(gettingStartedTemplates...),
		"/register":                          handler.FileHandler(registerTemplates...),
		"/register/code/check":               HandlerFunc(c.CodeCheck),
		"/register/username/check":           HandlerFunc(c.UserNameCheck),
		"/register/{userID}/certificate/add": HandlerFunc(c.CertificateAdd),
	}
}

func (c UnauthorizedController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}

func (c UnauthorizedController) UserNameCheck(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = helper.InputPrompt(writer, "Provide your account's username:")
		return
	}

	userName := request.URL.RawQuery

	user, found, err := c.repo.GetByUserName(userName)
	if err != nil {
		return
	}

	if !found {
		_, err = writer.Write([]byte(Heredoc(`
			That user was not found.  If you feel this result was in error, click the link below to try again.
			=> /register/username/check Resubmit helperUser
		`)))
		if err != nil {
			return
		}
	}

	err = writer.SetHeader(gemini.CodeRedirect, fmt.Sprintf("/register/%d/certificate/add", user.ID))
}

func writeError(writer ResponseWriter, err error) {
	if err != nil {
		helper.InternalServerError(writer, err)
	}
}
