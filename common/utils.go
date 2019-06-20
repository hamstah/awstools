package common

import (
	log "github.com/sirupsen/logrus"
)

func FatalOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func Fatalln(message string) {
	log.Fatalln(message)
}
