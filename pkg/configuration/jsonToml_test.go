package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BurntSushi/toml"
	goj "github.com/goccy/go-json"
	"github.com/pixie-sh/logger-go/logger"
)

type Config struct {
	Name    string `json:"name" `
	Enabled bool   `json:"enabled" `
	Count   int    `json:"count" `
}

func getJson() string {
	return fmt.Sprintf(`{"name":"test","enabled":true,"count":%d}`, time.Now().UnixNano())
}

func TestJSONvsTOML(t *testing.T) {
	jsonData := `{"name":"test","enabled":true,"count":3}`
	tomlData := `
name = "test"
enabled = true
count = 3
`
	var config Config

	iter := 100000

	// Measure JSON deserialization
	startJson := time.Now()
	for i := 0; i < 1*iter; i++ {
		if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
			fmt.Println("JSON unmarshal error:", err)
			return
		}
	}
	durationJson := time.Since(startJson)
	fmt.Println("JSON Deserialization took:", durationJson)

	//// Measure JSON deserialization
	//startJson = time.Now()
	//for i := 0; i < 1*iter; i++ {
	//	err := oj.Unmarshal([]byte(getJson()), &config)
	//	if err != nil {
	//		fmt.Println("JSON unmarshal error:", err)
	//		return
	//	}
	//}
	//durationJson = time.Since(startJson)
	//fmt.Println("OJG Deserialization took:", durationJson)

	// Measure JSON deserialization
	startJson = time.Now()
	for i := 0; i < 1*iter; i++ {
		err := goj.Unmarshal([]byte(jsonData), &config)
		if err != nil {
			fmt.Println("JSON unmarshal error:", err)
			return
		}
	}
	durationJson = time.Since(startJson)
	fmt.Println("GO JSON Deserialization took:", durationJson)

	// Measure TOML deserialization
	startToml := time.Now()
	for i := 0; i < 1*iter; i++ {
		if _, err := toml.Decode(tomlData, &config); err != nil {
			fmt.Println("TOML unmarshal error:", err)
			return
		}
	}
	durationToml := time.Since(startToml)
	fmt.Println("TOML Deserialization took:", durationToml)
}

func TestJSONWithSharedBlocks(t *testing.T) {
	os.Setenv("REDIS_ADDRESS", "localhost:6379")
	os.Setenv("REDIS_DB", "0")

	expectedJson := `
	{
  "#ref": {
    "http_port": 8080,
    "metrics_port": 3000,
    "redis_config": {
      "address": "localhost:6379",
      "password": "",
      "db": "0"
    }
  },
  "listen_addr": 8080,
  "listen_metrics_addr": 3000,
  "token_services_bundle": {
    "token_service": {
      "validity_in_seconds_token": 5259492,
      "token_private_key": "aaaa",
      "token_public_key": "aaaa=="
    },
    "token_cache": {
      "address": "localhost:6379",
      "db": "0",
      "password": ""
    }
  }
}
	`

	jsonData := `
	{
  "#ref": {
    "http_port": 8080,
    "metrics_port": 3000,
    "redis_config": {
      "address": "${env.REDIS_ADDRESS}",
      "password": "",
      "db": "${env.REDIS_DB}"
    }
  },
  "listen_addr": "${#ref.http_port}",
  "listen_metrics_addr": "${#ref.metrics_port}",
  "token_services_bundle": {
    "token_service": {
      "validity_in_seconds_token": 5259492,
      "token_private_key": "aaaa",
      "token_public_key": "aaaa=="
    },
    "token_cache": "${#ref.redis_config}"
  }
}`

	var holder any
	res, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	var expectedHolder any
	err = json.Unmarshal([]byte(expectedJson), &expectedHolder)
	assert.Nil(t, err)

	assert.Equal(t, expectedHolder, holder)
}

func TestEnvPriorityOverride(t *testing.T) {
	// Set environment variables that should override JSON values
	os.Setenv("listen_addr", "9090")
	os.Setenv("listen_metrics_addr", "4000")
	os.Setenv("token_services_bundle.token_service.validity_in_seconds_token", "7200")

	defer func() {
		os.Unsetenv("listen_addr")
		os.Unsetenv("listen_metrics_addr")
		os.Unsetenv("token_services_bundle.token_service.validity_in_seconds_token")
	}()

	jsonData := `
	{
  "#ref": {
    "http_port": 8080,
    "metrics_port": 3000,
    "redis_config": {
      "address": "localhost:6379",
      "password": "",
      "db": "0"
    }
  },
  "listen_addr": "${#ref.http_port}",
  "listen_metrics_addr": "${#ref.metrics_port}",
  "token_services_bundle": {
    "token_service": {
      "validity_in_seconds_token": 5259492,
      "token_private_key": "aaaa",
      "token_public_key": "aaaa=="
    },
    "token_cache": "${#ref.redis_config}"
  }
}`

	var holder testCfg
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.Nil(t, err)

	// Verify environment variables took priority with correct types
	assert.Equal(t, "9090", holder.ListenAddr)
	assert.Equal(t, "4000", holder.ListenMetricsAddr)
	assert.Equal(t, 7200, holder.TokenServicesBundle.TokenService.ValidityInSecondsToken) // Should be int, not string

	// Verify non-overridden values remain from JSON/ref resolution
	assert.Equal(t, "aaaa", holder.TokenServicesBundle.TokenService.TokenPrivateKey)
	assert.Equal(t, "aaaa==", holder.TokenServicesBundle.TokenService.TokenPublicKey)
	assert.Equal(t, "localhost:6379", holder.TokenServicesBundle.TokenCache.Address)
	assert.Equal(t, "0", holder.TokenServicesBundle.TokenCache.DB)
}

