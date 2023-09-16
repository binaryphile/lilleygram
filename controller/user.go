package controller

import (
	"fmt"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/model"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
	"text/template"
)

type (
	UserController struct {
		get         FnHandler
		list        FnHandler
		passwordAdd FnHandler
		passwordGet FnHandler
	}
)

func NewUserController(repo sqlrepo.UserRepo, middlewares ...Middleware) UserController {
	getWith := func(repo sqlrepo.UserRepo) FnHandler {
		return func(writer ResponseWriter, request *Request) {
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
	}

	listWith := func(repo sqlrepo.UserRepo) FnHandler {
		return func(writer ResponseWriter, request *Request) {
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
	}

	passwordAddWith := func(repo sqlrepo.UserRepo) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			if request.URL.RawQuery == "" {
				err := writer.SetHeader(gemini.CodeInput, "New Password:")
				if err != nil {
					log.Println(err)
				}
				return
			}

			user, _ := UserFromContext(request.Context)

			_ = request.URL.RawQuery

			p := model.Password{}

			p.Salt = "salt"

			p.Argon2 = "argon2"

			err := repo.PasswordAdd(user.ID, p)
			if err != nil {
				log.Println(err)
			}
		}
	}

	passwordGetWith := func(repo sqlrepo.UserRepo) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			user, _ := UserFromContext(request.Context)

			ts, err := template.ParseFiles(
				"view/password.tmpl",
				"view/base.layout.tmpl",
				"view/footer.partial.tmpl",
			)
			if err != nil {
				log.Println(err)

				err := writer.SetHeader(gemini.CodeCGIError, "internal error")
				if err != nil {
					log.Panic(err)
				}

				return
			}

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
				Password model.Password
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

	return UserController{
		get:         ExtendFnHandler(getWith(repo), middlewares...),
		list:        ExtendFnHandler(listWith(repo), middlewares...),
		passwordAdd: ExtendFnHandler(passwordAddWith(repo), middlewares...),
		passwordGet: ExtendFnHandler(passwordGetWith(repo), middlewares...),
	}
}

func (c UserController) Get(writer ResponseWriter, request *Request) {
	c.get(writer, request)
}

func (c UserController) List(writer ResponseWriter, request *Request) {
	c.list(writer, request)
}

func (c UserController) PasswordAdd(writer ResponseWriter, request *Request) {
	c.passwordAdd(writer, request)
}

func (c UserController) PasswordGet(writer ResponseWriter, request *Request) {
	c.passwordGet(writer, request)
}
