package shortcuts

import (
	"context"
	"github.com/MakeNowJust/heredoc/v2"
	"net/url"
)

var (
	Heredoc = heredoc.Doc
)

type (
	Context = context.Context

	URL = url.URL
)