func TestEnvPriorityOverrideWithDifferentTypes(t *testing.T) {
	// Test with different data types
	os.Setenv("test_string", "hello")
	os.Setenv("test_int", "42")
	os.Setenv("test_bool", "true")
	os.Setenv("test_float", "3.14")

	defer func() {
		os.Unsetenv("test_string")
		os.Unsetenv("test_int")
		os.Unsetenv("test_bool")
		os.Unsetenv("test_float")
	}()

	type testTypes struct {
		TestString string  `json:"test_string"`
		TestInt    int     `json:"test_int"`
		TestBool   bool    `json:"test_bool"`
		TestFloat  float64 `json:"test_float"`
	}

	jsonData := `{
		"test_string": "default",
		"test_int": 0,
		"test_bool": false,
		"test_float": 0.0
	}`

	var holder testTypes
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.Nil(t, err)

	// Verify all types are correctly converted
	assert.Equal(t, "hello", holder.TestString)
	assert.Equal(t, 42, holder.TestInt)
	assert.Equal(t, true, holder.TestBool)
	assert.Equal(t, 3.14, holder.TestFloat)
}

func TestJSONEnvPlaceholderInjectsRawJSONNode(t *testing.T) {
	t.Setenv("EMAIL_CONFIGURATION_PAYLOAD", `{"host":"mailpit","port":1025,"from_sender":"Test Sender <noreply@example.test>","auth_mode":"none","tls_mode":"none","tls_skip_verify":false}`)

	type emailServiceConfig struct {
		Driver        string         `json:"driver"`
		Configuration map[string]any `json:"configuration"`
	}
	type configWithEmail struct {
		EmailService emailServiceConfig `json:"email_service"`
	}

	jsonData := `{
		"email_service": {
			"driver": "smtp",
			"configuration": "${env.json.EMAIL_CONFIGURATION_PAYLOAD}"
		}
	}`

	var holder configWithEmail
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.Nil(t, err)
	assert.Equal(t, "smtp", holder.EmailService.Driver)
	assert.Equal(t, "mailpit", holder.EmailService.Configuration["host"])
	assert.Equal(t, float64(1025), holder.EmailService.Configuration["port"])
	assert.Equal(t, false, holder.EmailService.Configuration["tls_skip_verify"])
}

func TestJSONEnvPlaceholderRequiresValidJSON(t *testing.T) {
	t.Setenv("EMAIL_CONFIGURATION_PAYLOAD", `not-json`)

	jsonData := `{"configuration":"${env.json.EMAIL_CONFIGURATION_PAYLOAD}"}`

	var holder map[string]any
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.NotNil(t, err)
}

func TestJSONEnvPlaceholderMustBeFullStringValue(t *testing.T) {
	t.Setenv("EMAIL_CONFIGURATION_PAYLOAD", `{}`)

	jsonData := `{"configuration":"prefix-${env.json.EMAIL_CONFIGURATION_PAYLOAD}"}`

	var holder map[string]any
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.NotNil(t, err)
}

func TestReplaceRefBlocks_Jsonfiles_HappyPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "email_service.json")
	fileContent := `{"driver":"smtp","configuration":{"host":"mailpit","port":1025,"from_sender":"Test Sender <noreply@example.test>","auth_mode":"none","tls_mode":"none","tls_skip_verify":false}}`
	require.NoError(t, os.WriteFile(filePath, []byte(fileContent), 0o644))
	t.Setenv("TEST_PROVIDER_HAPPY", filePath)

	type emailServiceConfig struct {
		Driver        string         `json:"driver"`
		Configuration map[string]any `json:"configuration"`
	}
	type configWithEmail struct {
		EmailService emailServiceConfig `json:"email_service"`
	}

	jsonData := `{"email_service": "${#ref.jsonfiles.TEST_PROVIDER_HAPPY}"}`

	var holder configWithEmail
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	require.NoError(t, err)
	assert.Equal(t, "smtp", holder.EmailService.Driver)
	assert.Equal(t, "mailpit", holder.EmailService.Configuration["host"])
	assert.Equal(t, float64(1025), holder.EmailService.Configuration["port"])
	assert.Equal(t, false, holder.EmailService.Configuration["tls_skip_verify"])
}

