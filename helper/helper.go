package helper

import (
	"github.com/binaryphile/lilleygram/model"
	"regexp"
	"strings"
)

type Gram struct {
	model.Gram
	Me        bool
	UpdatedAt string
}

func GramFromModel(userID uint64) func(model.Gram) Gram {
	return func(m model.Gram) Gram {
		return Gram{
			Gram:      m,
			Me:        m.AuthorID == userID,
			UpdatedAt: model.HumanTime(m.UpdatedAt),
		}
	}
}

const (
	// Emoji regex pattern
	avatarPattern = `^[\x{1F300}-\x{1F5FF}\x{1F600}-\x{1F64F}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}]$`

	namePattern = `^[a-zA-ZÀ-ÿ][a-zA-ZÀ-ÿ' -]*[a-zA-ZÀ-ÿ]$`

	userNamePattern = `^[a-zA-ZÀ-ÿ][a-zA-ZÀ-ÿ'_-]*[a-zA-ZÀ-ÿ]$`
)

var (
	avatarRegex = regexp.MustCompile(avatarPattern)

	nameRegex = regexp.MustCompile(namePattern)

	userNameRegex = regexp.MustCompile(userNamePattern)
)

func ValidateAvatar(avatar string) (_ string, ok bool) {
	avatar = strings.TrimSpace(avatar)

	if len(avatar) == 0 {
		return
	}

	avatar = string(avatar[0])

	return avatar, avatarRegex.MatchString(avatar)
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
