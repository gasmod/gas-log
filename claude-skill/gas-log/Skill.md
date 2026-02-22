---
name: gas-log
description: >
  Reference documentation for the gas-log Go package
  (github.com/gasmod/gas-log) — pluggable logging backends for the Gas
  ecosystem. Use this skill when writing, reviewing, or debugging Go code
  involving structured logging in Gas services. Covers the three backends
  (ZeroLogLogger, SlogLogger, NoOpLogger), the fluent gas.Logger/LogEvent/
  LoggerContext/MutableLoggerContext interfaces, sub-loggers via With(),
  in-place logger mutation via SetBaseFields()/Apply(), context-scoped
  logging via gas.WithLogger/gas.LoggerFromContext, DI registration via
  gas.WithServiceInstance, and level mapping across backends.
---

# Gas Log Package Reference

Pluggable logging backends for the Gas ecosystem. Implements `gas.Logger`,
`gas.LogEvent`, `gas.LoggerContext`, and `gas.MutableLoggerContext` with three interchangeable backends.

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

### Field methods (on LogEvent, LoggerContext, and MutableLoggerContext)

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

## Mutating base fields (SetBaseFields)

Unlike `With()` which branches into a new sub-logger, `SetBaseFields()` accumulates
fields and on `Apply()` mutates the originating logger in-place.

**When to use `SetBaseFields` vs `With`:**
- Use `With()` when you want a new, independent sub-logger (e.g. a per-request logger passed down to child calls).
- Use `SetBaseFields()` when middleware owns one logger instance and needs to stamp persistent fields onto it before the rest of the request runs.

```go
// In middleware: mutate the shared logger in-place, then continue
logger.SetBaseFields().
    Str("request_id", reqID).
    Str("user_id", userID).
    Apply()
// all subsequent events from logger include request_id and user_id
```

Typical middleware pattern:

```go
func loggingMiddleware(logger gas.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.SetBaseFields().
                Str("request_id", r.Header.Get("X-Request-ID")).
                Str("method", r.Method).
                Str("path", r.URL.Path).
                Apply()
            next.ServeHTTP(w, r)
        })
    }
}
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
