package helper

import (
	"github.com/a-h/gemini"
	. "github.com/binaryphile/lilleygram/controller/shortcuts"
	"log"
	"regexp"
	"strings"
)

const (
	namePattern = "^[a-zA-ZÀ-ÿ][a-zA-ZÀ-ÿ' -]*[a-zA-ZÀ-ÿ]$"

	userNamePattern = "^[a-zA-ZÀ-ÿ][a-zA-ZÀ-ÿ'_-]*[a-zA-ZÀ-ÿ]$"
)

var (
	nameRegex = regexp.MustCompile(namePattern)

	userNameRegex = regexp.MustCompile(userNamePattern)
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

func ValidateName(name string) (_ string, ok bool) {
	name = strings.TrimSpace(name)

	if len(name) < 1 || len(name) > 25 {
		return
	}

	return name, nameRegex.MatchString(name)
}

func ValidateUserName(name string) (_ string, ok bool) {
	name = strings.TrimSpace(name)

	if len(name) < 5 || len(name) > 50 {
		return
	}

	return name, userNameRegex.MatchString(name)
}
