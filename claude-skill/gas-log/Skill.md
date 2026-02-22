---
name: gas-log
description: >
  Reference documentation for the gas-log Go package
  (github.com/gasmod/gas-log) — pluggable logging backends for the Gas
  ecosystem. Use this skill when writing, reviewing, or debugging Go code
  involving structured logging in Gas services. Covers the three backends
  (ZeroLogLogger, SlogLogger, NoOpLogger), the fluent gas.Logger/LogEvent/
  LoggerContext interfaces, sub-loggers via With(), context-scoped logging
  via gas.WithLogger/gas.LoggerFromContext, DI registration via
  gas.WithServiceInstance, and level mapping across backends.
---

# Gas Log Package Reference

Pluggable logging backends for the Gas ecosystem. Implements `gas.Logger`,
`gas.LogEvent`, and `gas.LoggerContext` with three interchangeable backends.

```
import gaslog "github.com/gasmod/gas-log"
```

## Backends

| Backend     | Constructor                                                    | Backing library     | Notes                                                                 |
|-------------|----------------------------------------------------------------|---------------------|-----------------------------------------------------------------------|
| **Zerolog** | `NewZeroLogLogger(logger *zerolog.Logger)`                     | rs/zerolog          | High-performance structured JSON. Full level support including Trace. |
| **Slog**    | `NewSlogLogger(logger *slog.Logger, eventInitialCapacity int)` | `log/slog` (stdlib) | Zero external deps. Trace maps to Debug (slog has no Trace).          |
| **NoOp**    | `NewNoOpLogger()`                                              | none                | Discards all output. Singleton, zero-allocation. For tests.           |

## Fluent API

All backends share the same interface. Call a level method to get a `gas.LogEvent`,
chain fields, finalize with `Send()`:

```go
logger.Info("request handled").
    Str("method", r.Method).
    Int("status", code).
    Duration("latency", elapsed).
    Send()
```

### Field methods (on both LogEvent and LoggerContext)

```go
Str(key, val string)
Int(key string, val int)
Int64(key string, val int64)
Float64(key string, val float64)
Bool(key string, val bool)
Err(key string, val error)
Duration(key string, val time.Duration)
Any(key string, val any)
```

## Sub-loggers (With)

Create a sub-logger with persistent fields baked in:

```go
reqLogger := logger.With().
    Str("request_id", reqID).
    Str("service", "auth").
    Logger()

reqLogger.Debug("validating token").Send()
// all events from reqLogger include request_id and service
```

## Context-Scoped Logging

Propagate loggers through `context.Context` using helpers defined in gas core:

```go
// Store logger in context
ctx := gas.WithLogger(r.Context(), reqLogger)

// Retrieve from context (returns nil if absent)
logger := gas.LoggerFromContext(ctx)
logger.Debug("processing").Send()
```

Typical pattern in a request handler:

```go
func (s *Service) handleRequest(w http.ResponseWriter, r *http.Request) {
    reqLogger := s.logger.With().
        Str("request_id", r.Header.Get("X-Request-ID")).
        Logger()

    ctx := gas.WithLogger(r.Context(), reqLogger)
    s.process(ctx)
}

func (s *Service) process(ctx context.Context) {
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

## DI Registration

Register as a `gas.Logger` instance (not a service — loggers don't need
lifecycle management):

```go
zl := zerolog.New(os.Stdout).With().Timestamp().Logger()
logger := gaslog.NewZeroLogLogger(&zl)

app := gas.NewApp(
    gas.WithServiceInstance[gas.Logger](logger),
    // ...
)
```

Or with slog (zero external deps):

```go
sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger := gaslog.NewSlogLogger(sl, 5)

app := gas.NewApp(
    gas.WithServiceInstance[gas.Logger](logger),
)
```

Or NoOp for tests:

```go
app := gas.NewApp(
    gas.WithServiceInstance[gas.Logger](gaslog.NewNoOpLogger()),
)
```

## Consuming in Services

Services receive the logger through constructor injection:

```go
type Service struct {
    logger gas.Logger
    router *gas.Router
}

func New(logger gas.Logger, router *gas.Router) *Service {
    return &Service{logger: logger, router: router}
}

func (s *Service) Init() error {
    s.logger.Info("service initialized").Str("name", s.Name()).Send()
    s.router.Handle(s.Name(), "GET", "/hello", s.handleHello)
    return nil
}
```

## Backend Details

### ZeroLogLogger

```go
func NewZeroLogLogger(logger *zerolog.Logger) *ZeroLogLogger
```

Wraps a `zerolog.Logger`. Full level support. `Flush()` calls
`zerolog.Logger`'s writer flush if available.

### SlogLogger

```go
func NewSlogLogger(logger *slog.Logger, eventInitialCapacity int) *SlogLogger
```

`eventInitialCapacity` controls pre-allocated attribute slice capacity per
event. Reduces allocations when you know the typical field count. Values ≤ 0
default to 5. Each `LogEvent` collects fields and emits them as a single
`slog.LogAttrs` call on `Send()`.

### NoOpLogger

```go
func NewNoOpLogger() *NoOpLogger
```

Singleton pattern — all instances share the same underlying value. Zero
allocations per log call. All methods return chainable no-ops.
