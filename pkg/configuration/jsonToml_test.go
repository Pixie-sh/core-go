package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
