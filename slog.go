package log

import (
	"context"
	"log/slog"
	"time"

	"github.com/gasmod/gas"
)

// SlogLogger adapts a standard library [slog.Logger] to the gas.Logger interface.
// It requires no external dependencies beyond the Go standard library.
// Since slog has no Trace level, both Trace and Debug map to [slog.LevelDebug].
type SlogLogger struct {
	logger               *slog.Logger
	eventInitialCapacity int
}

var _ gas.Logger = (*SlogLogger)(nil)

// NewSlogLogger creates a gas.Logger backed by the given slog logger.
// eventInitialCapacity controls the pre-allocated capacity for attribute slices
// on each log event, reducing allocations for events with a known number of fields.
// Values <= 0 default to 5.
func NewSlogLogger(logger *slog.Logger, eventInitialCapacity int) *SlogLogger {
	if eventInitialCapacity <= 0 {
		eventInitialCapacity = 5
	}
	return &SlogLogger{logger: logger, eventInitialCapacity: eventInitialCapacity}
}

func (l *SlogLogger) Trace(msg string) gas.LogEvent {
	return &SlogLogEvent{
		logger: l.logger,
		lvl:    slog.LevelDebug,
		msg:    msg,
		attrs:  make([]slog.Attr, 0, l.eventInitialCapacity),
	}
}

func (l *SlogLogger) Debug(msg string) gas.LogEvent {
	return &SlogLogEvent{
		logger: l.logger,
		lvl:    slog.LevelDebug,
		msg:    msg,
		attrs:  make([]slog.Attr, 0, l.eventInitialCapacity),
	}
}

func (l *SlogLogger) Info(msg string) gas.LogEvent {
	return &SlogLogEvent{
		logger: l.logger,
		lvl:    slog.LevelInfo,
		msg:    msg,
		attrs:  make([]slog.Attr, 0, l.eventInitialCapacity),
	}
}

func (l *SlogLogger) Warn(msg string) gas.LogEvent {
	return &SlogLogEvent{
		logger: l.logger,
		lvl:    slog.LevelWarn,
		msg:    msg,
		attrs:  make([]slog.Attr, 0, l.eventInitialCapacity),
	}
}

func (l *SlogLogger) Error(msg string) gas.LogEvent {
	return &SlogLogEvent{
		logger: l.logger,
		lvl:    slog.LevelError,
		msg:    msg,
		attrs:  make([]slog.Attr, 0, l.eventInitialCapacity),
	}
}

func (l *SlogLogger) Flush() {}

func (l *SlogLogger) With() gas.LoggerContext {
	return &SlogLoggerContext{
		logger:               l.logger,
		attrs:                make([]any, 0, l.eventInitialCapacity),
		eventInitialCapacity: l.eventInitialCapacity,
	}
}

// SlogLoggerContext accumulates structured fields and produces a sub-logger
// with those fields baked in when Logger is called. Implements gas.LoggerContext.
type SlogLoggerContext struct {
	logger               *slog.Logger
	attrs                []any
	eventInitialCapacity int
}

var _ gas.LoggerContext = (*SlogLoggerContext)(nil)

func (c *SlogLoggerContext) Str(key, val string) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.String(key, val))
	return c
}

func (c *SlogLoggerContext) Int(key string, val int) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Int(key, val))
	return c
}

func (c *SlogLoggerContext) Int64(key string, val int64) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Int64(key, val))
	return c
}

func (c *SlogLoggerContext) Float64(key string, val float64) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Float64(key, val))
	return c
}

func (c *SlogLoggerContext) Bool(key string, val bool) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Bool(key, val))
	return c
}

func (c *SlogLoggerContext) Err(key string, val error) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Any(key, val))
	return c
}

func (c *SlogLoggerContext) Duration(key string, val time.Duration) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Duration(key, val))
	return c
}

func (c *SlogLoggerContext) Any(key string, val any) gas.LoggerContext {
	c.attrs = append(c.attrs, slog.Any(key, val))
	return c
}

func (c *SlogLoggerContext) Logger() gas.Logger {
	return &SlogLogger{
		logger:               c.logger.With(c.attrs...),
		eventInitialCapacity: c.eventInitialCapacity,
	}
}

// SlogLogEvent collects structured fields and emits them as a single log record
// via [slog.Logger.LogAttrs] when Send is called. Implements gas.LogEvent.
type SlogLogEvent struct {
	logger *slog.Logger
	msg    string
	attrs  []slog.Attr
	lvl    slog.Level
}

var _ gas.LogEvent = (*SlogLogEvent)(nil)

func (e *SlogLogEvent) Str(key string, value string) gas.LogEvent {
	e.attrs = append(e.attrs, slog.String(key, value))
	return e
}

func (e *SlogLogEvent) Int(key string, value int) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Int(key, value))
	return e
}

func (e *SlogLogEvent) Int64(key string, value int64) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Int64(key, value))
	return e
}

func (e *SlogLogEvent) Float64(key string, value float64) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Float64(key, value))
	return e
}

func (e *SlogLogEvent) Bool(key string, value bool) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Bool(key, value))
	return e
}

func (e *SlogLogEvent) Err(key string, value error) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Any(key, value))
	return e
}

func (e *SlogLogEvent) Duration(key string, value time.Duration) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Duration(key, value))
	return e
}

func (e *SlogLogEvent) Any(key string, value any) gas.LogEvent {
	e.attrs = append(e.attrs, slog.Any(key, value))
	return e
}

func (e *SlogLogEvent) Send() {
	e.logger.LogAttrs(context.Background(), e.lvl, e.msg, e.attrs...)
}
