package configuration

import (
	"bytes"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	gojson "github.com/goccy/go-json"
	"github.com/joho/godotenv"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/types"
)

// Regular expression to match both quoted and unquoted ${shared.path.to.node} and ${#ref.path.to.node} patterns
// This will match: "session_cache": ${#ref.singleton} or "session_cache": "${#ref.singleton}"
var sharedBlocks = regexp.MustCompile(`["']?(\$\{(#ref\.[^}]+)\})["']?`)

// Regular expression to match ${env.VAR_NAME} or ${#env.VAR_NAME} patterns
// This will match environment variable references in both forms:
// - ${env.MY_VAR} - standard environment variable reference
// - ${#env.MY_VAR} - environment variable reference with # prefix
var envRegex = regexp.MustCompile(`\$\{(#?)env\.([A-Za-z0-9_.]+)\}`)

// Pattern to match quoted JSON objects or arrays
// This matches: "{"key":"value"}" or "[1,2,3]"
// Pattern to match quoted JSON objects or arrays - now handles nested structures
var jsonPattern = regexp.MustCompile(`"(\{.*?\}|\[.*?\])"`)

// SetEnvFromFile load env in provided file
func SetEnvFromFile(envFilePath string) error {
	return godotenv.Overload(envFilePath)
}

// SetEnvFromString load env in provided file
func SetEnvFromString(envData string) error {
	envs, err := godotenv.Unmarshal(envData)
	if err != nil {
		return err
	}

	for k, v := range envs {
		_ = os.Setenv(k, v)
	}

	return nil
}

// StructFromFileWithEnvReplace loads file data into memory and replace with env variables
// looking for the default pattern ${env.XXXXXX}
func StructFromFileWithEnvReplace(filePath string, holder interface{}, log logger.Interface) ([]byte, error) {
	originalFileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.With("error", err).Error("error reading configuration file %s", filePath)
		return nil, err
	}

	var fileContent []byte
	if strings.HasSuffix(filePath, ".json") {
		fileContent, err = StructFromJSONBytesWithEnvReplace(originalFileContent, holder, log)
	} else if strings.HasSuffix(filePath, ".toml") {
		fileContent, err = StructFromTOMLBytesWithEnvReplace(originalFileContent, holder, log)
	}

	if err != nil {
		log.With("raw", originalFileContent).With("error", err).Error("error reading parsed config")
		return nil, err
	}

	return fileContent, nil
}

// StructFromTOMLBytesWithEnvReplace with toml []byte into and replace with env variables
// looking for the default pattern ${env.XXXXXX}
func StructFromTOMLBytesWithEnvReplace(fileContent []byte, holder interface{}, log logger.Interface) ([]byte, error) {
	modifiedContent := replaceEnvVarsInContent(fileContent, log)
	err := fromTomlBytes(modifiedContent, holder)
	if err != nil {
		log.With("error", err).Error("error reading parsed toml")
		return nil, err
	}

	return modifiedContent, nil
}

// StructFromJSONBytesWithEnvReplace with json []byte into and replace with env variables
// looking for the default pattern ${env.XXXXXX}
// looking for the default pattern ${#ref.YYYYY}
// and prioritize environment variables over JSON values for expected struct tags
func StructFromJSONBytesWithEnvReplace(fileContent []byte, holder interface{}, log logger.Interface) ([]byte, error) {
	modifiedContent := replaceEnvVarsInContent(fileContent, log)
	modifiedContent = fixQuotedJSONObjects(types.UnsafeString(modifiedContent))

	replacedJson, err := replaceRefBlocks(types.UnsafeString(modifiedContent), log)
	if err != nil {
		log.With("raw_with_error", modifiedContent).Error("error replacing ref blocks")
		return nil, errors.NewWithError(err, "error replacing shared blocks")
	}

	fileContent, err = applyEnvPriorityOverrides(types.UnsafeBytes(replacedJson), holder, log)
	if err != nil {
		log.With("raw_with_error", modifiedContent).Error("error applying environment priority overrides")
		return nil, errors.Wrap(err, "error applying environment priority overrides")
	}

	err = fromJSONBytes(fileContent, holder)
	if err != nil {
		log.
			With("error", err).
			With("modified_content", modifiedContent).
			With("content", fileContent).
			With("env", os.Environ()).
			Error("error reading parsed json")

		return nil, err
	}

	return fileContent, nil
}

