package controller

import (
	"database/sql"
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/extensions"
	"github.com/binaryphile/lilleygram/must/osmust"
	"text/template"
)

type Controller struct {
	db *sql.DB
}

func New(db *sql.DB) Controller {
	return Controller{
		db: db,
	}
}

func (x Controller) Home(writer gemini.ResponseWriter, request *gemini.Request) {
	if request.URL.Path != "/" {
		gemini.NotFound(writer, request)
		return
	}

	user, userOk := UserFromContext(request.Context)

	if !userOk {
		handler := gemini.FileContentHandler("gemtext/home.page.tmpl", osmust.Open("gemtext/home.page.tmpl"))

		handler.ServeGemini(writer, request)

		return
	}

	tmpl, err := template.New("home.page.user.tmpl").ParseFiles("gemtext/home.page.user.tmpl")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(writer, user)
	if err != nil {
		panic(err)
	}
}

func (x Controller) Users(writer gemini.ResponseWriter, request *gemini.Request) {
	user, _ := UserFromContext(request.Context)

	var firstName, lastName string

	err := x.db.QueryRow(heredoc.Doc(`
		SELECT first_name
			,last_name
		FROM users
		WHERE id = $1
	`), user.ID).Scan(&firstName, &lastName)
	if err != nil {
		panic(err)
	}

	_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", user.ID, firstName, lastName, user.UserName)))
	if err != nil {
		panic(err)
	}
}
