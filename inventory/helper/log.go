package helper

import (
	"log"
	"os"

	"github.com/spf13/cast"
)

var DEBUG = cast.ToBool(os.Getenv("DEBUG"))

func Println(args ...interface{}) {
	if DEBUG {
		log.Println(args...)
	}
}
