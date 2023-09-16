package controller

import (
	"fmt"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
)

type (
	UserController struct {
		get  FnHandler
		list FnHandler
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

	return UserController{
		get:  ExtendFnHandler(getWith(repo), middlewares...),
		list: ExtendFnHandler(listWith(repo), middlewares...),
	}
}

func (c UserController) Get(writer ResponseWriter, request *Request) {
	c.get(writer, request)
}

func (c UserController) List(writer ResponseWriter, request *Request) {
	c.list(writer, request)
}
