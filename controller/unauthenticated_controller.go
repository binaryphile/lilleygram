package controller

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/gmni"
	"github.com/binaryphile/lilleygram/handler"
	"github.com/binaryphile/lilleygram/hash"
	"github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/opt"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"math/rand"
)

type (
	UnauthenticatedController struct {
		handler *Mux
		repo    sqlrepo.UserRepo
	}
)

func NewUnauthenticatedController(repo sqlrepo.UserRepo) UnauthenticatedController {
	c := UnauthenticatedController{
		repo: repo,
	}

	return c
}

func (c UnauthenticatedController) CertificateAdd(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = gmni.InputSensitive(writer, "Password:")
		return
	}

	rawPassword := request.URL.RawQuery

	userID, ok := middleware.Uint64FromRequest(request, "userID")
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("no user id")
		return
	}

	password, found, err := c.repo.PasswordGet(userID)
	if err != nil {
		gmni.InternalServerError(writer, err)
		return
	}

	salt, err := base64.RawStdEncoding.DecodeString(password.Salt)
	if err != nil {
		gmni.InternalServerError(writer, err)
	}

	if hash.ComparePasswords(rawPassword, salt, password.Argon2) && found {
		certID := request.Certificate.ID

		if certID == "" {
			_, err := writer.Write([]byte(Heredoc(`
				Error: no certificate supplied.  Please enable your certificate and try again.
				=> . Try Again
			`)))
			if err != nil {
				gmni.InternalServerError(writer, err)
				return
			}

			return
		}

		idHash := sha256.Sum256([]byte(certID))

		certSHA256 := hex.EncodeToString(idHash[:])

		err = c.repo.CertificateAdd(certSHA256, 0, uint64(userID))
		if err != nil {
			gmni.InternalServerError(writer, err)
			return
		}

		_, err = writer.Write([]byte(Heredoc(`
			Certificate added successfully.
			=> / Return home
		`)))
		if err != nil {
			gmni.InternalServerError(writer, err)
			return
		}

		return
	}

	_, err = writer.Write([]byte(Heredoc(`
		Either the username or password were incorrect.
		=> /register/username/check Try again
	`)))
	if err != nil {
		gmni.InternalServerError(writer, err)
		return
	}
}

func (c UnauthenticatedController) CodeCheck(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = gmni.InputSensitive(writer, "Registration code:")
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
				Error: no certificate supplied.  In order to register, you must go to the home page, enable your certificate, and then return to this page to register.
				Note, if you do not enable the certificate on the home page rather than this page, registration will not work.
				=> / Go Home
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
		func(tx sqlrepo.UserRepo) error {
			userID, err = tx.Add("", "", fmt.Sprintf("slug %d", rand.Int()), "")
			if err != nil {
				return err
			}

			return tx.CertificateAdd(certSHA256, 0, userID)
		},
	)
	if err != nil {
		return
	}

	err = gmni.Redirect(writer, fmt.Sprintf("/users/%d/username/set", userID))
}

func (c UnauthenticatedController) Handler(routes ...map[string]Handler) *Mux {
	handlers := opt.OfFirst(routes).Or(c.Routes())

	router := mux.NewMux()

	for pattern, h := range handlers {
		router.AddRoute(pattern, h)
	}

	return router
}

func (c UnauthenticatedController) Routes() map[string]Handler {
	baseTemplates := []string{
		"view/unauthenticated/partial/nav.tmpl",
		"view/layout/base.tmpl",
		"view/partial/footer.tmpl",
	}

	gettingStartedTemplates := append([]string{"view/unauthenticated/getting-started.tmpl"}, baseTemplates...)
	homeTemplates := append([]string{"view/unauthenticated/home.tmpl"}, baseTemplates...)
	registerTemplates := append([]string{"view/unauthenticated/register.tmpl"}, baseTemplates...)

	return map[string]Handler{
		"/":                                  handler.FileHandler(homeTemplates...),
		"/getting-started":                   handler.FileHandler(gettingStartedTemplates...),
		"/register":                          handler.FileHandler(registerTemplates...),
		"/register/code/check":               HandlerFunc(c.CodeCheck),
		"/register/username/check":           HandlerFunc(c.UserNameCheck),
		"/register/{userID}/certificate/add": HandlerFunc(c.CertificateAdd),
	}
}

func (c UnauthenticatedController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}

func (c UnauthenticatedController) UserNameCheck(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = gmni.InputPrompt(writer, "Provide your account's username:")
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
			=> /register/username/check Try again
		`)))
		if err != nil {
			return
		}
	}

	err = gmni.Redirect(writer, fmt.Sprintf("/register/%d/certificate/add", user.ID))
}

func writeError(writer ResponseWriter, err error) {
	if err != nil {
		gmni.InternalServerError(writer, err)
	}
}
