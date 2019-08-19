package resources

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
)

func DecodeInlinePolicyDocument(inlineDocument string) (map[string]interface{}, error) {
	decodedDocument, err := url.QueryUnescape(inlineDocument)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode inline policy document")
	}

	document := map[string]interface{}{}
	err = json.Unmarshal([]byte(decodedDocument), &document)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse inline policy document to JSON")
	}
	return document, nil
}
