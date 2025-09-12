package uid

import (
	"fmt"
	"testing"

	pulid "github.com/pixie-sh/ulid-go"
	"github.com/stretchr/testify/assert"
)

func TestUUIDv4Parsing(t *testing.T) {
	var id UID

	err := id.UnmarshalText([]byte("998c402f-eec4-4d91-8e51-0751b7f4d8bf"))
	assert.NoError(t, err)
	fmt.Println(id.String())
	fmt.Println(id.UUID())
	assert.Equal(t, "998c402f-eec4-4d91-8e51-0751b7f4d8bf", id.UUID())
}

func TestScan(t *testing.T) {
	// Testing struct
	var id pulid.ULID

	t.Run("Scan valid UID binary (16 bytes)", func(t *testing.T) {
		binaryULID, _ := pulid.MustNew().MarshalBinary()
		err := id.Scan(binaryULID)
		assert.NoError(t, err, "expected no error when scanning valid 16-byte binary UID")
	})

	t.Run("Scan valid UID string (36 characters)", func(t *testing.T) {
		uuidString := pulid.MustNew().UUID()
		err := id.Scan(uuidString)
		assert.NoError(t, err, "expected no error when scanning valid UID string")
	})

	t.Run("Scan invalid binary size", func(t *testing.T) {
		invalidBinary := []byte{1, 2, 3} // A 3-byte binary
		err := id.Scan(invalidBinary)
		assert.Error(t, err, "expected an error when scanning invalid binary size")
		assert.Contains(t, err.Error(), "invalid storage format", "should return 'invalid storage format' error")
	})

	t.Run("Scan invalid string size", func(t *testing.T) {
		invalidString := "1234"
		err := id.Scan(invalidString)
		assert.Error(t, err, "expected an error when scanning invalid string size")
		assert.Contains(t, err.Error(), "invalid storage format", "should return 'invalid storage format' error")
	})

	t.Run("Scan invalid type", func(t *testing.T) {
		invalidType := 12345
		err := id.Scan(invalidType)
		assert.Error(t, err, "expected an error when scanning invalid type")
		assert.Contains(t, err.Error(), "invalid storage format", "should return 'invalid storage format' error")
	})
}

// Benchmark for Scan function
func BenchmarkScan(b *testing.B) {
	// Testing struct
	var id pulid.ULID

	// Generate UID binary and string for tests
	validBinaryULID, _ := pulid.MustNew().MarshalBinary()
	//ulidstring := pulid.MustNew().String()
	uuidString := pulid.MustNew().UUID()

	b.Run("Gen UID v4 compatibles", func(b *testing.B) {
		mapp := make(map[string]bool)
		for i := 0; i < b.N; i++ {
			uuid := pulid.MustNew().UUID()
			if _, ok := mapp[uuid]; ok {
				panic("duplicate uuid")
			}
			mapp[uuid] = true

			uuid1 := pulid.MustNew().UUID()
			if _, ok := mapp[uuid1]; ok {
				panic("duplicate uuid")
			}
			mapp[uuid1] = true

			uuid2 := pulid.MustNew().UUID()
			if _, ok := mapp[uuid2]; ok {
				panic("duplicate uuid")
			}
			mapp[uuid2] = true

		}

	})

	b.Run("Scan valid UID binary (16 bytes)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := id.Scan(validBinaryULID); err != nil {
				b.Fatalf("unexpected error during benchmark: %v", err)
			}
		}
	})

	b.Run("Scan valid UID string (36 characters)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuid2 := pulid.MustNew().UUID()
			if err := id.Scan(uuidString); err != nil {
				b.Fatalf("unexpected error during benchmark: %v", err)
			}

			if err := id.Scan(uuid2); err != nil {
				b.Fatalf("unexpected error during benchmark: %v", err)
			}
		}
	})

	b.Run("Scan invalid binary size", func(b *testing.B) {
		invalidBinary := []byte{1, 2, 3} // A 3-byte binary
		for i := 0; i < b.N; i++ {
			_ = id.Scan(invalidBinary)
		}
	})

	b.Run("Scan invalid string size", func(b *testing.B) {
		invalidString := "1234" // A short invalid string
		for i := 0; i < b.N; i++ {
			_ = id.Scan(invalidString)
		}
	})
}
