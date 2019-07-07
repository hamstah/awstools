package common

import (
	"encoding/json"
	"io/ioutil"

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

func LoadJSON(filename string, res interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, res)
}
