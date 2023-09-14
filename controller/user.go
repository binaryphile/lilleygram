package controller

import (
	"fmt"
	"github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/opt"
	"github.com/binaryphile/lilleygram/sql"
	"log"
)

func NewUserController(userRepo sql.UserRepo, specific map[string][]Middleware, general ...Middleware) Controller {
	c := Controller{
		"get":  getUser(userRepo),
		"list": listUsers(userRepo),
	}

	for methodName, handler := range c {
		m := opt.OfIndex(specific, methodName)

		c[methodName] = ExtendHandler(handler, m.OrZero()...)

		c[methodName] = ExtendHandler(c[methodName], general...)
	}

	return c
}

func getUser(userRepo sql.UserRepo) shortcuts.HandlerFunc {
	return func(writer shortcuts.ResponseWriter, request *shortcuts.Request) {
		user, ok := UserFromContext(request.Context)
		if !ok {
			return
		}

		u, err := userRepo.Get(user.ID)
		if err != nil {
			log.Panic(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
		if err != nil {
			log.Panic(err)
		}
	}
}

func listUsers(userRepo sql.UserRepo) shortcuts.HandlerFunc {
	return func(writer shortcuts.ResponseWriter, request *shortcuts.Request) {
		user, ok := UserFromContext(request.Context)
		if !ok {
			return
		}

		u, err := userRepo.Get(user.ID)
		if err != nil {
			log.Panic(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", u.ID, u.FirstName, u.LastName, u.UserName)))
		if err != nil {
			log.Panic(err)
		}
	}
}