// applyEnvPriorityOverrides applies environment variable values to override JSON values
// for expected struct tags that exist in the environment
func applyEnvPriorityOverrides(jsonContent []byte, holder interface{}, log logger.Interface) ([]byte, error) {
	expectedTagsWithTypes := collectJSONTagsWithTypes(reflect.ValueOf(holder))

	var jsonMap map[string]any
	err := gojson.Unmarshal(jsonContent, &jsonMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON for env priority overrides: %s", err.Error())
	}

	for _, tagInfo := range expectedTagsWithTypes {
		envValue := os.Getenv(tagInfo.Tag)
		if envValue != "" {
			convertedValue, err := convertEnvValueToType(envValue, tagInfo.Type)
			if err != nil {
				log.With("tag", tagInfo.Tag).With("env_value", envValue).With("error", err).Warn("failed to convert environment variable to required type")
				continue
			}

			err = setJSONPathValue(jsonMap, strings.Split(tagInfo.Tag, "."), convertedValue)
			if err != nil {
				log.With("tag", tagInfo.Tag).With("error", err).Warn("failed to set environment override for tag")
			}
		}
	}

	// Marshal back to JSON
	modifiedJSON, err := gojson.Marshal(jsonMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal JSON after env priority overrides; %s", err.Error())
	}

	return modifiedJSON, nil
}

// tagInfo holds information about a struct field's JSON tag and type
type tagInfo struct {
	Tag  string
	Type reflect.Type
}

// collectJSONTagsWithTypes collects JSON tags along with their corresponding field types
func collectJSONTagsWithTypes(v reflect.Value) []tagInfo {
	var tagInfos []tagInfo
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return tagInfos
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		tag = strings.Split(tag, ",")[0]
		fieldValue := v.Field(i)

		if fieldValue.Kind() == reflect.Struct {
			// Add the struct field itself as a tag (for direct JSON replacement)
			tagInfos = append(tagInfos, tagInfo{
				Tag:  tag,
				Type: field.Type,
			})

			// Also collect nested tags for individual field access
			nestedTagInfos := collectJSONTagsWithTypes(fieldValue)
			for _, nestedTagInfo := range nestedTagInfos {
				tagInfos = append(tagInfos, tagInfo{
					Tag:  tag + "." + nestedTagInfo.Tag,
					Type: nestedTagInfo.Type,
				})
			}
		} else {
			tagInfos = append(tagInfos, tagInfo{
				Tag:  tag,
				Type: field.Type,
			})
		}
	}

	return tagInfos
}

// convertEnvValueToType converts a string environment variable value to the specified type
func convertEnvValueToType(envValue string, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.String:
		return envValue, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(envValue, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(envValue, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(envValue, 64)
	case reflect.Bool:
		return strconv.ParseBool(envValue)
	case reflect.Struct:
		return convertJSONStringToStruct(envValue, targetType)
	case reflect.Ptr:
		if targetType.Elem().Kind() == reflect.Struct {
			return convertJSONStringToStruct(envValue, targetType.Elem())
		}
		return nil, errors.New("unsupported pointer type conversion for type %v", targetType)
	default:
		return nil, errors.New("unsupported type conversion for type %v", targetType)
	}
}

// convertJSONStringToStruct attempts to unmarshal a JSON string into a struct
func convertJSONStringToStruct(jsonStr string, targetType reflect.Type) (interface{}, error) {
	structValue := reflect.New(targetType).Interface()
	err := gojson.Unmarshal([]byte(jsonStr), structValue)
	if err != nil {
		return nil, errors.New("failed to unmarshal JSON string to struct %v: %w", targetType, err)
	}

	return reflect.ValueOf(structValue).Elem().Interface(), nil
}

// setJSONPathValue sets a value at a specific path in a JSON map
func setJSONPathValue(jsonMap map[string]interface{}, path []string, value interface{}) error {
	if len(path) == 0 {
		return errors.New("empty path provided")
	}

	current := jsonMap
	for i, key := range path {
		if i == len(path)-1 {
			// Last key, set the value
			current[key] = value
			return nil
		}

		// Navigate to the next level
		if nextMap, ok := current[key].(map[string]interface{}); ok {
			current = nextMap
		} else {
			// Create intermediate objects if they don't exist
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}

	return nil
}

func replaceEnvVarsInContent(content []byte, log logger.Interface) []byte {
	result := envRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		// Use submatch to extract the variable name using the regex groups
		submatches := envRegex.FindSubmatch(match)
		if len(submatches) < 3 {
			log.With("match", string(match)).Warn("invalid environment variable pattern found")
			return match
		}

		// submatches[0] contains the placeholder ${#env.VAR} or ${env.VAR}
		// submatches[1] contains the optional "#" character
		// submatches[2] contains the environment variable name
		envVarName := types.UnsafeString(submatches[2])
		envVal := os.Getenv(envVarName)
		if len(envVal) == 0 {
			log.With("env_var", envVarName).With("placeholder", string(match)).Debug("environment variable not found, keeping placeholder")
			return match
		}

		log.With("env_var", envVarName).With("placeholder", string(match)).Debug("replacing environment variable")

		// Sanitize JSON if it's a JSON string
		sanitizedEnvVal := sanitizeJSONString(envVal)
		return fixQuotedJSONObjects(sanitizedEnvVal)
	})

	return result
}

