package uidgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemUUID(t *testing.T) {
	// Test that the UUIDs match the required format
	expectedSystemUUID := "00000000-1111-0000-0000-000000000000"
	assert.Equal(t, expectedSystemUUID, SystemUUID, "SystemUUID should match the expected format")
}

func TestHealthCheckerUUID(t *testing.T) {
	expectedHealthCheckerUUID := "00000000-1111-0001-0000-000000000000"
	assert.Equal(t, expectedHealthCheckerUUID, HealthCheckerUUID, "HealthCheckerUUID should match the expected format")
}

func TestTestUUID(t *testing.T) {
	expectedTestUUID := "00000000-2222-0000-0000-000000000000"
	assert.Equal(t, expectedTestUUID, BotUUID, "TestUUID should match the expected format")
}

func TestBroadcastAllUUID(t *testing.T) {
	expectedBroadcastAllUUID := "00000000-0000-0000-0000-000000000000"
	assert.Equal(t, expectedBroadcastAllUUID, BroadcastAllUUID, "BroadcastAllUUID should match the expected format")
}

func TestULIDToUUID_Conversion(t *testing.T) {
	// Test that the ULID to UUID conversion is correct
	assert.Equal(t, SystemUID.UUID(), SystemUUID, "SystemUID should convert to SystemUUID")
	assert.Equal(t, HealthCheckerUID.UUID(), HealthCheckerUUID, "HealthCheckerUID should convert to HealthCheckerUUID")
	assert.Equal(t, BotUID.UUID(), BotUUID, "BotUID should convert to BotUUID")
	assert.Equal(t, BroadcastAllUID.UUID(), BroadcastAllUUID, "BroadcastAllUID should convert to BroadcastAllUUID")
}

// Additional test to verify that the binary representations actually match the expected values
func TestULIDStringRepresentations(t *testing.T) {
	// This test ensures that the byte sequences used to create the ULIDs
	// actually produce UUIDs matching our expected format
	assert.Equal(t, "000000048H0000000000000000", SystemUID.String())
	assert.Equal(t, "000000048H000G000000000000", HealthCheckerUID.String())
	assert.Equal(t, "00000008H20000000000000000", BotUID.String())
	assert.Equal(t, "00000000000000000000000000", BroadcastAllUID.String())
}
