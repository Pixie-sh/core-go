package files

import (
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type File = multipart.File

func GetMimeType(fileName string) string {
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}

func GetExtension(fileName string, withLeadingDot bool) string {
	leadingDotExt := strings.ToLower(filepath.Ext(fileName))
	if withLeadingDot {
		return leadingDotExt
	}
	return strings.TrimPrefix(leadingDotExt, ".")
}
