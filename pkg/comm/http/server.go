package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	recoverMdw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pixie-sh/di-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
)

type CORSConfig = cors.Config

// ServerHandler type alias
type ServerHandler = fiber.Handler

// ServerGroup type alias
type ServerGroup = fiber.Router

// App type alias
type App = fiber.App

// ServerCtx type alias
type ServerCtx = *fiber.Ctx

// ServerConfiguration alias
type ServerConfiguration fiber.Config

func (s *ServerConfiguration) LookupNode(lookupPath string) (any, error) {
	return di.ConfigurationNodeLookup(s, lookupPath)
}

// DefaultServerConfiguration default config with error handling
var DefaultServerConfiguration = ServerConfiguration{
	ErrorHandler: func(ctx ServerCtx, errInput error) error {
		return errorHandlerWithCustomProcessor(ctx, errInput)
	},
}

// server just a wrapper for fiber.App
type server struct {
	*App
}

type Server interface {
	fiber.Router
	Listen(addr string) error
}

// NewServer returns new server based on Server with given handlers server wide
// health endpoint is set unprotected
func NewServer(config ServerConfiguration, serverWideHandlers ...ServerHandler) Server {
	fiberServer := fiber.New([]fiber.Config{fiber.Config(config)}...)

	if env.IsDebugActive() {
		cfg := recoverMdw.Config{
			EnableStackTrace: true,
			StackTraceHandler: func(c ServerCtx, e interface{}) {
				logger.With("st", e).Error("panic recover stack trace")
			},
		}

		fiberServer.Use(recoverMdw.New(cfg))
	} else {
		fiberServer.Use(recoverMdw.New(recoverMdw.ConfigDefault))
	}

	var i []interface{}
	for _, handler := range serverWideHandlers {
		i = append(i, handler)
	}
	if len(i) > 0 {
		fiberServer.Use(i...)
	}

	return &server{fiberServer}
}

// Group create a group for routes
func (s *server) Group(groupPrefix string, groupHandlers ...ServerHandler) ServerGroup {
	return s.App.Group(groupPrefix, groupHandlers...)
}
