# core-go

The core libraries behind Pixie's Go services: configuration, HTTP and Lambda entry points, events, caching, rate limiting, state machines and a set of small typed helpers.

[![Go Reference](https://pkg.go.dev/badge/github.com/pixie-sh/core-go.svg)](https://pkg.go.dev/github.com/pixie-sh/core-go)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Made by [Pixie](https://pixie.sh). Project page: <https://pixie-sh.github.io/core-go/>

## Install

```sh
go get github.com/pixie-sh/core-go
```

Import only the packages you need; each one lives under `pkg/` or `infra/`.

## How it fits together

core-go sits on top of the other Pixie libraries and wires them into a service:

| Library | Role |
| --- | --- |
| [logger-go](https://github.com/pixie-sh/logger-go) | structured logging |
| [errors-go](https://github.com/pixie-sh/errors-go) | errors with codes and HTTP status |
| [di-go](https://github.com/pixie-sh/di-go) | dependency injection and configuration lookup |
| [ulid-go](https://github.com/pixie-sh/ulid-go) | scoped, time-sortable IDs |
| [database-helpers-go](https://github.com/pixie-sh/database-helpers-go) | GORM helpers and query pipelines |
| [clone-go](https://github.com/pixie-sh/clone-go) | deep clone |

- `pkg/` holds building blocks with no infrastructure behind them: configuration, HTTP server, Lambda handlers, models, types.
- `infra/` holds components that talk to something: Redis, Kafka, SQS, the database.

## Usage

### Configuration with environment placeholders

Load JSON or TOML into a struct and fill `${env.NAME}` placeholders from the environment.

```go
import (
	"github.com/pixie-sh/core-go/pkg/configuration"
	"github.com/pixie-sh/logger-go/logger"
)

type Config struct {
	Port     int `json:"port"`
	Database struct {
		DSN string `json:"dsn"`
	} `json:"database"`
}

var cfg Config
_, err := configuration.StructFromJSONBytesWithEnvReplace(
	[]byte(`{"port": 8080, "database": {"dsn": "${env.DB_DSN}"}}`),
	&cfg,
	logger.Logger,
)
// with DB_DSN=postgres://app@db:5432/app
// cfg.Port == 8080, cfg.Database.DSN == "postgres://app@db:5432/app"
```

`StructFromFileWithEnvReplace` and `StructFromTOMLBytesWithEnvReplace` do the same from a file or from TOML. `SetEnvFromFile` loads a `.env` file first.

### IDs

`pkg/uid` wraps ulid-go so services share one ID type.

```go
import "github.com/pixie-sh/core-go/pkg/uid"

const ScopeInvoice uid.Scope = 7

id := uid.NewScoped(ScopeInvoice)
fmt.Println(id.String())      // 26-character ULID
scope, _ := id.Scope()        // 7

same, err := uid.FromString(id.String())
```

### Slices, sets and maps

```go
import (
	"github.com/pixie-sh/core-go/pkg/types/sets"
	"github.com/pixie-sh/core-go/pkg/types/slices"
)

names := slices.Map([]int{1, 2, 2, 3}, func(i int) string {
	return fmt.Sprintf("user-%d", i)
})

slices.Deduplicate(names)       // [user-1 user-2 user-3]
slices.Contains(names, "user-3") // true

seen := sets.From([]string{"a", "b", "a"}) // 2 items
_, ok := sets.Contains(seen, "b")          // true
```

`slices` also has `Find`, `MapOrError`, `ShuffleSlice` and `SortByCustomOrder`; `maps` has `MapKeys`, `MapValues` and friends.

### Services

`pkg/microservice` bootstraps a service from its configuration through di-go, starts it, and handles shutdown signals. A service implements `Starter`:

```go
ms := microservice.BootstrapMust[*BillingService, *BillingConfiguration](ctx)
```

`SetupHealthController`, `SetupMetricsController` and `SetupPprofControllers` add the standard endpoints to the HTTP server from `pkg/comm/http`.

### Events

`infra/events` defines typed events and the producer and consumer interfaces, with Kafka and SQS implementations in `infra/events/kafka` and `infra/events/sqs`.

```go
event := events.NewEventWrapper("evt-1", "invoice.paid", InvoicePaid{InvoiceID: id})
```

`events.Emit` and `events.EmitBatch` send through the default producers; `Forwarder` routes consumed events to handlers by type.

## Package map

### pkg

| Package | What it does |
| --- | --- |
| `configuration` | load JSON/TOML config with environment replacement and references |
| `comm/http` | Fiber-based HTTP server, request parsing, API responses |
| `comm/http_middlewares` | logging, metrics, tracing and CORS middlewares |
| `comm/rest` | outbound REST client, proxying, gzip and Brotli decoding |
| `lambda`, `lambda/router`, `lambda/lambda_api`, `lambda/lambda_sqs` | AWS Lambda entry points for API Gateway, SQS and Kafka, with routing and request helpers |
| `microservice` | service bootstrap, health, metrics and pprof endpoints, graceful shutdown |
| `context` | logger and trace ID carried in `context.Context` |
| `metrics` | tracer, exporter and metrics registry |
| `models` and subpackages | shared request, response, pagination and database models; serialisation and validators |
| `notifications` | email (SMTP, Mailgun), SMS and push notification services |
| `pubsub` | in-process publisher and subscriber |
| `s3` | S3 client |
| `files` | downloads, MIME detection, CSV to structs, image resizing |
| `templates` | template expansion with helpers and partials |
| `geometry` | points, WKB and Open Location Code helpers |
| `uid` | ID type built on ulid-go |
| `types` and subpackages | generic helpers for slices, sets, maps, strings, dates, decimals and channels |
| `pagination` | process a paginated source page by page |
| `hash`, `base64`, `urls`, `validation`, `time`, `utils`, `env` | small utilities |
| `errors`, `errors/db_errors` | error codes and database error mapping on top of errors-go |

### infra

| Package | What it does |
| --- | --- |
| `cache` | Redis cache, batch operations and distributed locks |
| `rate_limiter` | request and limit-based rate limiting on top of a cache |
| `events`, `events/kafka`, `events/sqs` | event model, producers, consumers and forwarding |
| `message_buses`, `message_router`, `message_factory`, `message_wrapper` | typed messages, routing and broadcasting between sources |
| `state_machine` | state machines with transitions, conditions and persisted history |
| `cron` | scheduled jobs |
| `database_migrator` | run ordered database migrations as a service |
| `geohash` | geohash seeding and lookup |
| `tags` | tagging of entities with activity logs |
| `apis` | shared API configuration and error handling |
| `uidgen`, `version` | ID generation and build version |

## Development

```sh
make test
```

## License

[MIT](LICENSE)
