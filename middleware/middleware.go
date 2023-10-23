package middleware

import (
	"context"
	"github.com/a-h/gemini"
	"github.com/binaryphile/lilleygram/gmnifc"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware/shortcuts"
	"github.com/binaryphile/lilleygram/opt"
	"log"
)

type (
	FnAuthorize = func(certID, certKey string) (helper.User, bool)

	Middleware = func(Handler) Handler
)

func LocalEnvFromRequest(r *Request) opt.Bool {
	return opt.Apply(equals("local"), opt.OfAssert[string](r.Context.Value(keyDeployEnv)))
}

func ExtendHandler(handler Handler, extensions ...Middleware) Handler {
	extended := handler

	for _, extend := range extensions {
		extended = extend(extended)
	}

	return extended
}

func EyesOnly(handler Handler) Handler {
	return HandlerFunc(func(writer ResponseWriter, request *Request) {
		user, _ := CertUserFromRequest(request)

		userID, _ := Uint64FromRequest(request, "id")

		if user.UserID != userID {
			gmnifc.NotFound(writer, request)
			return
		}

		handler.ServeGemini(writer, request)
	})
}

func WithLocalDeployEnv(handler Handler) Handler {
	return HandlerFunc(func(w ResponseWriter, request *Request) {
		request.Context = context.WithValue(request.Context, keyDeployEnv, "local")

		handler.ServeGemini(w, request)
	})
}

func WithRefresh(refresh func()) Middleware {
	return func(handler Handler) Handler {
		return HandlerFunc(func(writer ResponseWriter, request *Request) {
			LocalEnvFromRequest(request).AndDo(refresh)

			handler.ServeGemini(writer, request)
		})
	}
}

func WithOptionalAuthentication(authorizer FnAuthorize) Middleware {
	return func(handler Handler) Handler {
		return HandlerFunc(func(w ResponseWriter, request *Request) {
			certID := request.Certificate.ID

			if certID != "" {
				user, ok := authorizer(certID, request.Certificate.Key)

				if ok {
					request.Context = context.WithValue(request.Context, keyUser, user)
				}
			}

			handler.ServeGemini(w, request)
		})
	}
}

func WithRequiredAuthentication(authorizer FnAuthorize) Middleware {
	return func(handler Handler) Handler {
		return HandlerFunc(func(writer ResponseWriter, request *Request) {
			_, ok := CertUserFromRequest(request)
			if ok {
				handler.ServeGemini(writer, request)
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

			handler.ServeGemini(writer, request)
		})
	}
}

func equals[T comparable](orig T) func(T) bool {
	return func(t T) bool {
		return t == orig
	}
}
