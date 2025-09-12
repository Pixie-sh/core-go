package http

import (
	"github.com/pixie-sh/errors-go"

	pixieFiles "github.com/pixie-sh/core-go/pkg/files"
)

func ExtractFilesFromRequest(ctx ServerCtx, fileField string) (map[string]pixieFiles.File, error) {
	multipartForm, err := ctx.MultipartForm()
	if err != nil {
		return nil, errors.NewValidationError("failed to parse multipart form", &errors.FieldError{
			Field:   "multipart_form",
			Rule:    "unable_parse_multipart_form",
			Message: "failed to parse multipart form",
		}).WithNestedError(err)
	}

	files := multipartForm.File[fileField]
	if len(files) == 0 {
		return nil, errors.NewValidationError("Validation Error", &errors.FieldError{
			Field:   "multipart_form",
			Rule:    "no_files_multipart_form",
			Message: "No files found in multipart form",
		})
	}

	fileMap := make(map[string]pixieFiles.File)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, errors.NewValidationError("Validation Error", &errors.FieldError{
				Field:   "multipart_form",
				Rule:    "stream_error_multipart_form",
				Message: "unable to open file stream",
			}).WithNestedError(err)
		}

		fileMap[fileHeader.Filename] = file
	}

	return fileMap, nil
}

func CloseFilesFromRequest(ctx ServerCtx, files map[string]pixieFiles.File) error {
	for _, file := range files {
		err := file.Close()
		if err != nil {
			return errors.NewValidationError("Validation Error", &errors.FieldError{
				Field:   "multipart_form",
				Rule:    "stream_error_multipart_form",
				Message: "unable to close file stream",
			}).WithNestedError(err)
		}
	}

	return nil
}
