package middleware

import (
	"context"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"log"
)

const (
	keyUser contextKey = "user"
)

type (
	FnAuthorize = func(certID, certKey string) (struct {
		Avatar   string
		ID       uint64
		UserName string
	}, bool)

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

func ExtendRouter(handler Handler, extensions ...Middleware) HandlerFunc {
	return ExtendFnHandler(handler.ServeGemini, extensions...)
}

func UserFromContext(ctx Context) (_ struct {
	Avatar   string
	ID       uint64
	UserName string
}, ok bool) {
	user, ok := ctx.Value(keyUser).(struct {
		Avatar   string
		ID       uint64
		UserName string
	})

	return user, ok
}

func WithAuthentication(authorizer FnAuthorize) Middleware {
	return func(handler FnHandler) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			certID := request.Certificate.ID

			if certID == "" {
				err := writer.SetHeader(gemini.CodeClientCertificateRequired, "client certificate required")
				if err != nil {
					log.Panic(err)
				}

				return
			}

			user, found := authorizer(certID, request.Certificate.Key)
			if !found {
				err := writer.SetHeader(gemini.CodeClientCertificateNotAuthorised, "not authorised")
				if err != nil {
					log.Panic(err)
				}

				return
			}

			request.Context = context.WithValue(request.Context, keyUser, user)

			HandlerFunc(handler).ServeGemini(writer, request)
		}
	}
}

func WithOptionalAuthentication(authorizer FnAuthorize) Middleware {
	return func(handler FnHandler) FnHandler {
		return func(w ResponseWriter, request *Request) {
			certID := request.Certificate.ID

			if certID != "" {
				user, found := authorizer(certID, request.Certificate.Key)

				if found {
					request.Context = context.WithValue(request.Context, keyUser, user)
				}
			}

			handler(w, request)
		}
	}
}
