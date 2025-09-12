package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiKey(t *testing.T) {
	keys := map[string]string{
		"test.abc.pt":  "token1",
		"test.abc.com": "token2",
		"test.abc.es":  "token3",
		"test.abc.sh":  "token4",
	}

	assert.Equal(t, "token1", getHostAPIKey("http://test.abc.pt/login", keys))
	assert.Equal(t, "token2", getHostAPIKey("http://test.abc.com/login", keys))
	assert.Equal(t, "token3", getHostAPIKey("http://test.abc.es/login", keys))
	assert.Equal(t, "token4", getHostAPIKey("http://test.abc.sh/login", keys))
	assert.Equal(t, "", getHostAPIKey("http://a.test.abc.pt/login", keys))
	assert.Equal(t, "", getHostAPIKey("http://atest.abc.pt/login", keys))
}
