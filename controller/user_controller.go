package controller

import (
	"bytes"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	"github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/model"
	. "github.com/binaryphile/lilleygram/must"
	"github.com/binaryphile/lilleygram/opt"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"text/template"
)

type (
	UserController struct {
		baseTemplateNames []string
		funcs             template.FuncMap
		handler           *Mux
		repo              sqlrepo.UserRepo
		templates         map[string]*Template
	}
)

func NewUserController(repo sqlrepo.UserRepo) UserController {
	c := UserController{
		baseTemplateNames: []string{
			"view/layout/base.tmpl",
			"view/partial/footer.tmpl",
			"view/partial/nav.tmpl",
		},
		funcs: template.FuncMap{
			"incr": func(index int) int {
				return index + 1
			},
		},
		repo:      repo,
		templates: make(map[string]*Template),
	}

	fileNames := map[string]string{
		"passwordGet": "view/password.get.tmpl",
		"profileGet":  "view/profile.get.tmpl",
	}

	for method, fileName := range fileNames {
		templates := append([]string{fileName}, c.baseTemplateNames...)

		c.templates[method] = Must(template.New(filepath.Base(fileName)).Funcs(c.funcs).ParseFiles(templates...))
	}

	return c
}

func (c UserController) AvatarSet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	u, _ := middleware.CertUserFromRequest(request)

	if request.URL.RawQuery == "" {
		err = helper.InputPrompt(writer, "Enter your avatar emoji:")
		return
	}

	query, err := url.QueryUnescape(request.URL.RawQuery)
	if err != nil {
		return
	}

	avatar, ok := helper.ValidateAvatar(query)
	if !ok {
		_, err = writer.Write([]byte(Heredoc(`
			Avatar must be a single character and may be any emoji.
			=> set Try again
		`)))
		if err != nil {
			return
		}
	}

	err = c.repo.UpdateAvatar(u.UserID, avatar)
	if err != nil { // TODO: userName conflict
		return
	}

	err = helper.Redirect(writer, fmt.Sprintf("/users/%d/firstname/set", u.UserID))
}

func (c UserController) FirstNameSet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	strID, _ := middleware.StrFromRequest(request, "userID")

	intID, err := strconv.Atoi(strID)
	if err != nil {
		return
	}

	pathUserID := uint64(intID)

	if user.UserID != pathUserID {
		gemini.NotFound(writer, request)
		return
	}

	if request.URL.RawQuery == "" {
		err = helper.InputPrompt(writer, "Enter your first name:")
		return
	}

	firstName, ok := helper.ValidateName(request.URL.RawQuery)
	if !ok {
		_, err = writer.Write([]byte(Heredoc(`
			Name must be between 1 and 25 characters and may include letters, space, apostrophe and hyphen.
			=> set Try again
		`)))
		if err != nil {
			return
		}
	}

	err = c.repo.UpdateFirstName(pathUserID, firstName)
	if err != nil { // TODO: userName conflict
		return
	}

	err = helper.Redirect(writer, fmt.Sprintf("/users/%d/lastname/set", pathUserID))
}

func (c UserController) Handler(routes ...map[string]Handler) *Mux {
	handlers := opt.OfFirst(routes).Or(c.Routes())

	router := mux.NewMux()

	for pattern, h := range handlers {
		router.AddRoute(pattern, h)
	}

	return router
}

func (c UserController) LastNameSet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	u, _ := middleware.CertUserFromRequest(request)

	if request.URL.RawQuery == "" {
		err = helper.InputPrompt(writer, "Enter your last name:")
		return
	}

	lastName, ok := helper.ValidateName(request.URL.RawQuery)
	if !ok {
		_, err = writer.Write([]byte(Heredoc(`
			Name must be between 1 and 25 characters and may include letters, space, apostrophe and hyphen.
			=> set Try again
		`)))
		if err != nil {
			return
		}
	}

	err = c.repo.UpdateLastName(u.UserID, lastName)
	if err != nil { // TODO: userName conflict
		return
	}

	err = helper.Redirect(writer, fmt.Sprintf("/users/%d/profile", u.UserID))
}

func (c UserController) PasswordGet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	password, _, err := c.repo.PasswordGet(user.UserID)
	if err != nil {
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
		return
	}
}

