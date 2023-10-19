package shortcuts

import (
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	"text/template"
)

type (
	Handler = gemini.Handler

	HandlerFunc = gemini.HandlerFunc

	Mux = mux.Mux

	Request = gemini.Request

	ResponseWriter = gemini.ResponseWriter

	Template = template.Template
)
