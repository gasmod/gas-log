# gas-log

Logging backends for the [Gas](https://github.com/gasmod/gas) ecosystem. Implements the `gas.Logger`, `gas.LogEvent`, and `gas.LoggerContext` interfaces with three interchangeable backends.

```
go get github.com/gasmod/gas-log
```

## Backends

| Backend     | Constructor                                                    | Backing library                             | Notes                                                                         |
|-------------|----------------------------------------------------------------|---------------------------------------------|-------------------------------------------------------------------------------|
| **Zerolog** | `NewZeroLogLogger(logger *zerolog.Logger)`                     | [rs/zerolog](https://github.com/rs/zerolog) | High-performance structured JSON logging. Full level support including Trace. |
| **Slog**    | `NewSlogLogger(logger *slog.Logger, eventInitialCapacity int)` | `log/slog` (stdlib)                         | Zero-dependency option. Trace maps to Debug (slog has no Trace level).        |
| **NoOp**    | `NewNoOpLogger()`                                              | none                                        | Silently discards all output. Singleton, zero-allocation. Useful for tests.   |

## Usage

### Zerolog

```go
package main

import (
	"os"

	"github.com/rs/zerolog"
	gaslog "github.com/gasmod/gas-log"
)

func main() {
	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger := gaslog.NewZeroLogLogger(&zl)

	logger.Info("server started").Str("addr", ":8080").Send()
	// {"level":"info","addr":":8080","time":"...","message":"server started"}
}
```

### Slog

```go
package main

import (
	"log/slog"
	"os"

	gaslog "github.com/gasmod/gas-log"
)

func main() {
	sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := gaslog.NewSlogLogger(sl, 5) // 5 = initial attr capacity per event

	logger.Info("server started").Str("addr", ":8080").Send()
	// {"level":"INFO","msg":"server started","addr":":8080"}
}
```

### NoOp

```go
logger := gaslog.NewNoOpLogger() // all calls are no-ops
logger.Error("this goes nowhere").Send()
```

## Fluent API

All backends share the same fluent interface defined by `gas.Logger`:

```go
// Create a log event, chain fields, then send
logger.Info("request handled").
    Str("method", "GET").
    Str("path", "/api/users").
    Int("status", 200).
    Duration("latency", elapsed).
    Send()

// Create a sub-logger with persistent fields
reqLogger := logger.With().
    Str("request_id", reqID).
    Str("service", "auth").
    Logger()

reqLogger.Debug("validating token").Send()
```

### Available field methods

| Method     | Signature                         |
|------------|-----------------------------------|
| `Str`      | `(key, val string)`               |
| `Int`      | `(key string, val int)`           |
| `Int64`    | `(key string, val int64)`         |
| `Float64`  | `(key string, val float64)`       |
| `Bool`     | `(key string, val bool)`          |
| `Err`      | `(key string, val error)`         |
| `Duration` | `(key string, val time.Duration)` |
| `Any`      | `(key string, val any)`           |

These methods are available on both `gas.LogEvent` (returned by level methods) and `gas.LoggerContext` (returned by `With()`).

## DI Integration

Register a logger in the Gas DI container so services can receive it via constructor injection:

```go
package main

import (
	"os"

	"github.com/gasmod/gas"
	gaslog "github.com/gasmod/gas-log"
	"github.com/rs/zerolog"
)

func main() {
	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger := gaslog.NewZeroLogLogger(&zl)

	app := gas.NewApp(
		gas.WithServiceInstance[gas.Logger](logger),
		// ...
	)
	app.Run()
}
```

Services receive the logger through their constructor:

```go
type MyService struct {
	logger gas.Logger
}

func NewMyService(logger gas.Logger) *MyService {
	return &MyService{logger: logger}
}

func (s *MyService) Init() error {
	s.logger.Info("service initialized").Str("name", s.Name()).Send()
	return nil
}
```

### Context-scoped logging

Use `gas.WithLogger` and `gas.LoggerFromContext` to propagate loggers through `context.Context`:

```go
func (s *MyService) handleRequest(w http.ResponseWriter, r *http.Request) {
	reqLogger := s.logger.With().
		Str("request_id", r.Header.Get("X-Request-ID")).
		Logger()

	ctx := gas.WithLogger(r.Context(), reqLogger)
	s.process(ctx)
}

func (s *MyService) process(ctx context.Context) {
	logger := gas.LoggerFromContext(ctx)
	logger.Debug("processing").Send()
}
```

## Level Mapping

| gas.Logger method | Zerolog level        | Slog level        |
|-------------------|----------------------|-------------------|
| `Trace`           | `zerolog.TraceLevel` | `slog.LevelDebug` |
| `Debug`           | `zerolog.DebugLevel` | `slog.LevelDebug` |
| `Info`            | `zerolog.InfoLevel`  | `slog.LevelInfo`  |
| `Warn`            | `zerolog.WarnLevel`  | `slog.LevelWarn`  |
| `Error`           | `zerolog.ErrorLevel` | `slog.LevelError` |
