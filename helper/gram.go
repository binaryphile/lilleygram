package helper

import (
	"fmt"
	"github.com/binaryphile/lilleygram/model"
)

type Gram struct {
	ID        string
	Avatar    string
	Gram      string
	UserName  string
	UpdatedAt string
}

func GramFromModel(m model.Gram) Gram {
	return Gram{
		ID:        fmt.Sprintf("%d", m.ID),
		Avatar:    m.Avatar,
		Gram:      m.Gram,
		UserName:  m.UserName,
		UpdatedAt: model.HumanTime(m.UpdatedAt),
	}
}