// sanitizeJSONString sanitizes JSON strings from environment variables by removing unnecessary whitespace
func sanitizeJSONString(jsonStr string) string {
	// First, try to parse the JSON to validate it
	var parsed interface{}
	if err := gojson.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return jsonStr // Return original if not valid JSON
	}

	// Re-marshal to get compact JSON without extra whitespace
	if compacted, err := gojson.Marshal(parsed); err == nil {
		return string(compacted)
	}

	return jsonStr
}

func fixQuotedJSONObjects(content string) []byte {
	result := jsonPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Remove the outer quotes (first and last character)
		if len(match) < 2 {
			return match
		}

		inner := match[1 : len(match)-1]

		// Check if it's valid JSON
		if isValidJSON(inner) {
			// Parse and re-marshal to compact the JSON
			var parsed interface{}
			if err := gojson.Unmarshal([]byte(inner), &parsed); err == nil {
				if compacted, err := gojson.Marshal(parsed); err == nil {
					return string(compacted)
				}
			}
		}

		// If not valid JSON, return the original match
		return match
	})

	return types.UnsafeBytes(result)
}

func findMatchingBraceWithQuote(s string, start int) (int, bool) {
	if start >= len(s) || s[start] != '{' {
		return -1, false
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		char := s[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if char == '{' {
			depth++
		} else if char == '}' {
			depth--
			if depth == 0 {
				// Check if the next character is a quote (closing the JSON string)
				if i+1 < len(s) && s[i+1] == '"' {
					return i, true
				}
				return -1, false
			}
		}
	}

	return -1, false
}

func findMatchingBracketWithQuote(s string, start int) (int, bool) {
	if start >= len(s) || s[start] != '[' {
		return -1, false
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		char := s[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if char == '[' {
			depth++
		} else if char == ']' {
			depth--
			if depth == 0 {
				// Check if the next character is a quote (closing the JSON string)
				if i+1 < len(s) && s[i+1] == '"' {
					return i, true
				}
				return -1, false
			}
		}
	}

	return -1, false
}

func isValidJSON(s string) bool {
	var js interface{}
	return gojson.Unmarshal([]byte(s), &js) == nil
}

func replaceRefBlocks(validJSON string, log logger.Interface) (string, error) {
	matches := sharedBlocks.FindAllStringSubmatch(validJSON, -1)
	if len(matches) == 0 {
		log.Debug("no ref blocks found in configuration")
		return validJSON, nil
	}

	log.With("ref_blocks_count", len(matches)).Debug("found ref blocks in configuration")

	var rawData map[string]interface{}
	if err := gojson.Unmarshal(types.UnsafeBytes(validJSON), &rawData); err != nil {
		log.With("error", err).Error("failed to parse JSON for ref block resolution")
		return "", errors.New("failed to parse JSON for shared resolution due to invalid json", err)
	}

	replacements := make(map[string]string)
	for _, match := range matches {
		if len(match) < 3 {
			log.With("match", match).Warn("invalid ref block pattern found")
			continue
		}

		fullMatch := match[1]
		sharedPath := match[2]

		if _, exists := replacements[fullMatch]; exists {
			log.With("ref_block", fullMatch).Debug("ref block already processed, skipping")
			continue
		}

		log.With("ref_block", fullMatch).With("path", sharedPath).Debug("resolving ref block")

		referencedNode, err := nodeFromJson(rawData, sharedPath)
		if err != nil {
			log.With("ref_block", fullMatch).With("path", sharedPath).With("error", err).Error("failed to resolve ref block")
			return "", errors.New("failed to resolve reference %s: %w", fullMatch, err)
		}

		var buf bytes.Buffer
		encoder := gojson.NewEncoder(&buf)
		encoder.SetEscapeHTML(false)
		err = encoder.Encode(referencedNode)
		if err != nil {
			log.With("ref_block", fullMatch).With("error", err).Error("failed to marshal referenced node")
			return "", errors.New("failed to marshal referenced node %s: %w", fullMatch, err)
		}

		replacements[fullMatch] = strings.TrimSuffix(buf.String(), "\n")
		log.With("ref_block", fullMatch).Debug("successfully resolved ref block")
	}

	result := validJSON
	for placeholder, replacement := range replacements {
		result = strings.ReplaceAll(result, `"`+placeholder+`"`, replacement)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	log.With("replacements_count", len(replacements)).Debug("completed ref block replacements")
	return result, nil
}

func nodeFromJson(data map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil, errors.New("path component '%s' not found in path '%s'", part, path)
		}

		if i == len(parts)-1 {
			return value, nil
		}

		nextMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.New("path component '%s' is not an object, cannot navigate further in path '%s'", part, path)
		}

		current = nextMap
	}

	return current, nil
}

func fromTomlBytes(tomlBytes []byte, holder interface{}) error {
	err := toml.Unmarshal(tomlBytes, holder)
	if err != nil {
		return err
	}

	return nil
}

func fromJSONBytes(jsonBytes []byte, holder interface{}) error {
	err := gojson.Unmarshal(jsonBytes, holder)
	if err != nil {
		return err
	}

	return nil
}
