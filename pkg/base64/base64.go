package base64

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/types"
)

func Decode(base64String string) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return "", err
	}

	return types.UnsafeString(decodedData), nil
}

func MustEncode(input any) string {
	b64, err := Encode(input)
	errors.Must(err)
	return b64
}

func Encode(obj any) (string, error) {
	switch expr := obj.(type) {
	case string:
		var dst []byte = make([]byte, base64.StdEncoding.EncodedLen(len(expr)))
		base64.StdEncoding.Encode(dst, types.UnsafeBytes(expr))
		return types.UnsafeString(dst), nil
	case []byte:
		var dst []byte = make([]byte, base64.StdEncoding.EncodedLen(len(expr)))
		base64.StdEncoding.Encode(dst, expr)
		return types.UnsafeString(dst), nil
	default:
		plain, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}

		var dst []byte = make([]byte, base64.StdEncoding.EncodedLen(len(plain)))
		base64.StdEncoding.Encode(dst, plain)
		return types.UnsafeString(dst), nil
	}
}
