package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

// Calculator provides utilities for calculating and comparing SHA256 and SHA512 hashes
type Calculator struct {
	// Could potentially cache previous hashes for more efficient comparison
	cache map[string]string // optional cache for future optimization
}

// NewCalculator creates a new instance of Calculator
func NewCalculator() *Calculator {
	return &Calculator{
		cache: make(map[string]string),
	}
}

// CalculateSHA256 calculates a deterministic SHA256 hash from any data structure
// The data is normalized by sorting keys to ensure consistent hash generation
func (c *Calculator) CalculateSHA256(data interface{}) (string, error) {
	normalizedData, err := c.normalizeData(data)
	if err != nil {
		return "", err
	}

	return c.calculateSHA256FromMap(normalizedData)
}

// CalculateSHA256FromMap calculates SHA256 from a map with sorted keys for deterministic results
func (c *Calculator) CalculateSHA256FromMap(data map[string]interface{}) (string, error) {
	return c.calculateSHA256FromMap(data)
}

// CalculateSHA512 calculates a deterministic SHA512 hash from any data structure
// The data is normalized by sorting keys to ensure consistent hash generation
func (c *Calculator) CalculateSHA512(data interface{}) (string, error) {
	normalizedData, err := c.normalizeData(data)
	if err != nil {
		return "", err
	}

	return c.calculateSHA512FromMap(normalizedData)
}

// CalculateSHA512FromMap calculates SHA512 from a map with sorted keys for deterministic results
func (c *Calculator) CalculateSHA512FromMap(data map[string]interface{}) (string, error) {
	return c.calculateSHA512FromMap(data)
}

// CompareSHA compares two SHA256 hashes and returns true if they are different
// For more efficient comparison, this uses string comparison which is fast in Go
func (c *Calculator) CompareSHA(currentSHA, newSHA string) bool {
	// Using string comparison is already efficient in Go
	// For future optimization, we could:
	// 1. Use bytes.Equal if we store as byte slices
	// 2. Implement short-circuit comparison for very long hashes
	// 3. Cache previous comparisons to avoid repeat work
	return currentSHA != newSHA
}

// CompareSHA512 compares two SHA512 hashes and returns true if they are different
func (c *Calculator) CompareSHA512(currentSHA, newSHA string) bool {
	return currentSHA != newSHA
}

// ValidateSHA256 checks if a SHA256 string is valid (64 character hex string)
func (c *Calculator) ValidateSHA256(sha string) bool {
	if len(sha) != 64 {
		return false
	}

	// Check if it's a valid hex string
	_, err := hex.DecodeString(sha)
	return err == nil
}

// ValidateSHA512 checks if a SHA512 string is valid (128 character hex string)
func (c *Calculator) ValidateSHA512(sha string) bool {
	if len(sha) != 128 {
		return false
	}

	// Check if it's a valid hex string
	_, err := hex.DecodeString(sha)
	return err == nil
}

// normalizeData converts any data structure to a normalized map for consistent hashing
func (c *Calculator) normalizeData(data interface{}) (map[string]interface{}, error) {
	// First marshal to JSON to normalize the data structure
	jsonData, err := serializer.Serialize(data)
	if err != nil {
		return nil, err
	}

	// Then unmarshal back to a map for key sorting
	var normalizedData map[string]interface{}
	err = serializer.Deserialize(jsonData, &normalizedData, false)
	if err != nil {
		return nil, err
	}

	return normalizedData, nil
}

// calculateSHA256FromMap creates a deterministic SHA256 hash from a map of data
func (c *Calculator) calculateSHA256FromMap(data map[string]interface{}) (string, error) {
	// Sort keys to ensure deterministic hash generation
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create ordered map for consistent JSON marshaling
	orderedData := make(map[string]interface{})
	for _, k := range keys {
		orderedData[k] = data[k]
	}

	jsonData, err := serializer.Serialize(orderedData)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// calculateSHA512FromMap creates a deterministic SHA512 hash from a map of data
func (c *Calculator) calculateSHA512FromMap(data map[string]interface{}) (string, error) {
	// Sort keys to ensure deterministic hash generation
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create ordered map for consistent JSON marshaling
	orderedData := make(map[string]interface{})
	for _, k := range keys {
		orderedData[k] = data[k]
	}

	jsonData, err := json.Marshal(orderedData)
	if err != nil {
		return "", err
	}

	hash := sha512.Sum512(jsonData)
	return hex.EncodeToString(hash[:]), nil
}
