package shortcuts

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	"github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/must"
	"github.com/binaryphile/lilleygram/shortcuts"
	"text/template"
)

var (
	ExtendHandler = middleware.ExtendHandler

	Heredoc = shortcuts.Heredoc

	Must = must.Must[*Template]

	Must1 = must.Must1[string, int]

	LocalEnvFromRequest = middleware.LocalEnvFromRequest

	WithRefresh = middleware.WithRefresh
)

type (
	Handler = gemini.Handler

	HandlerFunc = gemini.HandlerFunc

	Mux = mux.Mux

	Request = gemini.Request

	ResponseWriter = gemini.ResponseWriter

	Template = template.Template
)
