package middleware

import (
	"context"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"log"
)

const (
	keyUser contextKey = "user"
)

type (
	FnAuthorize = func(certID, certKey string) (helper.User, bool)

	FnHandler = func(ResponseWriter, *Request)

	Middleware = func(FnHandler) FnHandler

	contextKey string
)

func ExtendFnHandler(handler FnHandler, extensions ...Middleware) FnHandler {
	extended := handler

	for _, extend := range extensions {
		extended = extend(extended)
	}

	return extended
}

func ExtendHandler(handler Handler, extensions ...Middleware) HandlerFunc {
	return ExtendFnHandler(handler.ServeGemini, extensions...)
}

func UserFromContext(ctx Context) (_ helper.User, ok bool) {
	user, ok := ctx.Value(keyUser).(helper.User)

	return user, ok
}

func WithOptionalAuthentication(authorizer FnAuthorize) Middleware {
	return func(handler FnHandler) FnHandler {
		return func(w ResponseWriter, request *Request) {
			certID := request.Certificate.ID

			if certID != "" {
				user, ok := authorizer(certID, request.Certificate.Key)

				if ok {
					request.Context = context.WithValue(request.Context, keyUser, user)
				}
			}

			handler(w, request)
		}
	}
}

func WithRequiredAuthentication(authorizer FnAuthorize) Middleware {
	return func(handler FnHandler) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			_, ok := UserFromContext(request.Context)
			if ok {
				handler(writer, request)
				return
			}

			certID := request.Certificate.ID

			if certID == "" {
				err := writer.SetHeader(gemini.CodeClientCertificateRequired, "client certificate required")
				if err != nil {
					log.Print(err)
				}

				return
			}

			user, ok := authorizer(certID, request.Certificate.Key)
			if !ok {
				err := writer.SetHeader(gemini.CodeClientCertificateNotAuthorised, "not authorised")
				if err != nil {
					log.Print(err)
				}

				return
			}

			request.Context = context.WithValue(request.Context, keyUser, user)

			handler(writer, request)
		}
	}
}
