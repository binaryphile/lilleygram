package controller

import (
	"fmt"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/sql"
	"log"
)

type (
	UserController struct {
		handlers map[string]Handler
		repo     sql.UserRepo
	}
)

func NewUserController(repo sql.UserRepo, specific map[string][]Middleware, general ...Middleware) UserController {
	c := UserController{
		handlers: make(map[string]Handler),
		repo:     repo,
	}

	handlers := map[string]Handler{
		"get":  HandlerFunc(c.Get),
		"list": HandlerFunc(c.List),
	}

	for key, handler := range handlers {
		s := opt.OfIndex(specific, key)

		c.handlers[key] = ExtendHandler(handler, s.OrZero()...)

		c.handlers[key] = ExtendHandler(c.handlers[key], general...)
	}

	return c
}

func (c UserController) Get(writer ResponseWriter, request *Request) {
	user, ok := UserFromContext(request.Context)
	if !ok {
		return
	}

	u, err := c.repo.Get(user.ID)
	if err != nil {
		log.Panic(err)
	}

	_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
	if err != nil {
		log.Panic(err)
	}
}

func (c UserController) List(writer ResponseWriter, request *Request) {
	user, ok := UserFromContext(request.Context)
	if !ok {
		return
	}

	u, err := c.repo.Get(user.ID)
	if err != nil {
		log.Panic(err)
	}

	_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
	if err != nil {
		log.Panic(err)
	}
}
