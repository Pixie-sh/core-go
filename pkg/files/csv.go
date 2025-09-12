package files

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/types/strings"
)

type CSVData struct {
	Records []CSVRecord
	Headers []string
}

type CSVRecord map[string]string

// CSVDataToStructs : It uses struct field order to fill the struct
func CSVDataToStructs[T any](cd *CSVData) ([]T, error) {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() != reflect.Struct {
		return nil, errors.New("T must be a struct type, got %v", t.Kind())
	}

	result := make([]T, 0, len(cd.Records))

	for _, record := range cd.Records {
		var item T
		structValue := reflect.ValueOf(&item).Elem()

		// Fill struct fields based on CSV column index
		numFields := structValue.NumField()
		for i := 0; i < numFields && i < len(cd.Headers); i++ {
			field := structValue.Field(i)

			if !field.CanSet() {
				return nil, errors.New("cannot set field %v", field)
			}

			csvValue := record[cd.Headers[i]]

			if field.Kind() == reflect.String {
				field.SetString(csvValue)
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func ParseCSVFromBytes(data []byte, lineNumber int) (CSVData, error) {
	delimiter, err := detectCSVDelimiter(data)
	if err != nil {
		return CSVData{}, err
	}
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	reader.Comma = delimiter

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return CSVData{}, errors.New("failed to read CSV headers: %w", err)
	}

	// Trim whitespace from headers
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}

	var records []CSVRecord

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return CSVData{}, errors.New("failed to read CSV row at line %d: %w", lineNumber, err)
		}

		// Skip empty rows
		if isEmptyRow(row) {
			lineNumber++
			continue
		}

		// Validate row length
		if len(row) != len(headers) {
			return CSVData{}, errors.New("row at line %d has %d columns, expected %d", lineNumber, len(row), len(headers))
		}

		// Create record map
		record := make(CSVRecord)
		for i, value := range row {
			if value != "" {
				record[headers[i]] = strings.TrimSpace(value)
			}
		}

		records = append(records, record)
		lineNumber++
	}

	return CSVData{
		Records: records,
		Headers: headers,
	}, nil
}

func ParseCSVFromFile(file multipart.File, lineNumber int) ([]CSVRecord, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Trim whitespace from headers
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}

	var records []CSVRecord

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row at line %d: %w", lineNumber, err)
		}

		// Skip empty rows
		if isEmptyRow(row) {
			lineNumber++
			continue
		}

		// Validate row length
		if len(row) != len(headers) {
			return nil, fmt.Errorf("row at line %d has %d columns, expected %d", lineNumber, len(row), len(headers))
		}

		// Create record map
		record := make(CSVRecord)
		for i, value := range row {
			record[headers[i]] = strings.TrimSpace(value)
		}

		records = append(records, record)
		lineNumber++
	}

	return records, nil
}

func detectCSVDelimiter(data []byte) (rune, error) {
	if len(data) == 0 {
		return ',', fmt.Errorf("empty CSV data")
	}

	sampleSize := 1024
	if len(data) < sampleSize {
		sampleSize = len(data)
	}

	sample := data[:sampleSize]
	lines := bytes.Split(sample, []byte("\n"))

	maxLines := 3
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	if maxLines == 0 {
		return ',', fmt.Errorf("no lines found in CSV")
	}

	delimiters := []rune{',', ';'}

	bestDelimiter := ','
	bestScore := -1

	for _, delimiter := range delimiters {
		score := scoreDelimiter(lines[:maxLines], delimiter)
		if score > bestScore {
			bestScore = score
			bestDelimiter = delimiter
		}
	}

	if bestScore <= 0 {
		return ',', nil
	}

	return bestDelimiter, nil
}

func scoreDelimiter(lines [][]byte, delimiter rune) int {
	if len(lines) == 0 {
		return -1
	}

	fieldCounts := make([]int, 0, len(lines))
	delimiterByte := byte(delimiter)

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		fieldCount := 1 // At least one field
		inQuotes := false

		for _, b := range line {
			if b == '"' {
				// Toggle quote state
				inQuotes = !inQuotes
			} else if b == delimiterByte && !inQuotes {
				fieldCount++
			}
		}

		fieldCounts = append(fieldCounts, fieldCount)
	}

	if len(fieldCounts) == 0 {
		return -1
	}

	if len(fieldCounts) == 1 {
		return fieldCounts[0]
	}

	firstCount := fieldCounts[0]
	consistentLines := 0
	totalFields := 0

	for _, count := range fieldCounts {
		totalFields += count
		if count == firstCount {
			consistentLines++
		}
	}

	consistencyScore := consistentLines * 10
	fieldScore := totalFields

	// Bonus for reasonable field counts (2-50 fields typical)
	if firstCount >= 2 && firstCount <= 50 {
		consistencyScore += 5
	}

	return consistencyScore + fieldScore
}

func isEmptyRow(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