func TestReplaceRefBlocks_Jsonfiles_EnvUnset(t *testing.T) {
	require.NoError(t, os.Unsetenv("TEST_PROVIDER_UNSET"))

	jsonData := `{"email_service": "${#ref.jsonfiles.TEST_PROVIDER_UNSET}"}`

	var holder map[string]any
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_PROVIDER_UNSET")
}

func TestReplaceRefBlocks_Jsonfiles_FileMissing(t *testing.T) {
	t.Setenv("TEST_PROVIDER_MISSING", "/definitely/does/not/exist/email_service.json")

	jsonData := `{"email_service": "${#ref.jsonfiles.TEST_PROVIDER_MISSING}"}`

	var holder map[string]any
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_PROVIDER_MISSING")
}

func TestReplaceRefBlocks_Jsonfiles_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "invalid.json")
	require.NoError(t, os.WriteFile(filePath, []byte("not-valid-json-at-all"), 0o644))
	t.Setenv("TEST_PROVIDER_INVALID", filePath)

	jsonData := `{"email_service": "${#ref.jsonfiles.TEST_PROVIDER_INVALID}"}`

	var holder map[string]any
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_PROVIDER_INVALID")
}

func TestReplaceRefBlocks_Jsonfiles_MixedWithInConfigRef(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "provider.json")
	fileContent := `{"driver":"smtp","configuration":{"host":"smtp.example.com","port":587}}`
	require.NoError(t, os.WriteFile(filePath, []byte(fileContent), 0o644))
	t.Setenv("TEST_PROVIDER_MIXED", filePath)

	jsonData := `{
		"#ref": {
			"redis_config": {
				"address": "localhost:6379",
				"password": "",
				"db": "0"
			}
		},
		"email_service": "${#ref.jsonfiles.TEST_PROVIDER_MIXED}",
		"token_cache": "${#ref.redis_config}"
	}`

	type emailServiceConfig struct {
		Driver        string         `json:"driver"`
		Configuration map[string]any `json:"configuration"`
	}
	type mixedConfig struct {
		EmailService emailServiceConfig `json:"email_service"`
		TokenCache   tokenCache         `json:"token_cache"`
	}

	var holder mixedConfig
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	require.NoError(t, err)
	assert.Equal(t, "smtp", holder.EmailService.Driver)
	assert.Equal(t, "smtp.example.com", holder.EmailService.Configuration["host"])
	assert.Equal(t, float64(587), holder.EmailService.Configuration["port"])
	assert.Equal(t, "localhost:6379", holder.TokenCache.Address)
	assert.Equal(t, "0", holder.TokenCache.DB)
}

type testCfg struct {
	ListenAddr          string              `json:"listen_addr"`
	ListenMetricsAddr   string              `json:"listen_metrics_addr"`
	TokenServicesBundle tokenServicesBundle `json:"token_services_bundle"`
}

type tokenServicesBundle struct {
	TokenService tokenService `json:"token_service"`
	TokenCache   tokenCache   `json:"token_cache"`
}

type tokenService struct {
	ValidityInSecondsToken int    `json:"validity_in_seconds_token"`
	TokenPrivateKey        string `json:"token_private_key"`
	TokenPublicKey         string `json:"token_public_key"`
}

type tokenCache struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       string `json:"db"`
}

func TestEnvReplacementWithStructInJSON(t *testing.T) {
	// Set environment variable with JSON struct
	os.Setenv("redis_config", `{"address":"env-redis:6379","password":"secret","db":"1"}`)

	defer func() {
		os.Unsetenv("redis_config")
	}()

	jsonData := `
	{
	  "listen_addr": "8080",
	  "listen_metrics_addr": "3000",
	  "token_services_bundle": {
		"token_service": {
		  "validity_in_seconds_token": 5259492,
		  "token_private_key": "aaaa",
		  "token_public_key": "aaaa=="
		},
		"token_cache": "${#env.redis_config}"
  	}
}`

	var holder testCfg
	_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
	assert.Nil(t, err)

	// Verify the struct from environment variable was properly replaced and unmarshaled
	assert.Equal(t, "env-redis:6379", holder.TokenServicesBundle.TokenCache.Address)
	assert.Equal(t, "secret", holder.TokenServicesBundle.TokenCache.Password)
	assert.Equal(t, "1", holder.TokenServicesBundle.TokenCache.DB)

	// Verify other values remain unchanged
	assert.Equal(t, "8080", holder.ListenAddr)
	assert.Equal(t, "3000", holder.ListenMetricsAddr)
}

