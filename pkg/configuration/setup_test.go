package configuration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/pixie-sh/database-helpers-go/database"
	loggerEnv "github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
	"github.com/pixie-sh/logger-go/mapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pixie-sh/core-go/pkg/env"
	"github.com/pixie-sh/core-go/pkg/lambda"
)

type meh struct {
	ID        uint   ``
	FirstName string ``
	LastName  string ``
}

func TestSetup(t *testing.T) {
	_ = os.Setenv(loggerEnv.AppName, "TestSetup")
	_ = os.Setenv(loggerEnv.AppVersion, "unversioned")
	_ = os.Setenv(loggerEnv.Scope, "unscoped")
	_ = os.Setenv(env.EnvConfigs, "/tmp/meh.toml")

	toDefer := createMehConfigFile()
	defer toDefer()

	// load up configuration
	var cfg meh
	Setup(&cfg, true)

	logger.With("meh", cfg).Log("config loaded")
	assert.Equal(t, meh{ID: 1, FirstName: "John", LastName: "Doe"}, cfg)
}

func createMehConfigFile() func() {
	// Create a new meh struct with desired values
	myData := meh{ID: 1, FirstName: "John", LastName: "Doe"}

	// Create a new file
	f, err := os.Create("/tmp/meh.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Encode the struct to TOML format
	encoder := toml.NewEncoder(f)
	err = encoder.Encode(myData)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully wrote TOML config file")
	return func() {
		_ = os.Remove("/tmp/meh.toml")
		log.Println("Successfully removed TOML config file")
	}
}

// config represents the configuration structure
type config struct {
	Database database.Configuration `json:"database"`
	Port     int                    `json:"port"`
	Bool     bool                   `json:"bool"`
}

func TestEnvOverload(t *testing.T) {
	_ = os.Setenv("DBPASSWORD", "SuperStrong")
	_ = os.Setenv("PORT", "7777")
	_ = os.Setenv("BOOL", "true")

	content := `
port = ${env.PORT}
bool = ${env.BOOL}

[database]
driver = "gorm_db_driver"
[database.values]
driver = "mysql_db_driver"
dsn = "auth_user:${env.DBPASSWORD}@tcp(db:3306)/dev-db"
`
	var c config
	_, err := StructFromTOMLBytesWithEnvReplace([]byte(content), &c, logger.Logger)
	assert.Nil(t, err)

	assert.Equal(t, c.Port, 7777)
	assert.Equal(t, c.Bool, true)
	assert.Equal(t, c.Database.Driver, "gorm_db_driver")

	var gormCfg database.GormDbConfiguration
	err = mapper.ObjectToStruct(c.Database.Values, &gormCfg)
	assert.Nil(t, err)

	assert.Equal(t, gormCfg.Driver, "mysql_db_driver")
	assert.Equal(t, gormCfg.Dsn, "auth_user:SuperStrong@tcp(db:3306)/dev-db")
}

func TestEnvOverloadWithJSON(t *testing.T) {
	_ = os.Setenv("DBPASSWORD", "SuperStrong")
	_ = os.Setenv("PORT", "7777")
	_ = os.Setenv("BOOL", "true")

	content := `{
  "port": ${env.PORT},
  "bool": ${env.BOOL},
  "database": {
    "driver": "gorm_db_driver",
    "values": {
      "driver": "mysql_db_driver",
      "dsn": "auth_user:${env.DBPASSWORD}@tcp(db:3306)/dev-db"
    }
  }
}
`
	var c config
	_, err := StructFromJSONBytesWithEnvReplace([]byte(content), &c, logger.Logger)
	assert.Nil(t, err)

	assert.Equal(t, c.Port, 7777)
	assert.Equal(t, c.Bool, true)
	assert.Equal(t, c.Database.Driver, "gorm_db_driver")

	var gormCfg database.GormDbConfiguration
	err = mapper.ObjectToStruct(c.Database.Values, &gormCfg)
	assert.Nil(t, err)

	assert.Equal(t, gormCfg.Driver, "mysql_db_driver")
	assert.Equal(t, gormCfg.Dsn, "auth_user:SuperStrong@tcp(db:3306)/dev-db")
}

type KeyValConfiguration struct {
	Value  string `json:"value"`
	Header string `json:"header"`
}

type Configuration struct {
	lambda.AbstractLambdaConfiguration

	AuthorizationGate KeyValConfiguration               `json:"key_val_config"`
	Databases         map[string]database.Configuration `json:"databases"`
}

func TestSetup_MultiFileMergeOverridesRefBlock(t *testing.T) {
	dir := t.TempDir()

	base := filepath.Join(dir, "base.json")
	baseContent := `{
  "#ref": {
    "endpoint": {
      "host": "base-host",
      "port": 1111
    }
  },
  "service_endpoint": "${#ref.endpoint}"
}`
	require.NoError(t, os.WriteFile(base, []byte(baseContent), 0o644))

	overlay := filepath.Join(dir, "overlay.json")
	overlayContent := `{
  "#ref": {
    "endpoint": {
      "host": "overlay-host",
      "port": 2222
    }
  }
}`
	require.NoError(t, os.WriteFile(overlay, []byte(overlayContent), 0o644))

	t.Setenv(env.EnvConfigs, base+","+overlay)

	type endpoint struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	type holderCfg struct {
		ServiceEndpoint endpoint `json:"service_endpoint"`
	}

	var cfg holderCfg
	Setup(&cfg, true, false)

	assert.Equal(t, "overlay-host", cfg.ServiceEndpoint.Host, "overlay #ref.endpoint should override base #ref.endpoint")
	assert.Equal(t, 2222, cfg.ServiceEndpoint.Port, "overlay #ref.endpoint should override base #ref.endpoint")
}

func TestDb(t *testing.T) {
	content := `{
    "key_val_config": {
        "header": "X-Authkey",
        "value": "local"
    },
    "databases": {
        "ChatDomain": {
            "driver": "gorm_db_driver",
            "values": {
                "driver": "psql_db_driver",
                "dsn": "host=docker.for.mac.host.internal user=m password=mm dbname=mmm port=9998 sslmode=disable"
            }
        }
    }
}

`
	var c Configuration
	_, err := StructFromJSONBytesWithEnvReplace([]byte(content), &c, logger.Logger)
	assert.Nil(t, err)

	fmt.Println(c)
}
