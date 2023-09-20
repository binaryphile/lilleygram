package controller

import (
	"encoding/hex"
	"fmt"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"github.com/binaryphile/lilleygram/view"
	"log"
	"text/template"
)

type (
	UserController struct {
		CertificateAdd  Handler
		CertificateList Handler
		Get             Handler
		List            Handler
		PasswordAdd     Handler
		PasswordGet     Handler
	}
)

func NewUserController(repo sqlrepo.UserRepo, middlewares ...map[string][]Middleware) UserController {
	certificateAdd := func(writer ResponseWriter, request *Request) {
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

		_, err := repo.CertificateAdd(sha256, 0, user.ID)
		if err != nil {
			log.Println(err)
		}
	}

	var certificateList FnHandler

	{
		ts, err := template.ParseFiles(
			"view/certificate.index.tmpl",
			"view/base.layout.tmpl",
			"view/footer.partial.tmpl",
		)
		if err != nil {
			log.Panic(err)
		}

		certificateList = func(writer ResponseWriter, request *Request) {
			user, _ := UserFromContext(request.Context)

			certificates, err := repo.CertificateListByUser(user.ID)
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
				Certificates []view.Certificate
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
	}

	get := func(writer ResponseWriter, request *Request) {
		user, ok := UserFromContext(request.Context)
		if !ok {
			return
		}

		u, err := repo.Get(user.ID)
		if err != nil {
			log.Panic(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
		if err != nil {
			log.Panic(err)
		}
	}

	list := func(writer ResponseWriter, request *Request) {
		user, ok := UserFromContext(request.Context)
		if !ok {
			return
		}

		u, err := repo.Get(user.ID)
		if err != nil {
			log.Panic(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
		if err != nil {
			log.Panic(err)
		}
	}

	passwordAdd := func(writer ResponseWriter, request *Request) {
		if request.URL.RawQuery == "" {
			err := writer.SetHeader(gemini.CodeInputSensitive, "New Password:")
			if err != nil {
				log.Println(err)
			}
			return
		}

		user, _ := UserFromContext(request.Context)

		_ = request.URL.RawQuery

		p := view.Password{
			Argon2: "argon2",
			Salt:   "salt",
		}

		err := repo.PasswordAdd(user.ID, p)
		if err != nil {
			log.Println(err)
		}
	}

	var passwordGet FnHandler

	{
		ts, err := template.ParseFiles(
			"view/password.tmpl",
			"view/base.layout.tmpl",
			"view/footer.partial.tmpl",
		)
		if err != nil {
			log.Panic(err)
		}

		passwordGet = func(writer ResponseWriter, request *Request) {
			user, _ := UserFromContext(request.Context)

			password, err := repo.PasswordGet(user.ID)
			if err != nil {
				return
			}

			type User struct {
				Avatar   string
				ID       uint64
				UserName string
			}

			data := struct {
				Password view.Password
				User     User
			}{
				Password: password,
				User:     user,
			}

			err = ts.Execute(writer, data)
			if err != nil {
				log.Panicf("couldn't execute template: %s", err)
			}
		}
	}

	handlerNames := []string{
		"certificateAdd",
		"certificateList",
		"get",
		"list",
		"passwordAdd",
		"passwordGet",
	}

	// this would be easier if functions were comparable
	for _, name := range handlerNames {
		switch name {
		case "certificateAdd":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[name]; ok {
					certificateAdd = ExtendFnHandler(certificateAdd, extensions...)
				}
			}
		case "certificateList":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[name]; ok {
					certificateList = ExtendFnHandler(certificateList, extensions...)
				}
			}
		case "get":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[name]; ok {
					get = ExtendFnHandler(get, extensions...)
				}
			}
		case "list":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[name]; ok {
					list = ExtendFnHandler(list, extensions...)
				}
			}
		case "passwordAdd":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[name]; ok {
					passwordAdd = ExtendFnHandler(passwordAdd, extensions...)
				}
			}
		case "passwordGet":
			for _, middleware := range middlewares {
				if extensions, ok := middleware[name]; ok {
					passwordGet = ExtendFnHandler(passwordGet, extensions...)
				}
			}
		}
	}

	return UserController{
		CertificateAdd:  HandlerFunc(certificateAdd),
		CertificateList: HandlerFunc(certificateList),
		Get:             HandlerFunc(get),
		List:            HandlerFunc(list),
		PasswordAdd:     HandlerFunc(passwordAdd),
		PasswordGet:     HandlerFunc(passwordGet),
	}
}

func (c UserController) Router() *mux.Mux {
	router := mux.NewMux()

	routes := map[string]Handler{
		"/users":                           c.List,
		"/users/{userID}":                  c.Get,
		"/users/{userID}/certificates":     c.CertificateList,
		"/users/{userID}/certificates/add": c.CertificateAdd,
		"/users/{userID}/password":         c.PasswordGet,
		"/users/{userID}/password/add":     c.PasswordAdd,
	}

	for pattern, handler := range routes {
		router.AddRoute(pattern, handler)
	}

	return router
}