func TestEnvReplacementWithStructInJSONExtensive(t *testing.T) {
	// Test Case 1: Basic JSON object replacement
	t.Run("BasicJSONObjectReplacement", func(t *testing.T) {
		os.Setenv("redis_config", `{"address":"env-redis:6379","password":"secret","db":"1"}`)
		defer os.Unsetenv("redis_config")

		jsonData := `
		{
		  "listen_addr": "8080",
		  "listen_metrics_addr": "3000",
		  "token_services_bundle": {
			"token_service": {
			  "validity_in_seconds_token": 5259492,
			  "token_private_key": "aaaa",
			  "token_public_key": "aaaa=="
			},
			"token_cache": "${env.redis_config}"
		  }
		}`

		var holder testCfg
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		// Verify the struct from environment variable was properly replaced and unmarshaled
		assert.Equal(t, "env-redis:6379", holder.TokenServicesBundle.TokenCache.Address)
		assert.Equal(t, "secret", holder.TokenServicesBundle.TokenCache.Password)
		assert.Equal(t, "1", holder.TokenServicesBundle.TokenCache.DB)

		// Verify other values remain unchanged
		assert.Equal(t, "8080", holder.ListenAddr)
		assert.Equal(t, "3000", holder.ListenMetricsAddr)
	})

	// Test Case 2: JSON object with # prefix
	t.Run("JSONObjectWithHashPrefix", func(t *testing.T) {
		os.Setenv("redis_config", `{"address":"hash-redis:6379","password":"hash-secret","db":"2"}`)
		defer os.Unsetenv("redis_config")

		jsonData := `
		{
		  "listen_addr": "8080",
		  "listen_metrics_addr": "3000",
		  "token_services_bundle": {
			"token_service": {
			  "validity_in_seconds_token": 5259492,
			  "token_private_key": "aaaa",
			  "token_public_key": "aaaa=="
			},
			"token_cache": "${#env.redis_config}"
		  }
		}`

		var holder testCfg
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		assert.Equal(t, "hash-redis:6379", holder.TokenServicesBundle.TokenCache.Address)
		assert.Equal(t, "hash-secret", holder.TokenServicesBundle.TokenCache.Password)
		assert.Equal(t, "2", holder.TokenServicesBundle.TokenCache.DB)
	})

	// Test Case 3: JSON array replacement
	t.Run("JSONArrayReplacement", func(t *testing.T) {
		os.Setenv("server_list", `["server1:8080","server2:8081","server3:8082"]`)
		defer os.Unsetenv("server_list")

		type configWithArray struct {
			Servers []string `json:"servers"`
		}

		jsonData := `{"servers": "${env.server_list}"}`

		var holder configWithArray
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		assert.Equal(t, []string{"server1:8080", "server2:8081", "server3:8082"}, holder.Servers)
	})

	// Test Case 4: Complex nested JSON object
	t.Run("ComplexNestedJSONObject", func(t *testing.T) {
		complexJSON := `{
			"database": {
				"host": "db.example.com",
				"port": 5432,
				"credentials": {
					"username": "admin",
					"password": "super-secret"
				},
				"pools": {
					"read": 10,
					"write": 5
				}
			},
			"enabled": true
		}`
		os.Setenv("db_config", complexJSON)
		defer os.Unsetenv("db_config")

		type dbCredentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		type dbPools struct {
			Read  int `json:"read"`
			Write int `json:"write"`
		}

		type database struct {
			Host        string        `json:"host"`
			Port        int           `json:"port"`
			Credentials dbCredentials `json:"credentials"`
			Pools       dbPools       `json:"pools"`
		}

		type complexConfig struct {
			Database database `json:"database"`
			Enabled  bool     `json:"enabled"`
		}

		type appConfig struct {
			DBConfig complexConfig `json:"db_config"`
		}

		jsonData := `{"db_config": "${env.db_config}"}`

		var holder appConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		assert.Equal(t, "db.example.com", holder.DBConfig.Database.Host)
		assert.Equal(t, 5432, holder.DBConfig.Database.Port)
		assert.Equal(t, "admin", holder.DBConfig.Database.Credentials.Username)
		assert.Equal(t, "super-secret", holder.DBConfig.Database.Credentials.Password)
		assert.Equal(t, 10, holder.DBConfig.Database.Pools.Read)
		assert.Equal(t, 5, holder.DBConfig.Database.Pools.Write)
		assert.Equal(t, true, holder.DBConfig.Enabled)
	})

	// Test Case 5: Multiple JSON objects in same config
	t.Run("MultipleJSONObjectsReplacement", func(t *testing.T) {
		os.Setenv("redis_config", `{"address":"redis:6379","password":"redis-pass","db":"0"}`)
		os.Setenv("mongo_config", `{"host":"mongo.example.com","port":27017,"database":"myapp"}`)
		defer func() {
			os.Unsetenv("redis_config")
			os.Unsetenv("mongo_config")
		}()

		type mongoConfig struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		}

		type multiConfig struct {
			RedisConfig tokenCache  `json:"redis_config"`
			MongoConfig mongoConfig `json:"mongo_config"`
			AppName     string      `json:"app_name"`
		}

		jsonData := `{
			"redis_config": "${env.redis_config}",
			"mongo_config": "${env.mongo_config}",
			"app_name": "test-app"
		}`

		var holder multiConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		// Verify Redis config
		assert.Equal(t, "redis:6379", holder.RedisConfig.Address)
		assert.Equal(t, "redis-pass", holder.RedisConfig.Password)
		assert.Equal(t, "0", holder.RedisConfig.DB)

		// Verify Mongo config
		assert.Equal(t, "mongo.example.com", holder.MongoConfig.Host)
		assert.Equal(t, 27017, holder.MongoConfig.Port)
		assert.Equal(t, "myapp", holder.MongoConfig.Database)

		// Verify regular string value
		assert.Equal(t, "test-app", holder.AppName)
	})

	// Test Case 6: JSON with escaped quotes
	t.Run("JSONWithEscapedQuotes", func(t *testing.T) {
		jsonWithEscapes := `{"message":"Hello \"World\"","path":"C:\\Users\\test","regex":"\\d+"}`
		os.Setenv("special_config", jsonWithEscapes)
		defer os.Unsetenv("special_config")

		type specialConfig struct {
			Message string `json:"message"`
			Path    string `json:"path"`
			Regex   string `json:"regex"`
		}

		type wrapperConfig struct {
			Special specialConfig `json:"special"`
		}

		jsonData := `{"special": "${env.special_config}"}`

		var holder wrapperConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		assert.Equal(t, `Hello "World"`, holder.Special.Message)
		assert.Equal(t, `C:\Users\test`, holder.Special.Path)
		assert.Equal(t, `\d+`, holder.Special.Regex)
	})

	// Test Case 7: JSON object mixed with regular env vars
	t.Run("JSONObjectMixedWithRegularEnvVars", func(t *testing.T) {
		os.Setenv("redis_config", `{"address":"mixed-redis:6379","password":"mixed-pass","db":"3"}`)
		os.Setenv("app_port", "9000")
		os.Setenv("debug_mode", "true")
		defer func() {
			os.Unsetenv("redis_config")
			os.Unsetenv("app_port")
			os.Unsetenv("debug_mode")
		}()

		type mixedConfig struct {
			AppPort   string     `json:"app_port"`
			DebugMode string     `json:"debug_mode"`
			Redis     tokenCache `json:"redis"`
		}

		jsonData := `{
			"app_port": "${env.app_port}",
			"debug_mode": "${env.debug_mode}",
			"redis": "${env.redis_config}"
		}`

		var holder mixedConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		assert.Equal(t, "9000", holder.AppPort)
		assert.Equal(t, "true", holder.DebugMode)
		assert.Equal(t, "mixed-redis:6379", holder.Redis.Address)
		assert.Equal(t, "mixed-pass", holder.Redis.Password)
		assert.Equal(t, "3", holder.Redis.DB)
	})

	// Test Case 8: Empty JSON object
	t.Run("EmptyJSONObject", func(t *testing.T) {
		os.Setenv("empty_config", `{}`)
		defer os.Unsetenv("empty_config")

		type emptyStruct struct{}
		type wrapperConfig struct {
			Empty emptyStruct `json:"empty"`
		}

		jsonData := `{"empty": "${env.empty_config}"}`

		var holder wrapperConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)

		// Should not error and should have empty struct
		assert.Equal(t, emptyStruct{}, holder.Empty)
	})

	// Test Case 9: Invalid JSON should remain as string
	t.Run("InvalidJSONRemainsAsString", func(t *testing.T) {
		os.Setenv("invalid_json", `{"invalid": json}`) // Missing quotes around json
		defer os.Unsetenv("invalid_json")

		type configWithString struct {
			Data string `json:"data"`
		}

		jsonData := `{"data": "${env.invalid_json}"}`

		var holder configWithString
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)

		// The function should succeed and the invalid JSON should remain quoted
		assert.Error(t, err)
	})
}

