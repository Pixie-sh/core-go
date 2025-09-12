package lambda_api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type ExtractedFile struct {
	FieldName   string            `json:"field_name"`   // Form field name
	Filename    string            `json:"filename"`     // Original filename
	ContentType string            `json:"content_type"` // MIME type
	Size        int64             `json:"size"`         // File size in bytes
	Content     []byte            `json:"-"`            // File content (excluded from JSON)
	Headers     map[string]string `json:"headers"`      // Additional file headers
}

type FileExtractionResult struct {
	Files      []ExtractedFile   `json:"files"`
	TotalSize  int64             `json:"total_size"`
	FileCount  int               `json:"file_count"`
	FormFields map[string]string `json:"form_fields"` // Non-file form fields
	Errors     []string          `json:"errors,omitempty"`
}

func ExtractFilesFromMultipart(ctx *pixiecontext.LambdaAPIContext) []ExtractedFile {
	ctxLogger := pixiecontext.GetCtxLogger(ctx)
	ctxLogger.With("body-start", ctx.RequestV2.Body[:min(100, len(ctx.RequestV2.Body))]).
		With("isBase64Encoded", ctx.RequestV2.IsBase64Encoded).
		Debug("logging req infos")

	var bodyBytes = types.UnsafeBytes(ctx.RequestV2.Body)
	// Check if the body is base64 encoded (common in AWS Lambda)
	if ctx.RequestV2.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(ctx.RequestV2.Body)
		if err != nil {
			panic(fmt.Sprintf("failed to decode base64 body: %v", err))
		}
		bodyBytes = decoded
	} else {
		bodyBytes = types.UnsafeBytes(ctx.RequestV2.Body)
	}

	var extractedFiles = make([]ExtractedFile, 0)
	contentType := getContentType(ctx.RequestV2.Headers)

	if !strings.Contains(contentType, "multipart/form-data") {
		panic("not a multipart form")
	}

	boundary := extractBoundary(contentType)
	if boundary == "" {
		panic("invalid boundary")
	}

	parts := bytes.Split(bodyBytes, []byte("--"+boundary))
	for _, part := range parts {
		if len(part) <= 2 { // Skip empty parts and boundary markers
			continue
		}

		// Skip final boundary marker
		if bytes.HasPrefix(bytes.TrimSpace(part), []byte("--")) {
			continue
		}

		extractedFile, err := processMultipartPart(part)
		if err != nil {
			continue
		}

		extractedFiles = append(extractedFiles, extractedFile)
	}

	return extractedFiles
}

func processMultipartPart(part []byte) (ExtractedFile, error) {
	// Find header section
	headerEnd := bytes.Index(part, []byte("\r\n\r\n"))
	if headerEnd == -1 {
		headerEnd = bytes.Index(part, []byte("\n\n"))
		if headerEnd == -1 {
			return ExtractedFile{}, errors.New("invalid multipart part format")
		}
	}

	headerSection := part[:headerEnd]
	content := part[headerEnd+4:] // Skip \r\n\r\n or \n\n

	// Remove trailing CRLF
	content = bytes.TrimSuffix(content, []byte("\r\n"))
	content = bytes.TrimSuffix(content, []byte("\n"))

	// Parse headers
	headers := parseMultipartHeaders(string(headerSection))

	// Get Content-Disposition
	disposition := headers["content-disposition"]
	if disposition == "" {
		return ExtractedFile{}, errors.New("missing Content-Disposition header")
	}

	// Parse Content-Disposition
	fieldName, filename, err := parseContentDisposition(disposition)
	if err != nil {
		return ExtractedFile{}, err
	}

	// If no filename, treat as form field
	if filename == "" {
		filename = uid.NewUUID()
	}

	// Create extracted file
	return ExtractedFile{
		FieldName:   fieldName,
		Filename:    filename,
		ContentType: headers["content-type"],
		Size:        int64(len(content)),
		Content:     content,
		Headers:     headers,
	}, nil
}

func extractBoundary(contentType string) string {
	parts := strings.Split(contentType, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "boundary=") {
			boundary := strings.TrimPrefix(part, "boundary=")
			// Remove quotes if present
			if strings.HasPrefix(boundary, `"`) && strings.HasSuffix(boundary, `"`) {
				boundary = strings.Trim(boundary, `"`)
			}
			return boundary
		}
	}
	return ""
}

func parseMultipartHeaders(headerSection string) map[string]string {
	headers := make(map[string]string)
	lines := strings.Split(headerSection, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}

func parseContentDisposition(disposition string) (fieldName, filename string, err error) {
	parts := strings.Split(disposition, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "name=") {
			fieldName = strings.Trim(strings.TrimPrefix(part, "name="), `"`)
		} else if strings.HasPrefix(part, "filename=") {
			filename = strings.Trim(strings.TrimPrefix(part, "filename="), `"`)
		}
	}

	if fieldName == "" {
		return "", "", fmt.Errorf("field name not found in Content-Disposition")
	}

	return fieldName, filename, nil
}

func getContentType(headers map[string]string) string {
	for _, key := range []string{"Content-Type", "content-type", "CONTENT-TYPE"} {
		if ct := headers[key]; ct != "" {
			return ct
		}
	}
	return ""
}
