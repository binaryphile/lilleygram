package osmust

import "os"

func Getenv(key string) (value string) {
	value = os.Getenv(key)

	if value == "" {
		panic("expected environment value")
	}

	return
}
