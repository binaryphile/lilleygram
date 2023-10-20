package shortcuts

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	"github.com/binaryphile/lilleygram/middleware"
	"github.com/binaryphile/lilleygram/shortcuts"
	"text/template"
)

var (
	LocalEnvFromRequest = middleware.LocalEnvFromRequest

	Heredoc = shortcuts.Heredoc
)

type (
	Handler = gemini.Handler

	HandlerFunc = gemini.HandlerFunc

	Mux = mux.Mux

	Request = gemini.Request

	ResponseWriter = gemini.ResponseWriter

	Template = template.Template
)
