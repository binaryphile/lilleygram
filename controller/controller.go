package controller

import (
	"database/sql"
	. "github.com/binaryphile/lilleygram/extensions"
	"github.com/binaryphile/lilleygram/opt"
)

type (
	Controller struct {
		db       *sql.DB
		handlers map[string]FnHandler
	}
)

func New(db *sql.DB, extensions map[string][]FnHandlerExtension) Controller {
	c := Controller{
		db: db,
	}

	c.handlers = map[string]FnHandler{
		"home":  home(c),
		"users": users(c),
	}

	for methodName, handler := range c.handlers {
		handlerExtensions := opt.OfIndex(extensions, methodName)

		c.handlers[methodName] = ExtendFnHandler(handler, handlerExtensions.OrZero()...)
	}

	return c
}
