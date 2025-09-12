package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"

	gojson "github.com/goccy/go-json"
)

func ComputeSHA256(v interface{}) (string, error) {
	data, err := gojson.Marshal(v)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

func ComputeSHA512(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	hash := sha512.Sum512(data)
	return fmt.Sprintf("%x", hash), nil
}
