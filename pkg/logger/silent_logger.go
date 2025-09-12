package logger

import (
	"context"

	"github.com/pixie-sh/logger-go/logger"
)

type SilentLogger struct{}

func (l *SilentLogger) With(string, any) logger.Interface {
	return l
}

func (l *SilentLogger) WithCtx(context.Context) logger.Interface {
	return l
}

func (l *SilentLogger) Clone() logger.Interface {
	return l
}

func (l *SilentLogger) Log(string, ...any) {

}

func (l *SilentLogger) Error(string, ...any) {

}

func (l *SilentLogger) Warn(string, ...any) {

}

func (l *SilentLogger) Debug(string, ...any) {

}
