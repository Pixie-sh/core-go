package strings

import (
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func Unescape(s string) (string, error) {
	ss, err := url.QueryUnescape(s)
	if err != nil {
		return ss, err
	}

	return ReplaceUnicodeEscapes(ss), nil
}

func StringToCamel(s string) string {
	split := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_'
	})

	for i := 0; i < len(split); i++ {
		rs := []rune(split[i])
		rs[0] = unicode.ToUpper(rs[0])
		split[i] = string(rs)
	}

	return strings.Join(split, "")
}

func CropString(message string, maxChars int) string {
	if len(message) > maxChars {
		return message[:maxChars-3]
	}
	return message
}

func StripString(message string, regex string) string {
	re := regexp.MustCompile(regex)
	return re.ReplaceAllString(message, "")
}

func ReplaceUnicodeEscapes(s string) string {
	replacer := strings.NewReplacer(
		"\\u003e", ">",
		"\\u003c", "<",
		"\\u0026", "&",
	)
	return replacer.Replace(s)
}

func RemoveNonDigitCharacters(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}

func MatchRegexp(toEval, exp string) bool {
	re := regexp.MustCompile(exp)
	return re.MatchString(toEval)
}

// StringSliceToUint64 Helper function to convert []string to []uint64
func StringSliceToUint64(inputs []string) ([]uint64, error) {
	result := make([]uint64, len(inputs))
	for i, input := range inputs {
		val, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

func StripHTMLTags(s string) string {
	unescaped := html.UnescapeString(s)
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(unescaped, "")
}
