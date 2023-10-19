package middleware

import (
	"context"
	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"github.com/binaryphile/lilleygram/helper"
	"log"
	"strconv"
)

const (
	keyUser      contextKey = "user"
	keyDeployEnv contextKey = "deploy_env"
)

type (
	FnAuthorize = func(certID, certKey string) (helper.User, bool)

	Middleware = func(Handler) Handler

	contextKey string
)

func CertUserFromRequest(r *Request) (_ helper.User, ok bool) {
	user, ok := r.Context.Value(keyUser).(helper.User)
	if !ok {
		return
	}

	return user, ok
}

func DeployEnvFromRequest(r *Request) (_ string, ok bool) {
	deployEnv, ok := r.Context.Value(keyDeployEnv).(string)
	if !ok {
		return
	}

	return deployEnv, ok
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
			gemini.NotFound(writer, request)
			return
		}

		handler.ServeGemini(writer, request)
	})
}

func StrFromRequest(request *Request, key string) (_ string, ok bool) {
	route, ok := mux.GetMatchedRoute(request.Context)
	if !ok {
		return
	}

	strVar, ok := route.PathVars[key]
	if !ok {
		return
	}

	return strVar, true
}

func Uint64FromRequest(request *Request, key string) (_ uint64, ok bool) {
	strVar, ok := StrFromRequest(request, key)
	if !ok {
		return
	}

	intVar, err := strconv.Atoi(strVar)
	if err != nil {
		return
	}

	return uint64(intVar), true
}

func WithLocalDeployEnv(handler Handler) Handler {
	return HandlerFunc(func(w ResponseWriter, request *Request) {
		request.Context = context.WithValue(request.Context, keyDeployEnv, "local")

		handler.ServeGemini(w, request)
	})
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
