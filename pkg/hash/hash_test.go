package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculator_CalculateSHA256(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should calculate consistent SHA for same data", func(t *testing.T) {
		data := map[string]interface{}{
			"id":          "123",
			"name":        "Test",
			"description": "Test description",
			"count":       42,
			"active":      true,
		}

		sha1, err := calculator.CalculateSHA256(data)
		require.NoError(t, err)
		assert.Len(t, sha1, 64) // SHA256 produces 64 character hex string

		sha2, err := calculator.CalculateSHA256(data)
		require.NoError(t, err)

		assert.Equal(t, sha1, sha2, "SHA should be consistent for same data")
	})

	t.Run("should produce different SHA for different data", func(t *testing.T) {
		data1 := map[string]interface{}{
			"id":   "123",
			"name": "Test",
		}

		data2 := map[string]interface{}{
			"id":   "123",
			"name": "Different Test", // Different value
		}

		sha1, err := calculator.CalculateSHA256(data1)
		require.NoError(t, err)

		sha2, err := calculator.CalculateSHA256(data2)
		require.NoError(t, err)

		assert.NotEqual(t, sha1, sha2, "SHA should be different for different data")
	})

	t.Run("should be deterministic regardless of field order", func(t *testing.T) {
		// This test ensures that the SHA is calculated consistently
		// even if the map iteration order changes
		data := map[string]interface{}{
			"z_field": "last",
			"a_field": "first",
			"m_field": "middle",
		}

		// Calculate SHA multiple times
		var shas []string
		for i := 0; i < 5; i++ {
			sha, err := calculator.CalculateSHA256(data)
			require.NoError(t, err)
			shas = append(shas, sha)
		}

		// All SHAs should be identical
		for i := 1; i < len(shas); i++ {
			assert.Equal(t, shas[0], shas[i], "SHA should be deterministic")
		}
	})

	t.Run("should handle nested structures", func(t *testing.T) {
		data := map[string]interface{}{
			"id": "123",
			"metadata": map[string]interface{}{
				"created_at": "2024-01-01",
				"tags":       []string{"tag1", "tag2"},
			},
		}

		sha, err := calculator.CalculateSHA256(data)
		require.NoError(t, err)
		assert.Len(t, sha, 64)
	})

	t.Run("should handle struct input", func(t *testing.T) {
		type TestStruct struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		data := TestStruct{
			ID:          "123",
			Name:        "Test",
			Description: "Test description",
		}

		sha, err := calculator.CalculateSHA256(data)
		require.NoError(t, err)
		assert.Len(t, sha, 64)
	})
}

func TestCalculator_CalculateSHA256FromMap(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should calculate SHA from map directly", func(t *testing.T) {
		data := map[string]interface{}{
			"id":   "123",
			"name": "Test",
		}

		sha, err := calculator.CalculateSHA256FromMap(data)
		require.NoError(t, err)
		assert.Len(t, sha, 64)
	})

	t.Run("should produce same result as CalculateSHA256 for map input", func(t *testing.T) {
		data := map[string]interface{}{
			"id":   "123",
			"name": "Test",
		}

		sha1, err := calculator.CalculateSHA256(data)
		require.NoError(t, err)

		sha2, err := calculator.CalculateSHA256FromMap(data)
		require.NoError(t, err)

		assert.Equal(t, sha1, sha2, "Both methods should produce same result for map input")
	})
}

func TestCalculator_CompareSHA(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should return false for identical SHAs", func(t *testing.T) {
		sha := "abc123def456"
		result := calculator.CompareSHA(sha, sha)
		assert.False(t, result, "Identical SHAs should return false (no change)")
	})

	t.Run("should return true for different SHAs", func(t *testing.T) {
		sha1 := "abc123def456"
		sha2 := "xyz789uvw012"
		result := calculator.CompareSHA(sha1, sha2)
		assert.True(t, result, "Different SHAs should return true (change detected)")
	})

	t.Run("should handle empty SHAs", func(t *testing.T) {
		result1 := calculator.CompareSHA("", "abc123")
		assert.True(t, result1, "Empty vs non-empty should return true")

		result2 := calculator.CompareSHA("abc123", "")
		assert.True(t, result2, "Non-empty vs empty should return true")

		result3 := calculator.CompareSHA("", "")
		assert.False(t, result3, "Empty vs empty should return false")
	})
}

