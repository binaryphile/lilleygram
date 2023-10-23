package gmnifc

import (
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/gmnifc/shortcuts"
	"log"
)

var (
	BadRequest = gemini.BadRequest

	NotFound = gemini.NotFound
)

func InputPrompt(writer ResponseWriter, message string) error {
	return writer.SetHeader(gemini.CodeInput, message)
}

func InputSensitive(writer ResponseWriter, message string) error {
	return writer.SetHeader(gemini.CodeInputSensitive, message)
}

func InternalServerError(writer ResponseWriter, err error) {
	if err := writer.SetHeader(gemini.CodePermanentFailure, "Internal Server Error"); err != nil {
		log.Print(err)
	}

	log.Print(err)
}

func Redirect(writer ResponseWriter, location string) error {
	return writer.SetHeader(gemini.CodeRedirect, location)
}
