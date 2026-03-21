---
name: gas-log
description: >
  Reference documentation for the gas-log Go package
  (github.com/gasmod/gas-log) — pluggable structured logging backends for the
  Gas ecosystem. Use this skill when writing, reviewing, or debugging Go code
  that involves structured logging in Gas services. Covers the two backends
  (ZeroLogLogger, SlogLogger), the fluent gas.Logger / gas.LogEvent /
  gas.LoggerContext / gas.MutableLoggerContext interfaces, constructor options,
  sub-loggers via With(), in-place logger mutation via SetBaseFields()/Apply(),
  context-scoped logging via gas.WithLogger / gas.LoggerFromContext, DI
  registration via gas.WithService with ServiceLifetimeScoped, and level
  mapping across backends. Make sure to use this skill whenever working with
  logging in a Gas application, even if the user doesn't explicitly mention
  "gas-log" — any code that imports gasmod/gas-log or references gas.Logger
  should trigger this skill.
---

# Gas Log Package Reference

Pluggable structured logging backends for the Gas ecosystem. Implements
`gas.Logger`, `gas.LogEvent`, `gas.LoggerContext`, and
`gas.MutableLoggerContext` with two interchangeable backends.

```
import gaslog "github.com/gasmod/gas-log"
```

> **Note:** A no-op logger (`gas.NewNopLogger()`) lives in gas core, not in
> this package. Use it for tests or when logging is disabled.

## Backends

| Backend     | Constructor                                     | Backing library     | Notes                                                                 |
|-------------|-------------------------------------------------|---------------------|-----------------------------------------------------------------------|
| **Zerolog** | `NewZeroLogLogger(opts ...ZeroLogLoggerOption)` | rs/zerolog          | High-performance structured JSON. Full level support including Trace. |
| **Slog**    | `NewSlogLogger(opts ...SlogLoggerOption)`       | `log/slog` (stdlib) | Zero external deps. Trace maps to Debug (slog has no Trace).          |

Each constructor returns a constructor function type (`ZeroLogLoggerCtor` /
`SlogLoggerCtor`) that the Gas DI container accepts directly. When no options
are provided, backends use sensible defaults: Zerolog uses the global
`zerolog/log.Logger`; Slog uses `slog.Default()` with `eventInitialCapacity`
of 5.

### Choosing a Backend

- **Zerolog** — prefer when you need Trace-level logging, high throughput, or
  already use zerolog elsewhere in your stack. Adds one external dependency.
- **Slog** — prefer when you want zero external dependencies (stdlib only) or
  your deployment already uses slog-based tooling. Trace and Debug both map to
  `slog.LevelDebug`.

## Constructor Types and Options

### ZeroLogLogger

```go
// Constructor type — passed to gas.WithService[gas.Logger](...)
type ZeroLogLoggerCtor func() *ZeroLogLogger

// Constructor — returns a ZeroLogLoggerCtor
func NewZeroLogLogger(opts ...ZeroLogLoggerOption) ZeroLogLoggerCtor

// Options
type ZeroLogLoggerOption func(*ZeroLogLogger)

func WithZeroLogInstance(logger *zerolog.Logger) ZeroLogLoggerOption
```

### SlogLogger

```go
// Constructor type — passed to gas.WithService[gas.Logger](...)
type SlogLoggerCtor func() *SlogLogger

// Constructor — returns a SlogLoggerCtor
func NewSlogLogger(opts ...SlogLoggerOption) SlogLoggerCtor

// Options
type SlogLoggerOption func(*SlogLogger)

func WithSlogInstance(logger *slog.Logger) SlogLoggerOption
func WithEventInitialCapacity(capacity int) SlogLoggerOption
```

`WithEventInitialCapacity` controls the pre-allocated attribute slice capacity
per event. Reduces allocations when you know the typical field count. Values
≤ 0 default to 5. Each `LogEvent` collects fields and emits them as a single
`slog.LogAttrs` call on `Send()`.

## Fluent API

All backends share the same interface. Call a level method to get a
`gas.LogEvent`, chain fields, finalize with `Send()`:

```go
logger.Info("request handled").
    Str("method", r.Method).
    Int("status", code).
    Duration("latency", elapsed).
    Send()
```

### Level Methods

| Method  | Returns        |
|---------|----------------|
| `Trace` | `gas.LogEvent` |
| `Debug` | `gas.LogEvent` |
| `Info`  | `gas.LogEvent` |
| `Warn`  | `gas.LogEvent` |
| `Error` | `gas.LogEvent` |

### Field Methods

Available on `gas.LogEvent`, `gas.LoggerContext`, and
`gas.MutableLoggerContext`:

