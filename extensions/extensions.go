package extensions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/shortcuts"
	"log"
)

const (
	keyUser contextKey = "user"
)

type (
	FnHandler = func(gemini.ResponseWriter, *gemini.Request)

	FnHandlerExtension = func(FnHandler) FnHandler

	contextKey string
)

func ExtendFnHandler(handler FnHandler, extensions ...FnHandlerExtension) FnHandler {
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

func WithAuthentication(db *sql.DB, authorizer func(certID, certKey string) bool) FnHandlerExtension {
	return func(handler FnHandler) FnHandler {
		return func(writer gemini.ResponseWriter, request *gemini.Request) {
			contextHandler := gemini.HandlerFunc(WithOptionalAuthentication(db)(handler))

			gemini.RequireCertificateHandler(contextHandler, authorizer).ServeGemini(writer, request)
		}
	}
}

func WithOptionalAuthentication(db *sql.DB) FnHandlerExtension {
	return func(handler FnHandler) FnHandler {
		return func(w gemini.ResponseWriter, request *gemini.Request) {
			certID := request.Certificate.ID

			if certID != "" {
				hash := sha256.Sum256([]byte(certID))

				hexHash := hex.EncodeToString(hash[:])

				var userAvatar, userName string

				var userID uint64

				err := db.QueryRow(heredoc.Doc(`
					SELECT avatar, users.user_id, user_name FROM users
					INNER JOIN certificates ON users.user_id = certificates.user_id
					WHERE cert_sha256 = $1
				`), hexHash).Scan(&userAvatar, &userID, &userName)
				if err != nil {
					log.Panicf("couldn't query user from db: %s", err)
				}

				if userName != "" {
					request.Context = context.WithValue(request.Context, keyUser, struct {
						Avatar   string
						ID       uint64
						UserName string
					}{
						Avatar:   userAvatar,
						ID:       userID,
						UserName: userName,
					})
				}
			}

			handler(w, request)
		}
	}
}
