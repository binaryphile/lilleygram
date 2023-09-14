package controller

import (
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/middleware"
	"log"
)

func withRequiredInput(prompt string) Middleware {
	return func(handler Handler) Handler {
		return HandlerFunc(func(writer ResponseWriter, request *Request) {
			if request.URL.RawQuery == "" {
				err := writer.SetHeader(gemini.CodeInput, prompt)
				if err != nil {
					log.Println(err)
				}
				return
			}

			handler.ServeGemini(writer, request)
		})
	}
}
