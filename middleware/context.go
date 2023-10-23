package middleware

import (
	"github.com/a-h/gemini/mux"
	"github.com/binaryphile/lilleygram/helper"
	. "github.com/binaryphile/lilleygram/middleware/shortcuts"
	"strconv"
)

const (
	keyUser contextKey = "user"

	keyDeployEnv contextKey = "deploy_env"
)

type (
	contextKey string
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

func PageTokenFromRequest(request *Request) string {
	values := request.URL.Query()

	pageTokens, ok := values["page_token"]
	if !ok {
		return ""
	}

	if len(pageTokens) == 0 {
		return ""
	}

	return pageTokens[0]
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
