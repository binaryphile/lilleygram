package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"github.com/binaryphile/lilleygram/sql"
	"log"
)

const (
	keyUser contextKey = "user"
)

type (
	FnAuthorize = func(certID, certKey string) bool

	Middleware = func(Handler) Handler

	contextKey string
)

func ExtendHandler(handler Handler, extensions ...Middleware) Handler {
	extended := handler

	for _, extend := range extensions {
		extended = extend(extended)
	}

	return extended
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

func WithAuthentication(repo sql.UserRepo, authorizer FnAuthorize) Middleware {
	return func(handler Handler) Handler {
		return HandlerFunc(func(writer ResponseWriter, request *Request) {
			contextHandler := WithOptionalAuthentication(repo)(handler)

			gemini.RequireCertificateHandler(contextHandler, authorizer).ServeGemini(writer, request)
		})
	}
}

func WithOptionalAuthentication(repo sql.UserRepo) Middleware {
	return func(handler Handler) Handler {
		return HandlerFunc(func(w ResponseWriter, request *Request) {
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

			handler.ServeGemini(w, request)
		})
	}
}
