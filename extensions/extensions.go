package extensions

import (
	"context"
	"database/sql"
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

func ExtendFnHandler(do FnHandler, extensions ...FnHandlerExtension) FnHandler {
	extended := do

	for _, extend := range extensions {
		extended = extend(extended)
	}

	return extended
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
				rows, err := db.Query(heredoc.Doc(`
					SELECT avatar, id, user_name FROM users
					WHERE cert_id = ?
				`), certID)
				if err != nil {
					panic(err)
				}

				defer func() {
					err = rows.Close()
					if err != nil {
						log.Printf("couldn't close rows: %s", err)
					}
				}()

				var userAvatar, userName string

				var userID uint64

				for rows.Next() {
					err = rows.Scan(&userAvatar, &userID, &userName)
					if err != nil {
						panic(err)
					}
				}

				if err = rows.Err(); err != nil {
					panic(err)
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
