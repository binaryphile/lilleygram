package middleware

import (
	"github.com/a-h/gemini/mux"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware/shortcuts"
	"strconv"
)

type (
	contextKey string
)

const (
	keyUser      contextKey = "user"
	keyDeployEnv contextKey = "deploy_env"
)

func CertUserFromRequest(r *Request) (_ helper.User, ok bool) {
	user, ok := r.Context.Value(keyUser).(helper.User)
	if !ok {
		return
	}

	return user, ok
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
