package tags

import (
	"strings"

	"github.com/pixie-sh/errors-go"

	pixieerrors "github.com/pixie-sh/core-go/pkg/errors"
)

func Explode(t Tag) (TagScope, string, error) {
	parts := strings.Split(string(t), ".")
	if len(parts) == 2 {
		return TagScope(parts[0]), parts[1], nil
	}

	return "", "", errors.New("invalid tag format; %s", t.String(), pixieerrors.TagsInvalidFormatErrorCode)
}

func HasScope(tag Tag, scope TagScope) (TagScope, string, error) {
	exScope, val, err := Explode(tag)
	if err != nil {
		return "", "", err
	}

	if exScope != scope {
		return exScope, val, errors.New("invalid tag scope; %s", tag.String(), pixieerrors.TagsInvalidScopeErrorCode)
	}

	return scope, val, nil
}