// TestQuotedEnvVarTypeCasting tests the type-aware env var replacement feature.
// This is the main feature test for fixing "${env.PORT}" → 8080 (int) instead of "8080" (string).
func TestQuotedEnvVarTypeCasting(t *testing.T) {
	// TC-001: Quoted int to int
	t.Run("TC-001_QuotedIntToInt", func(t *testing.T) {
		os.Setenv("PORT", "8080")
		defer os.Unsetenv("PORT")

		type serverConfig struct {
			Port int `json:"port"`
		}

		jsonData := `{"port": "${env.PORT}"}`

		var holder serverConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, 8080, holder.Port)
	})

	// TC-002: Quoted bool to bool
	t.Run("TC-002_QuotedBoolToBool", func(t *testing.T) {
		os.Setenv("ENABLED", "true")
		defer os.Unsetenv("ENABLED")

		type featureConfig struct {
			Enabled bool `json:"enabled"`
		}

		jsonData := `{"enabled": "${env.ENABLED}"}`

		var holder featureConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, true, holder.Enabled)
	})

	// TC-003: Quoted float to float
	t.Run("TC-003_QuotedFloatToFloat", func(t *testing.T) {
		os.Setenv("RATE", "3.14")
		defer os.Unsetenv("RATE")

		type rateConfig struct {
			Rate float64 `json:"rate"`
		}

		jsonData := `{"rate": "${env.RATE}"}`

		var holder rateConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, 3.14, holder.Rate)
	})

	// TC-004: String stays string
	t.Run("TC-004_StringStaysString", func(t *testing.T) {
		os.Setenv("NAME", "test-service")
		defer os.Unsetenv("NAME")

		type nameConfig struct {
			Name string `json:"name"`
		}

		jsonData := `{"name": "${env.NAME}"}`

		var holder nameConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, "test-service", holder.Name)
	})

	// TC-005: Unquoted still works (backward compatibility)
	t.Run("TC-005_UnquotedStillWorks", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		defer os.Unsetenv("PORT")

		type serverConfig struct {
			Port int `json:"port"`
		}

		// Unquoted placeholder - existing behavior
		jsonData := `{"port": ${env.PORT}}`

		var holder serverConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, 9090, holder.Port)
	})

	// TC-006: Invalid int value - should fail with error
	t.Run("TC-006_InvalidIntFails", func(t *testing.T) {
		os.Setenv("PORT", "not-a-number")
		defer os.Unsetenv("PORT")

		type serverConfig struct {
			Port int `json:"port"`
		}

		jsonData := `{"port": "${env.PORT}"}`

		var holder serverConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		// Should fail because "not-a-number" cannot be converted to int
		// The fixTypedPrimitives will log debug and keep as string, but final unmarshal will fail
		assert.Error(t, err)
	})

	// TC-007: Nested struct field conversion
	t.Run("TC-007_NestedStructFieldConversion", func(t *testing.T) {
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_MAX_CONNS", "100")
		defer func() {
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_MAX_CONNS")
		}()

		type dbConfig struct {
			Port     int `json:"port"`
			MaxConns int `json:"max_conns"`
		}

		type appConfig struct {
			Database dbConfig `json:"database"`
		}

		jsonData := `{
			"database": {
				"port": "${env.DB_PORT}",
				"max_conns": "${env.DB_MAX_CONNS}"
			}
		}`

		var holder appConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, 5432, holder.Database.Port)
		assert.Equal(t, 100, holder.Database.MaxConns)
	})

	// TC-008: Multiple types in same config
	t.Run("TC-008_MultipleTypesInSameConfig", func(t *testing.T) {
		os.Setenv("APP_PORT", "8080")
		os.Setenv("APP_DEBUG", "true")
		os.Setenv("APP_TIMEOUT", "30.5")
		os.Setenv("APP_NAME", "my-service")
		defer func() {
			os.Unsetenv("APP_PORT")
			os.Unsetenv("APP_DEBUG")
			os.Unsetenv("APP_TIMEOUT")
			os.Unsetenv("APP_NAME")
		}()

		type fullConfig struct {
			Port    int     `json:"port"`
			Debug   bool    `json:"debug"`
			Timeout float64 `json:"timeout"`
			Name    string  `json:"name"`
		}

		jsonData := `{
			"port": "${env.APP_PORT}",
			"debug": "${env.APP_DEBUG}",
			"timeout": "${env.APP_TIMEOUT}",
			"name": "${env.APP_NAME}"
		}`

		var holder fullConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, 8080, holder.Port)
		assert.Equal(t, true, holder.Debug)
		assert.Equal(t, 30.5, holder.Timeout)
		assert.Equal(t, "my-service", holder.Name)
	})

	// TC-009: Int64 and Uint64 types (within safe JSON number range)
	t.Run("TC-009_LargeIntegerTypes", func(t *testing.T) {
		// Use numbers within JavaScript's safe integer range (2^53 - 1 = 9007199254740991)
		// to avoid JSON precision loss
		os.Setenv("BIG_INT", "9007199254740991")
		os.Setenv("BIG_UINT", "9007199254740991")
		defer func() {
			os.Unsetenv("BIG_INT")
			os.Unsetenv("BIG_UINT")
		}()

		type largeConfig struct {
			BigInt  int64  `json:"big_int"`
			BigUint uint64 `json:"big_uint"`
		}

		jsonData := `{
			"big_int": "${env.BIG_INT}",
			"big_uint": "${env.BIG_UINT}"
		}`

		var holder largeConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, int64(9007199254740991), holder.BigInt)
		assert.Equal(t, uint64(9007199254740991), holder.BigUint)
	})

	// TC-010: Bool variations (true/false, 1/0)
	t.Run("TC-010_BoolVariations", func(t *testing.T) {
		t.Run("true_false", func(t *testing.T) {
			os.Setenv("FLAG", "false")
			defer os.Unsetenv("FLAG")

			type flagConfig struct {
				Flag bool `json:"flag"`
			}

			jsonData := `{"flag": "${env.FLAG}"}`

			var holder flagConfig
			_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
			assert.Nil(t, err)
			assert.Equal(t, false, holder.Flag)
		})

		t.Run("1_0", func(t *testing.T) {
			os.Setenv("FLAG", "1")
			defer os.Unsetenv("FLAG")

			type flagConfig struct {
				Flag bool `json:"flag"`
			}

			jsonData := `{"flag": "${env.FLAG}"}`

			var holder flagConfig
			_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
			assert.Nil(t, err)
			assert.Equal(t, true, holder.Flag)
		})
	})

	// TC-011: Embedded (anonymous) struct field conversion
	// This tests the real-world case of MediaMicroserviceConfiguration embedding MetricsBundleConfiguration
	t.Run("TC-011_EmbeddedStructFieldConversion", func(t *testing.T) {
		os.Setenv("TRACING_ENABLED", "true")
		os.Setenv("EXPORTER_PORT", "4317")
		defer func() {
			os.Unsetenv("TRACING_ENABLED")
			os.Unsetenv("EXPORTER_PORT")
		}()

		// Embedded (anonymous) struct - similar to how bundles.MetricsBundleConfiguration is embedded
		type MetricsConfig struct {
			Enabled      bool `json:"tracing_enabled"`
			ExporterPort int  `json:"exporter_port"`
		}

		type MicroserviceConfig struct {
			ListenAddr    string                  `json:"listen_addr"`
			MetricsConfig `json:"metrics_bundle"` // EMBEDDED (anonymous) field
		}

		jsonData := `{
			"listen_addr": "localhost:8080",
			"metrics_bundle": {
				"tracing_enabled": "${env.TRACING_ENABLED}",
				"exporter_port": "${env.EXPORTER_PORT}"
			}
		}`

		var holder MicroserviceConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, "localhost:8080", holder.ListenAddr)
		assert.Equal(t, true, holder.Enabled, "embedded bool should be converted from string")
		assert.Equal(t, 4317, holder.ExporterPort, "embedded int should be converted from string")
	})

	// TC-012: Real-world MediaMicroserviceConfiguration structure reproduction
	// Tests the actual embedded bundle pattern with multiple nested embedded structs
	t.Run("TC-012_RealWorldMediaConfigPattern", func(t *testing.T) {
		os.Setenv("TRACING_ENABLED", "true")
		os.Setenv("APP_NAME", "test-service")
		defer func() {
			os.Unsetenv("TRACING_ENABLED")
			os.Unsetenv("APP_NAME")
		}()

		// Simulate the real bundle configuration structures
		type TracerConfiguration struct {
			ServiceName string `json:"service_name"`
		}

		type ExporterConfiguration struct {
			CollectorURL string `json:"collector_url"`
			Secure       bool   `json:"secure"`
		}

		type MetricsBundleConfiguration struct {
			Enabled               bool                  `json:"tracing_enabled"`
			TracerConfiguration   TracerConfiguration   `json:"metrics_tracer"`
			ExporterConfiguration ExporterConfiguration `json:"trace_exporter"`
		}

		type AuthorizationConfig struct {
			JwtHeaderKey string `json:"jwt_header_key"`
		}

		// The actual MediaMicroserviceConfiguration pattern with multiple embedded bundles
		type MediaMicroserviceConfiguration struct {
			ListenAddr                 string `json:"listen_addr"`
			ListenMetricsAddr          string `json:"listen_metrics_addr"`
			MetricsBundleConfiguration `json:"metrics_bundle"`
			AuthorizationConfig        `json:"authorization_gates_bundle"`
		}

		jsonData := `{
			"listen_addr": ":8080",
			"listen_metrics_addr": ":3000",
			"metrics_bundle": {
				"tracing_enabled": "${env.TRACING_ENABLED}",
				"metrics_tracer": {
					"service_name": "${env.APP_NAME}"
				},
				"trace_exporter": {
					"collector_url": "http://localhost:4317",
					"secure": false
				}
			},
			"authorization_gates_bundle": {
				"jwt_header_key": "authorization"
			}
		}`

		var holder MediaMicroserviceConfiguration
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, ":8080", holder.ListenAddr)
		assert.Equal(t, true, holder.Enabled, "embedded bool in metrics_bundle should be converted from string")
		assert.Equal(t, "test-service", holder.TracerConfiguration.ServiceName)
	})

	// TC-013: Double pointer with nil inner pointer (bootstrap scenario)
	// This tests the exact scenario from microservice.Bootstrap where:
	// - var cfg *Config (nil pointer)
	// - configuration.Setup(&cfg, ...) passes **Config
	t.Run("TC-013_DoublePointerNilBootstrapScenario", func(t *testing.T) {
		os.Setenv("ENABLED", "true")
		os.Setenv("COUNT", "42")
		defer func() {
			os.Unsetenv("ENABLED")
			os.Unsetenv("COUNT")
		}()

		type NestedConfig struct {
			Enabled bool `json:"enabled"`
			Count   int  `json:"count"`
		}

		type Config struct {
			Name   string       `json:"name"`
			Nested NestedConfig `json:"nested"`
		}

		jsonData := `{
			"name": "test",
			"nested": {
				"enabled": "${env.ENABLED}",
				"count": "${env.COUNT}"
			}
		}`

		// Simulate the bootstrap scenario: var cfg *Config; Setup(&cfg, ...)
		var cfg *Config // nil pointer
		holder := &cfg  // double pointer: **Config

		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), holder, logger.Logger)
		assert.Nil(t, err)
		assert.NotNil(t, *holder, "inner pointer should be initialized")
		assert.Equal(t, "test", (*holder).Name)
		assert.Equal(t, true, (*holder).Nested.Enabled, "nested bool should be converted from string")
		assert.Equal(t, 42, (*holder).Nested.Count, "nested int should be converted from string")
	})

	// TC-014: Map with struct values containing typed fields (recaptcha pattern)
	// This tests the real-world case of recaptcha_gates_bundle.clients which is map[string]RecaptchaAPIConfiguration
	// where RecaptchaAPIConfiguration.threshold is float32
	t.Run("TC-014_MapWithStructValuesTypedFields", func(t *testing.T) {
		os.Setenv("RECAPTCHA_THRESHOLD", "0.5")
		defer os.Unsetenv("RECAPTCHA_THRESHOLD")

		// Mimics RecaptchaAPIConfiguration from infra/apis/recaptcha_api.go
		type RecaptchaConfig struct {
			SiteKey   string  `json:"site_key"`
			SecretKey string  `json:"secret_key"`
			Threshold float32 `json:"threshold"`
		}

		type RecaptchaBundle struct {
			Clients map[string]RecaptchaConfig `json:"clients"`
		}

		type AppConfig struct {
			RecaptchaGatesBundle RecaptchaBundle `json:"recaptcha_gates_bundle"`
		}

		jsonData := `{
			"recaptcha_gates_bundle": {
				"clients": {
					"webClient": {
						"site_key": "web-key",
						"secret_key": "web-secret",
						"threshold": "${env.RECAPTCHA_THRESHOLD}"
					},
					"iosClient": {
						"site_key": "ios-key",
						"secret_key": "ios-secret",
						"threshold": "${env.RECAPTCHA_THRESHOLD}"
					}
				}
			}
		}`

		var holder AppConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err)
		assert.Equal(t, float32(0.5), holder.RecaptchaGatesBundle.Clients["webClient"].Threshold, "map value struct float should be converted from string")
		assert.Equal(t, float32(0.5), holder.RecaptchaGatesBundle.Clients["iosClient"].Threshold, "map value struct float should be converted from string")
		assert.Equal(t, "web-key", holder.RecaptchaGatesBundle.Clients["webClient"].SiteKey)
		assert.Equal(t, "ios-key", holder.RecaptchaGatesBundle.Clients["iosClient"].SiteKey)
	})

	// TC-015: Anonymous (embedded) struct fields must have their typed primitives coerced.
	// This reproduces the ms-ai-integrations runtime error:
	//   json: cannot unmarshal number " into Go struct field RedisCacheConfiguration.DB of type int
	// Root cause: fixTypedPrimitivesRecursive was skipping anonymous/embedded struct fields,
	// so env-substituted string values like "0" were never converted to int before unmarshalling.
	t.Run("TC-015_AnonymousEmbeddedStructTypedFieldCoercion", func(t *testing.T) {
		os.Setenv("REDIS_DB", "2")
		defer os.Unsetenv("REDIS_DB")

		// Mimics infra/cache.RedisCacheConfiguration
		type RedisCacheConfig struct {
			Host string `json:"host"`
			Port int    `json:"port"`
			DB   int    `json:"db"`
		}

		// Mimics bundles.TokenServicesBundleConfiguration — embeds RedisCacheConfig anonymously
		type TokenBundle struct {
			RedisCacheConfig // anonymous/embedded
		}

		// Mimics cmd/ms/ms_ai_integrations/applicationConfiguration — embeds TokenBundle anonymously
		type AppConfig struct {
			TokenBundle        // anonymous/embedded
			AppName     string `json:"app_name"`
		}

		// JSON is flat (embedded fields are promoted to the top level)
		jsonData := `{
			"app_name": "ai-integrations",
			"host": "localhost",
			"port": 6379,
			"db": "${env.REDIS_DB}"
		}`

		var holder AppConfig
		_, err := StructFromJSONBytesWithEnvReplace([]byte(jsonData), &holder, logger.Logger)
		assert.Nil(t, err, "should parse without error even with env-substituted int in embedded struct")
		assert.Equal(t, "ai-integrations", holder.AppName)
		assert.Equal(t, "localhost", holder.Host)
		assert.Equal(t, 6379, holder.Port)
		assert.Equal(t, 2, holder.DB, "embedded struct int field must be coerced from env string")
	})
}
