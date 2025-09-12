package env

import (
	"os"

	loggerEnv "github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
)

// UserAgent to be used in the proxy
const UserAgent = "USER_AGENT"

// LogLevel get log level from env
func LogLevel() logger.LogLevelEnum {
	switch os.Getenv("LOG_LEVEL") {
	case "WARN":
		return logger.WARN
	case "ERROR":
		return logger.ERROR
	case "DEBUG":
		return logger.DEBUG
	default:
		return logger.LOG
	}
}

func IsConfigChecksActive() bool {
	return os.Getenv("CONFIG_CHECKS") == "true" ||
		os.Getenv("CONFIG_CHECKS") == "TRUE" ||
		os.Getenv("CONFIG_CHECKS") == "1"
}

func EnvScope() string {
	return loggerEnv.EnvScope()
}

func EnvAppName() string {
	return loggerEnv.EnvAppName()
}

func EnvUserAgent() string {
	return os.Getenv(UserAgent)
}

func SetScope(scope string) {
	_ = os.Setenv(loggerEnv.Scope, scope)
}

func SetAppName(appName string) {
	_ = os.Setenv(loggerEnv.AppName, appName)
}

func SetUserAgent(userAgent string) {
	_ = os.Setenv(UserAgent, userAgent)
}

func EnvLogParser() string {
	return loggerEnv.EnvLogParser()
}

func SetLogParser(logParser string) {
	_ = os.Setenv(loggerEnv.LogParser, logParser)
}
