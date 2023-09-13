package shortcuts

import (
	"context"
	"net/url"

 "github.com/makenowjust/heredoc/v2"
)

var (
 Heredoc = heredoc.Doc
)

type (
	Context = context.Context

	URL = url.URL
)
