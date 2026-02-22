package log

import (
	"time"

	"github.com/gasmod/gas"

	"github.com/rs/zerolog"
)

// ZeroLogLogger adapts a [zerolog.Logger] to the gas.Logger interface.
// It provides high-performance structured JSON logging with full support
// for all gas log levels including Trace.
type ZeroLogLogger struct {
	logger *zerolog.Logger
}

var _ gas.Logger = (*ZeroLogLogger)(nil)

// NewZeroLogLogger creates a gas.Logger backed by the given zerolog logger.
func NewZeroLogLogger(logger *zerolog.Logger) *ZeroLogLogger {
	return &ZeroLogLogger{logger: logger}
}

func (l *ZeroLogLogger) Trace(msg string) gas.LogEvent {
	return &ZeroLogLogEvent{evt: l.logger.Trace(), msg: msg}
}

func (l *ZeroLogLogger) Debug(msg string) gas.LogEvent {
	return &ZeroLogLogEvent{evt: l.logger.Debug(), msg: msg}
}

func (l *ZeroLogLogger) Info(msg string) gas.LogEvent {
	return &ZeroLogLogEvent{evt: l.logger.Info(), msg: msg}
}

func (l *ZeroLogLogger) Warn(msg string) gas.LogEvent {
	return &ZeroLogLogEvent{evt: l.logger.Warn(), msg: msg}
}

func (l *ZeroLogLogger) Error(msg string) gas.LogEvent {
	return &ZeroLogLogEvent{evt: l.logger.Error(), msg: msg}
}

func (l *ZeroLogLogger) Flush() {}

func (l *ZeroLogLogger) With() gas.LoggerContext {
	return &ZeroLogLoggerContext{ctx: l.logger.With()}
}

// ZeroLogLoggerContext wraps a [zerolog.Context] to implement gas.LoggerContext.
// Fields added via chained methods are baked into the sub-logger returned by Logger.
type ZeroLogLoggerContext struct {
	ctx zerolog.Context
}

var _ gas.LoggerContext = (*ZeroLogLoggerContext)(nil)

func (c *ZeroLogLoggerContext) Str(key, val string) gas.LoggerContext {
	c.ctx = c.ctx.Str(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Int(key string, val int) gas.LoggerContext {
	c.ctx = c.ctx.Int(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Int64(key string, val int64) gas.LoggerContext {
	c.ctx = c.ctx.Int64(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Float64(key string, val float64) gas.LoggerContext {
	c.ctx = c.ctx.Float64(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Bool(key string, val bool) gas.LoggerContext {
	c.ctx = c.ctx.Bool(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Err(key string, val error) gas.LoggerContext {
	c.ctx = c.ctx.AnErr(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Duration(key string, val time.Duration) gas.LoggerContext {
	c.ctx = c.ctx.Dur(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Any(key string, val any) gas.LoggerContext {
	c.ctx = c.ctx.Any(key, val)
	return c
}

func (c *ZeroLogLoggerContext) Logger() gas.Logger {
	l := c.ctx.Logger()
	return &ZeroLogLogger{logger: &l}
}

// ZeroLogLogEvent wraps a [zerolog.Event] to implement gas.LogEvent.
// Fields are chained fluently and the event is emitted when Send is called.
type ZeroLogLogEvent struct {
	evt *zerolog.Event
	msg string
}

var _ gas.LogEvent = (*ZeroLogLogEvent)(nil)

func (e *ZeroLogLogEvent) Str(key string, value string) gas.LogEvent {
	e.evt.Str(key, value)
	return e
}

func (e *ZeroLogLogEvent) Int(key string, value int) gas.LogEvent {
	e.evt.Int(key, value)
	return e
}

func (e *ZeroLogLogEvent) Int64(key string, value int64) gas.LogEvent {
	e.evt.Int64(key, value)
	return e
}

func (e *ZeroLogLogEvent) Float64(key string, value float64) gas.LogEvent {
	e.evt.Float64(key, value)
	return e
}

func (e *ZeroLogLogEvent) Bool(key string, value bool) gas.LogEvent {
	e.evt.Bool(key, value)
	return e
}

func (e *ZeroLogLogEvent) Err(key string, value error) gas.LogEvent {
	e.evt.AnErr(key, value)
	return e
}

func (e *ZeroLogLogEvent) Duration(key string, value time.Duration) gas.LogEvent {
	e.evt.Dur(key, value)
	return e
}

func (e *ZeroLogLogEvent) Any(key string, value any) gas.LogEvent {
	e.evt.Any(key, value)
	return e
}

func (e *ZeroLogLogEvent) Send() {
	e.evt.Msg(e.msg)
}
