package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

func ComputeSHA256(v interface{}) (string, error) {
	data, err := serializer.Serialize(v)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

func ComputeSHA512(v interface{}) (string, error) {
	data, err := serializer.Serialize(v)
	if err != nil {
		return "", err
	}
	hash := sha512.Sum512(data)
	return fmt.Sprintf("%x", hash), nil
}
