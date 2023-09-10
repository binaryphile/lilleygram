package tlsmust

import (
	"crypto/tls"
	"github.com/binaryphile/lilleygram/must"
)

var (
	LoadX509KeyPair = must.Must2(tls.LoadX509KeyPair)
)
