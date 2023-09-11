package osmust

import (
	"fmt"
	"os"
)

func Getenv(key string) (value string) {
	value = os.Getenv(key)

	if value == "" {
		panic(fmt.Errorf("expected environment value for key %s", key))
	}

	return
}

func Open(name string) *os.File {
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}

	return file
}
