package controller

import (
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/extensions"
	"github.com/binaryphile/lilleygram/opt"
	"log"
	"text/template"
)

func home(_ Controller) FnHandler {
	return func(writer gemini.ResponseWriter, request *gemini.Request) {
		if request.URL.Path != "/" {
			gemini.NotFound(writer, request)
			return
		}

		user := opt.Of(UserFromContext(request.Context))

		fileNames := []string{
			opt.OkOrNot(user, "gemtext/home.page.user.tmpl", "gemtext/home.page.tmpl"),
			"gemtext/base.layout.tmpl",
			"gemtext/footer.partial.tmpl",
		}

		ts, err := template.ParseFiles(fileNames...)
		if err != nil {
			log.Println(err.Error())

			err := writer.SetHeader(gemini.CodeCGIError, "internal error")
			if err != nil {
				log.Println(err.Error())
				panic(err)
			}
		}

		err = ts.Execute(writer, user.OrZero())
		if err != nil {
			log.Panicf("couldn't execute template: %s", err)
		}
	}
}

func users(c Controller) FnHandler {
	return func(writer gemini.ResponseWriter, request *gemini.Request) {
		user, _ := UserFromContext(request.Context)

		var firstName, lastName string

		err := c.db.QueryRow(heredoc.Doc(`
			SELECT first_name, last_name
			FROM users
			WHERE user_id = $1
		`), user.ID).Scan(&firstName, &lastName)
		if err != nil {
			log.Panicf("couldn't query users: %s", err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("%d - %s - %s - %s", user.ID, firstName, lastName, user.UserName)))
		if err != nil {
			log.Panicf("couldn't write response: %s", err)
		}
	}
}