func TestCalculator_ValidateSHA256(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should validate correct SHA format", func(t *testing.T) {
		validSHA := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
		result := calculator.ValidateSHA256(validSHA)
		assert.True(t, result, "Valid 64-character hex string should be valid")
	})

	t.Run("should reject invalid SHA formats", func(t *testing.T) {
		testCases := []struct {
			name string
			sha  string
		}{
			{"too short", "abc123"},
			{"too long", "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef1234567890"},
			{"non-hex characters", "g1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"},
			{"empty string", ""},
			{"special characters", "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef12345!"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := calculator.ValidateSHA256(tc.sha)
				assert.False(t, result, "Invalid SHA should be rejected: %s", tc.sha)
			})
		}
	})
}

func TestCalculator_CalculateSHA512(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should calculate consistent SHA512 for same data", func(t *testing.T) {
		data := map[string]interface{}{
			"id":          "123",
			"name":        "Test",
			"description": "Test description",
			"count":       42,
		}

		sha1, err := calculator.CalculateSHA512(data)
		require.NoError(t, err)
		assert.Len(t, sha1, 128) // SHA512 produces 128 character hex string

		sha2, err := calculator.CalculateSHA512(data)
		require.NoError(t, err)

		assert.Equal(t, sha1, sha2, "SHA512 should be consistent for same data")
	})

	t.Run("should produce different SHA512 for different data", func(t *testing.T) {
		data1 := map[string]interface{}{
			"id":   "123",
			"name": "Test",
		}

		data2 := map[string]interface{}{
			"id":   "123",
			"name": "Different Test",
		}

		sha1, err := calculator.CalculateSHA512(data1)
		require.NoError(t, err)

		sha2, err := calculator.CalculateSHA512(data2)
		require.NoError(t, err)

		assert.NotEqual(t, sha1, sha2, "SHA512 should be different for different data")
	})
}

func TestCalculator_CompareSHA512(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should return false for identical SHA512s", func(t *testing.T) {
		sha := "abc123def456"
		result := calculator.CompareSHA512(sha, sha)
		assert.False(t, result, "Identical SHA512s should return false")
	})

	t.Run("should return true for different SHA512s", func(t *testing.T) {
		sha1 := "abc123def456"
		sha2 := "xyz789uvw012"
		result := calculator.CompareSHA512(sha1, sha2)
		assert.True(t, result, "Different SHA512s should return true")
	})
}

func TestCalculator_ValidateSHA512(t *testing.T) {
	calculator := NewCalculator()

	t.Run("should validate correct SHA512 format", func(t *testing.T) {
		validSHA := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
		result := calculator.ValidateSHA512(validSHA)
		assert.True(t, result, "Valid 128-character hex string should be valid")
	})

	t.Run("should reject invalid SHA512 formats", func(t *testing.T) {
		testCases := []struct {
			name string
			sha  string
		}{
			{"too short", "abc123"},
			{"too long", "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef1234567890a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef12345678"},
			{"non-hex characters", "g1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456g1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"},
			{"empty string", ""},
			{"SHA256 length instead of SHA512", "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := calculator.ValidateSHA512(tc.sha)
				assert.False(t, result, "Invalid SHA512 should be rejected: %s", tc.sha)
			})
		}
	})
}

// Test the original functions for backward compatibility
func TestComputeSHA256(t *testing.T) {
	data := map[string]interface{}{
		"id":   "123",
		"name": "Test",
	}

	sha, err := ComputeSHA256(data)
	require.NoError(t, err)
	assert.Len(t, sha, 64)
}

func TestComputeSHA512(t *testing.T) {
	data := map[string]interface{}{
		"id":   "123",
		"name": "Test",
	}

	sha, err := ComputeSHA512(data)
	require.NoError(t, err)
	assert.Len(t, sha, 128) // SHA512 produces 128 character hex string
}
