package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func marshal(u interface{}) ([]byte, error) {
	out := map[string]interface{}{}
	t := reflect.TypeOf(u)
	v := reflect.ValueOf(u)
	fmt.Println(t)
	fmt.Println(v)
	outName := ""
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		switch n := f.Tag.Get("json"); n {
		case "first":
			outName = "name"
		case "":
			outName = strings.ToLower(f.Name)
		case "-":
			outName = ""
		default:
			outName = n
		}

		if outName != "" {
			out[outName] = v.Field(i).Interface()
		}
	}

	return json.Marshal(out)
}