func (c UserController) PasswordSet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	if request.URL.RawQuery == "" {
		err = helper.InputSensitive(writer, "New Password:\n(at least 8 characters, at least one upper case, lower case, digit and special character)")
		return
	}

	user, _ := middleware.CertUserFromRequest(request)

	password := request.URL.RawQuery

	p, length, upper, lower, digit, special := model.NewPassword(password)
	if !(length && upper && lower && digit && special) {
		response := bytes.Buffer{}

		write := Must1(response.WriteString)

		write(Heredoc(`
			# Insufficient password complexity

			Try again and choose a password that meets the following requirements.

			Asterisks indicate unsatisfied requirements.

			Password must be at least:

		`))

		write("* 8 characters")

		if !length {
			write(" (*)")
		}

		write("\n\nand at least one each of:\n\n")

		write("* upper case")

		if !upper {
			write(" (*)")
		}

		write("\n* lower case")

		if !lower {
			write(" (*)")
		}

		write("\n* a digit")

		if !digit {
			write(" (*)")
		}

		write("\n* a special character")

		if !special {
			write(" (*)")
		}

		write("\n\n=> set Try again\n")

		_, err = writer.Write(response.Bytes())
		if err != nil {
			return
		}
	}

	err = c.repo.PasswordSet(user.UserID, p)
	if err != nil {
		return
	}

	err = helper.Redirect(writer, fmt.Sprintf("/users/%d/profile", user.UserID))
}

func (c UserController) ProfileGet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	userID, ok := middleware.Uint64FromRequest(request, "id")
	if !ok {
		gemini.BadRequest(writer, request)
		log.Print("no user id")
		return
	}

	u, _ := middleware.CertUserFromRequest(request)

	p, cs, found, err := c.repo.ProfileGet(userID)
	if err != nil || !found {
		return
	}

	certificates := make([]helper.Certificate, len(cs))

	for i, certificate := range cs {
		certificates[i] = helper.Certificate{
			CreatedAt: model.LongHumanTime(certificate.CreatedAt),
			ExpireAt:  model.LongHumanTime(certificate.ExpireAt),
		}
	}

	profile := helper.Profile{
		Avatar:        p.Avatar,
		Certificates:  certificates,
		CreatedAt:     model.LongHumanTime(p.CreatedAt),
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		LastSeen:      model.HumanTime(p.LastSeen),
		Me:            userID == u.UserID,
		PasswordFound: p.Password.Valid,
		UserID:        fmt.Sprintf("%d", userID),
		UserName:      p.UserName,
	}

	err = c.templates["profileGet"].Execute(writer, profile)
	if err != nil {
		return
	}
}

func (c UserController) Routes() map[string]Handler {
	return map[string]Handler{
		"/users/{id}/avatar/set":    middleware.EyesOnly(HandlerFunc(c.AvatarSet)),
		"/users/{id}/firstname/set": middleware.EyesOnly(HandlerFunc(c.FirstNameSet)),
		"/users/{id}/lastname/set":  middleware.EyesOnly(HandlerFunc(c.LastNameSet)),
		"/users/{id}/password":      HandlerFunc(c.PasswordGet),
		"/users/{id}/password/set":  middleware.EyesOnly(HandlerFunc(c.PasswordSet)),
		"/users/{id}/profile":       HandlerFunc(c.ProfileGet),
		"/users/{id}/username/set":  middleware.EyesOnly(HandlerFunc(c.UserNameSet)),
	}
}

func (c UserController) ServeGemini(writer ResponseWriter, request *Request) {
	if c.handler == nil {
		c.handler = c.Handler()
	}

	c.handler.ServeGemini(writer, request)
}

func (c UserController) UserNameSet(writer ResponseWriter, request *Request) {
	var err error

	defer writeError(writer, err)

	user, _ := middleware.CertUserFromRequest(request)

	if request.URL.RawQuery == "" {
		err = helper.InputPrompt(writer, "Choose your (permanent) username:")
		return
	}

	userName, ok := helper.ValidateUserName(request.URL.RawQuery)
	if !ok {
		_, err = writer.Write([]byte(Heredoc(`
			Username must be between 5 and 50 characters with no spaces or emojis.
			=> set Try again
		`)))
		if err != nil {
			return
		}
	}

	err = c.repo.UpdateUserName(user.UserID, userName)
	if err != nil { // TODO: userName conflict
		return
	}

	err = helper.Redirect(writer, fmt.Sprintf("/users/%d/avatar/set", user.UserID))
}
