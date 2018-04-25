package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"unicode"
)

func jsonify(w http.ResponseWriter, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// convert to map then rename the keys to snake case
	m := map[string]interface{}{}
	err = json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println(err)
		return err
	}
	c := convert(m)
	payload, err = json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
	return nil
}

// taken from https://gist.github.com/elwinar/14e1e897fdbe4d3432e1#gistcomment-2246837
func ToSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) && ((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// adapted from https://gist.github.com/hvoecking/10772475
func convert(u interface{}) interface{} {
	original := reflect.ValueOf(u)
	copy := reflect.New(original.Type()).Elem()
	translateRecursive(copy, original)
	return copy.Interface()
}

func translateRecursive(copy, original reflect.Value) {
	switch original.Kind() {
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		translateRecursive(copy.Elem(), originalValue)
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return
		}
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		translateRecursive(copyValue, originalValue)
		copy.Set(copyValue)
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			translateRecursive(copy.Index(i), original.Index(i))
		}
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			if !originalValue.IsValid() {
				continue
			}
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue)

			newKey := key

			if key.Kind() == reflect.String {

				newKey = reflect.ValueOf(ToSnakeCase(key.String()))
			}
			copy.SetMapIndex(newKey, copyValue)
		}
	default:
		copy.Set(original)
	}
}
