package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sqlrepo"
	"log"
)

const (
	keyUser contextKey = "user"
)

type (
	FnAuthorize = func(certID, certKey string) bool

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

func WithAuthentication(repo sqlrepo.UserRepo, authorizer FnAuthorize) Middleware {
	return func(handler FnHandler) FnHandler {
		return func(writer ResponseWriter, request *Request) {
			contextHandler := HandlerFunc(ExtendFnHandler(handler, WithOptionalAuthentication(repo)))

			gemini.RequireCertificateHandler(contextHandler, authorizer).ServeGemini(writer, request)
		}
	}
}

func WithOptionalAuthentication(repo sqlrepo.UserRepo) Middleware {
	return func(handler FnHandler) FnHandler {
		return func(w ResponseWriter, request *Request) {
			certID := request.Certificate.ID

			if certID != "" {
				hash := sha256.Sum256([]byte(certID))

				certSHA256 := hex.EncodeToString(hash[:])

				u, err := repo.GetByCertificate(certSHA256)
				if err != nil {
					log.Panic(err)
				}

				if u.UserName != "" {
					request.Context = context.WithValue(request.Context, keyUser, struct {
						Avatar   string
						ID       uint64
						UserName string
					}{
						Avatar:   u.Avatar,
						ID:       u.ID,
						UserName: u.UserName,
					})
				}
			}

			handler(w, request)
		}
	}
}
