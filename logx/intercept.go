package logx

import (
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

type interceptCore struct {
	core zapcore.Core
	logQueue chan LogEntry
	droppedCount *int64
	service string
	instanceID string
}

func newInterceptCore(core zapcore.Core, queue chan LogEntry, droppedCount *int64, service, instanceID string) zapcore.Core {
	return &interceptCore{core: core, logQueue: queue, droppedCount: droppedCount, service: service, instanceID: instanceID}
}

func (c *interceptCore) Enabled(level zapcore.Level) bool {
	return c.core.Enabled(level)
}

func (c *interceptCore) With(fields []zapcore.Field) zapcore.Core {
	return newInterceptCore(c.core.With(fields), c.logQueue, c.droppedCount, c.service, c.instanceID)
}

func (c *interceptCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return c.core.Check(entry, ce).AddCore(entry, c)
	}
	return ce
}

func (c *interceptCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	select {
	case c.logQueue <- c.formatLogEntry(entry, fields):
	default:
		// drop or log if queue full
		atomic.AddInt64(c.droppedCount, 1)
	}
	return nil
}

func (c *interceptCore) Sync() error {
	return c.core.Sync()
}

func (c *interceptCore) formatLogEntry(entry zapcore.Entry, fields []zapcore.Field) LogEntry {
	return LogEntry{
		Level: entry.Level,
		Time: entry.Time,
		LoggerName: entry.LoggerName,
		Message: entry.Message,
		Caller: entry.Caller,
		Stack: entry.Stack,
		Fields: c.fieldsToMap(fields),
		Service: c.service,
		InstanceID: c.instanceID,
	}
}

func (c *interceptCore) fieldsToMap(fields []zapcore.Field) map[string]interface{} {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	return enc.Fields
}