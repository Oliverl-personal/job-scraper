package utils

import (
	"fmt"
	"os"
)

var NoSuchElementError = fmt.Errorf("requested element was note found")

func FatalError(err error) {
	Logger.Error(fmt.Sprintf("%v", err))
	os.Exit(1)
}
