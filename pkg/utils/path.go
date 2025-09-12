package utils

import (
	"path/filepath"
	"strings"
)

func ConcatenatePaths(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}

	var cleanedPaths []string
	for _, path := range paths {
		cleanedPaths = append(cleanedPaths, strings.Trim(path, "/"))
	}

	return strings.Join(cleanedPaths, "/")
}

func SuffixFileNameWithKeyBetweenFileNameAndExt(filePath, key string, returnOnlyFilename bool) string {
	base := filepath.Base(filePath)

	ext := filepath.Ext(base)
	filename := strings.TrimSuffix(base, ext)

	if returnOnlyFilename {
		return filename + "." + key + ext
	}

	return strings.TrimSuffix(filePath, base) + filename + "." + key + ext

}
