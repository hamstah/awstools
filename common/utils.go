package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func FatalOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func FatalOnErrorW(err error, msg string) {
	if err != nil {
		log.Fatalln(errors.Wrap(err, msg))
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

func FlattenMap(input map[string]interface{}, transformKey func(string) string, separator string) (map[string]string, error) {
	result := map[string]string{}

	for key, value := range input {
		key = transformKey(key)
		switch value.(type) {
		case int, string, float64, bool:
			result[key] = value.(string)
		case map[string]interface{}:
			sub, err := FlattenMap(value.(map[string]interface{}), transformKey, separator)
			if err != nil {
				return nil, err
			}
			for subKey, subValue := range sub {
				result[fmt.Sprintf("%s%s%s", key, separator, transformKey(subKey))] = subValue
			}
		default:
			return nil, errors.New("Unsupported type")
		}
	}

	return result, nil
}

func TransformKeyEnvVar(key string) string {
	return strings.Replace(strings.ToUpper(key), "-", "_", -1)
}

func FlattenEnvVarMap(input map[string]interface{}) (map[string]string, error) {
	return FlattenMap(input, TransformKeyEnvVar, "_")
}
