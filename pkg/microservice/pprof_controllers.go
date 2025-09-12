package microservice

import (
	"context"
	netpprof "net/http/pprof"

	"github.com/pixie-sh/core-go/pkg/comm/http"
)

func SetupPprofControllers(_ context.Context, server http.Server) error {
	pprofGroup := server.Group("/debug")
	pprofGroup.Get("/pprof", http.GoHandlerAdaptor(netpprof.Index))
	pprofGroup.Get("/cmdline", http.GoHandlerAdaptor(netpprof.Cmdline))
	pprofGroup.Get("/profile", http.GoHandlerAdaptor(netpprof.Profile))
	pprofGroup.Get("/symbol", http.GoHandlerAdaptor(netpprof.Symbol))
	pprofGroup.Get("/trace", http.GoHandlerAdaptor(netpprof.Trace))
	pprofGroup.Get("/allocs", http.GoHandlerAdaptor(netpprof.Handler("allocs").ServeHTTP))
	pprofGroup.Get("/block", http.GoHandlerAdaptor(netpprof.Handler("block").ServeHTTP))
	pprofGroup.Get("/goroutine", http.GoHandlerAdaptor(netpprof.Handler("goroutine").ServeHTTP))
	pprofGroup.Get("/heap", http.GoHandlerAdaptor(netpprof.Handler("heap").ServeHTTP))
	pprofGroup.Get("/mutex", http.GoHandlerAdaptor(netpprof.Handler("mutex").ServeHTTP))
	pprofGroup.Get("/threadcreate", http.GoHandlerAdaptor(netpprof.Handler("threadcreate").ServeHTTP))

	return nil
}
