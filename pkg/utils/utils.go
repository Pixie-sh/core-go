package utils

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func ReadCloserToString(rc io.ReadCloser) (string, error) {
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)

	var sb strings.Builder
	_, err := io.Copy(&sb, rc)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// FormatDuration format to string a duration accordingly to:
// 1m 30s
// 1h 30m 15s
// 1d 50h 3m 1s
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	seconds := int(d.Seconds())

	days := seconds / (24 * 3600)
	seconds %= 24 * 3600

	hours := seconds / 3600
	seconds %= 3600

	minutes := seconds / 60
	seconds %= 60

	var sb strings.Builder
	sb.Grow(20)

	var needSpace bool

	if days > 0 {
		sb.WriteString(fmt.Sprintf("%dd", days))
		needSpace = true
	}
	if hours > 0 {
		if needSpace {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%dh", hours))
		needSpace = true
	}
	if minutes > 0 {
		if needSpace {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%dm", minutes))
		needSpace = true
	}
	if seconds > 0 || sb.Len() == 0 {
		if needSpace {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%ds", seconds))
	}

	return sb.String()
}

// OrderStrings sorts elements from the source array based on the order
// they appear in the reference array. Elements not in the reference array
// are appended at the end in their original order.
func OrderStringsSlices(source, reference []string) []string {
	// Create a map to track elements in the source array
	sourceMap := make(map[string]bool, len(source))
	for _, item := range source {
		sourceMap[item] = true
	}

	// Create a result slice with enough capacity
	ordered := make([]string, 0, len(source))

	// Track which elements have been processed
	processed := make(map[string]bool, len(source))

	// First add all elements from reference that exist in source
	for _, item := range reference {
		if sourceMap[item] && !processed[item] {
			ordered = append(ordered, item)
			processed[item] = true
		}
	}

	// Add remaining elements that weren't in reference
	for _, item := range source {
		if !processed[item] {
			ordered = append(ordered, item)
		}
	}

	return ordered
}

func ConvertJSONNumbersToUint64(i interface{}) interface{} {
	switch x := i.(type) {
	case map[string]interface{}:
		for k, v := range x {
			x[k] = ConvertJSONNumbersToUint64(v)
		}
		return x
	case []interface{}:
		for i, v := range x {
			x[i] = ConvertJSONNumbersToUint64(v)
		}
		return x
	case float64:
		// If the float has no decimal part, convert to uint64
		if x == float64(uint64(x)) {
			return uint64(x)
		}
		return x
	default:
		return i
	}
}
