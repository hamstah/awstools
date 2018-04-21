package common

import (
	"log"
	"os"
)

var (
	ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
)

func FatalOnError(err error) {
	if err != nil {
		ErrorLog.Fatalln(err)
	}
}

func Fatalln(message string) {
	ErrorLog.Fatalln(message)
}
