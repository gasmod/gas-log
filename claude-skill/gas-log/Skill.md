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
  gas.WithService with ServiceLifetimeScoped, and level mapping across backends.
---

# Gas Log Package Reference

Pluggable logging backends for the Gas ecosystem. Implements `gas.Logger`,
`gas.LogEvent`, `gas.LoggerContext`, and `gas.MutableLoggerContext` with three interchangeable backends.

```
import gaslog "github.com/gasmod/gas-log"
```

## Backends

| Backend     | Constructor                                     | Backing library     | Notes                                                                 |
|-------------|-------------------------------------------------|---------------------|-----------------------------------------------------------------------|
| **Zerolog** | `NewZeroLogLogger(opts ...ZeroLogLoggerOption)` | rs/zerolog          | High-performance structured JSON. Full level support including Trace. |
| **Slog**    | `NewSlogLogger(opts ...SlogLoggerOption)`       | `log/slog` (stdlib) | Zero external deps. Trace maps to Debug (slog has no Trace).          |

Each constructor returns a constructor function type that the Gas DI container accepts directly. When no options are provided, backends use defaults: Zerolog uses the global `zerolog/log.Logger`; Slog uses `slog.Default()` with `eventInitialCapacity` of 5.

## Constructor Types and Options

### ZeroLogLogger

```go
// Constructor type
type ZeroLogLoggerCtor func() *ZeroLogLogger

// Constructor
func NewZeroLogLogger(opts ...ZeroLogLoggerOption) ZeroLogLoggerCtor

// Options
type ZeroLogLoggerOption func(*ZeroLogLogger)

func WithZeroLogInstance(logger *zerolog.Logger) ZeroLogLoggerOption
```

### SlogLogger

```go
// Constructor type
type SlogLoggerCtor func() *SlogLogger

// Constructor
func NewSlogLogger(opts ...SlogLoggerOption) SlogLoggerCtor

// Options
type SlogLoggerOption func(*SlogLogger)

func WithSlogInstance(logger *slog.Logger) SlogLoggerOption
func WithEventInitialCapacity(capacity int) SlogLoggerOption
```

`WithEventInitialCapacity` controls the pre-allocated attribute slice capacity per event. Reduces allocations when you know the typical field count. Values ≤ 0 default to 5. Each `LogEvent` collects fields and emits them as a single `slog.LogAttrs` call on `Send()`.

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

Pass the constructor function to `gas.WithService` with `gas.ServiceLifetimeScoped`.
The DI container calls it once per request scope to produce the logger instance:

```go
zl := zerolog.New(os.Stdout).With().Timestamp().Logger()

app := gas.NewApp(
    gas.WithService[gas.Logger](gaslog.NewZeroLogLogger(gaslog.WithZeroLogInstance(&zl)), gas.ServiceLifetimeScoped),
    // ...
)
```

With slog (zero external deps):

```go
sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))

app := gas.NewApp(
    gas.WithService[gas.Logger](gaslog.NewSlogLogger(
        gaslog.WithSlogInstance(sl),
        gaslog.WithEventInitialCapacity(5),
    ), gas.ServiceLifetimeScoped),
)
```

Using defaults (no options — Zerolog uses `log.Logger`, Slog uses `slog.Default()`):

```go
app := gas.NewApp(
    gas.WithService[gas.Logger](gaslog.NewZeroLogLogger(), gas.ServiceLifetimeScoped),
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
