package log

import (
	"time"

	"github.com/gasmod/gas"
)

var nopLogger = &NoOpLogger{}

// NoOpLogger is a gas.Logger implementation that silently discards all log output.
// It uses a singleton pattern — all instances share the same underlying value,
// resulting in zero allocations per log call.
type NoOpLogger struct{}

var _ gas.Logger = (*NoOpLogger)(nil)

// NoOpLoggerCtor defines a constructor function that returns a nop-logger implementing the gas.Logger interface.
type NoOpLoggerCtor func() *NoOpLogger

// NewNoOpLogger returns a NoOpLoggerCtor that constructs a nop-logger implementing the gas.Logger interface.
func NewNoOpLogger() NoOpLoggerCtor {
	return func() *NoOpLogger {
		return nopLogger
	}
}

func (l *NoOpLogger) Trace(string) gas.LogEvent { return nopLogEvent }
func (l *NoOpLogger) Debug(string) gas.LogEvent { return nopLogEvent }
func (l *NoOpLogger) Info(string) gas.LogEvent  { return nopLogEvent }
func (l *NoOpLogger) Warn(string) gas.LogEvent  { return nopLogEvent }
func (l *NoOpLogger) Error(string) gas.LogEvent { return nopLogEvent }
func (l *NoOpLogger) Flush()                    {}

func (l *NoOpLogger) With() gas.LoggerContext {
	return nopLoggerContext
}

func (l *NoOpLogger) SetBaseFields() gas.MutableLoggerContext {
	return nopMutableLoggerContext
}

var nopLoggerContext = &NoOpLoggerContext{}

// NoOpLoggerContext is a gas.LoggerContext that discards all fields.
// All methods return the receiver for chaining and Logger returns a [NoOpLogger].
type NoOpLoggerContext struct{}

var _ gas.LoggerContext = (*NoOpLoggerContext)(nil)

func (c *NoOpLoggerContext) Str(string, string) gas.LoggerContext             { return c }
func (c *NoOpLoggerContext) Int(string, int) gas.LoggerContext                { return c }
func (c *NoOpLoggerContext) Int64(string, int64) gas.LoggerContext            { return c }
func (c *NoOpLoggerContext) Float64(string, float64) gas.LoggerContext        { return c }
func (c *NoOpLoggerContext) Bool(string, bool) gas.LoggerContext              { return c }
func (c *NoOpLoggerContext) Err(string, error) gas.LoggerContext              { return c }
func (c *NoOpLoggerContext) Duration(string, time.Duration) gas.LoggerContext { return c }
func (c *NoOpLoggerContext) Any(string, any) gas.LoggerContext                { return c }
func (c *NoOpLoggerContext) Logger() gas.Logger                               { return nopLogger }

var nopMutableLoggerContext = &NoOpMutableLoggerContext{}

// NoOpMutableLoggerContext is a gas.MutableLoggerContext that discards all fields.
// All methods return the receiver for chaining and Apply is a no-op.
type NoOpMutableLoggerContext struct{}

var _ gas.MutableLoggerContext = (*NoOpMutableLoggerContext)(nil)

func (c *NoOpMutableLoggerContext) Str(string, string) gas.MutableLoggerContext             { return c }
func (c *NoOpMutableLoggerContext) Int(string, int) gas.MutableLoggerContext                { return c }
func (c *NoOpMutableLoggerContext) Int64(string, int64) gas.MutableLoggerContext            { return c }
func (c *NoOpMutableLoggerContext) Float64(string, float64) gas.MutableLoggerContext        { return c }
func (c *NoOpMutableLoggerContext) Bool(string, bool) gas.MutableLoggerContext              { return c }
func (c *NoOpMutableLoggerContext) Err(string, error) gas.MutableLoggerContext              { return c }
func (c *NoOpMutableLoggerContext) Duration(string, time.Duration) gas.MutableLoggerContext { return c }
func (c *NoOpMutableLoggerContext) Any(string, any) gas.MutableLoggerContext                { return c }
func (c *NoOpMutableLoggerContext) Apply()                                                  {}

var nopLogEvent = &NoOpLogEvent{}

// NoOpLogEvent is a gas.LogEvent that discards all fields and performs no
// action on Send. All methods return the receiver for chaining.
type NoOpLogEvent struct{}

var _ gas.LogEvent = (*NoOpLogEvent)(nil)

func (e *NoOpLogEvent) Str(string, string) gas.LogEvent             { return e }
func (e *NoOpLogEvent) Int(string, int) gas.LogEvent                { return e }
func (e *NoOpLogEvent) Int64(string, int64) gas.LogEvent            { return e }
func (e *NoOpLogEvent) Float64(string, float64) gas.LogEvent        { return e }
func (e *NoOpLogEvent) Bool(string, bool) gas.LogEvent              { return e }
func (e *NoOpLogEvent) Err(string, error) gas.LogEvent              { return e }
func (e *NoOpLogEvent) Duration(string, time.Duration) gas.LogEvent { return e }
func (e *NoOpLogEvent) Any(string, any) gas.LogEvent                { return e }
func (e *NoOpLogEvent) Send()                                       {}
