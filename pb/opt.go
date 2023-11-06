package pb

import "github.com/binaryphile/lilleygram/opt"

var (
	NoTimePageToken = OptTimePageToken{}
)

type (
	OptTimePageToken = opt.Type[*TimePageToken, uint32]
)

func TimePageTokenOfNonZero(token string) OptTimePageToken {
	return OptTimePageToken{
		Value: opt.Of(UnmarshalTimePageToken(token), token != ""),
	}
}
