package http

import (
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/types"
)

func GetCtxLogger(c ServerCtx) logger.Interface {
	log := c.Locals(LocalsRequestLogger)
	if types.Nil(log) {
		return logger.Logger
	}

	castedLog, ok := log.(logger.Interface)
	if !ok {
		return logger.Logger
	}

	return castedLog
}