| Method     | Signature                         | Description                |
|------------|-----------------------------------|----------------------------|
| `Str`      | `(key, val string)`               | String field               |
| `Int`      | `(key string, val int)`           | Integer field              |
| `Int64`    | `(key string, val int64)`         | 64-bit integer field       |
| `Float64`  | `(key string, val float64)`       | Float field                |
| `Bool`     | `(key string, val bool)`          | Boolean field              |
| `Err`      | `(key string, val error)`         | Error field                |
| `Duration` | `(key string, val time.Duration)` | Duration field             |
| `Any`      | `(key string, val any)`           | Arbitrary type field       |

### Other Methods

| Method  | On             | Description                                      |
|---------|----------------|--------------------------------------------------|
| `Flush` | Logger         | No-op for both backends (included for interface)  |
| `Send`  | LogEvent       | Emit the log event                                |

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

`With()` returns a `gas.LoggerContext`. Chain field methods, then call
`Logger()` to produce a new `gas.Logger`. The original logger is unchanged.

## Mutating Base Fields (SetBaseFields)

Unlike `With()` which branches into a new sub-logger, `SetBaseFields()`
accumulates fields and on `Apply()` mutates the originating logger in-place.

**When to use `SetBaseFields` vs `With`:**
- Use `With()` when you want a new, independent sub-logger (e.g. a per-request
  logger passed down to child calls).
- Use `SetBaseFields()` when middleware owns one logger instance and needs to
  stamp persistent fields onto it before the rest of the request runs.

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

Pass the constructor function to `gas.WithService` with
`gas.ServiceLifetimeScoped`. The DI container calls it once per request scope
to produce a fresh logger instance:

```go
zl := zerolog.New(os.Stdout).With().Timestamp().Logger()

app := gas.NewApp(
    gas.WithService[gas.Logger](
        gaslog.NewZeroLogLogger(gaslog.WithZeroLogInstance(&zl)),
        gas.ServiceLifetimeScoped, // per-request, not singleton
    ),
    // ...other services
)
```

With slog (zero external deps):

```go
sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))

app := gas.NewApp(
    gas.WithService[gas.Logger](
        gaslog.NewSlogLogger(
            gaslog.WithSlogInstance(sl),
            gaslog.WithEventInitialCapacity(5),
        ),
        gas.ServiceLifetimeScoped,
    ),
)
```

Using defaults (Zerolog uses `log.Logger`, Slog uses `slog.Default()`):

```go
app := gas.NewApp(
    gas.WithService[gas.Logger](gaslog.NewZeroLogLogger(), gas.ServiceLifetimeScoped),
)
```

## Complete Example

A full service with DI-wired logging, sub-loggers, and context propagation:

```go
package myservice

import (
    "context"
    "net/http"

    "github.com/gasmod/gas"
)

type Service struct {
    logger gas.Logger
    router *gas.Router
}

func New(logger gas.Logger, router *gas.Router) *Service {
    return &Service{logger: logger, router: router}
}

func (s *Service) Name() string  { return "myservice" }
func (s *Service) Close() error  { return nil }

func (s *Service) Init() error {
    s.logger.Info("service initialized").Str("name", s.Name()).Send()
    s.router.Handle(s.Name(), "GET", "/users/{id}", s.handleGetUser)
    return nil
}

func (s *Service) handleGetUser(ctx gas.Context, db gas.DatabaseProvider) error {
    // Create a sub-logger scoped to this request
    reqLogger := s.logger.With().
        Str("request_id", ctx.Header("X-Request-ID")).
        Str("user_id", ctx.Param("id")).
        Logger()

    // Propagate via context for downstream calls
    reqCtx := gas.WithLogger(ctx, reqLogger)

    user, err := s.fetchUser(reqCtx, db, ctx.Param("id"))
    if err != nil {
        reqLogger.Error("failed to fetch user").Err("error", err).Send()
        return err
    }

    reqLogger.Info("user fetched").Send()
    return ctx.JSON(http.StatusOK, user)
}

func (s *Service) fetchUser(ctx context.Context, db gas.DatabaseProvider, id string) (*User, error) {
    logger := gas.LoggerFromContext(ctx)
    logger.Debug("querying database").Str("user_id", id).Send()
    // ... database query
    return nil, nil
}

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

App wiring:

```go
package main

import (
    "os"

    "github.com/gasmod/gas"
    gaslog "github.com/gasmod/gas-log"
    "github.com/rs/zerolog"

    "myapp/myservice"
)

func main() {
    zl := zerolog.New(os.Stdout).With().Timestamp().Logger()

    app := gas.NewApp(
        gas.WithService[gas.Logger](
            gaslog.NewZeroLogLogger(gaslog.WithZeroLogInstance(&zl)),
            gas.ServiceLifetimeScoped, // per-request, not singleton
        ),
        gas.WithSingletonService[*myservice.Service](myservice.New),
    )
    app.Run()
}
```
