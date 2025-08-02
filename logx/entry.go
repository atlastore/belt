package logx

import (
	"encoding/json"
	"time"

	"go.uber.org/zap/zapcore"
)

type LogEntry struct {
	Level      zapcore.Level
	Time       time.Time
	LoggerName string
	Message    string
	Caller     zapcore.EntryCaller
	Stack      string
	Fields     map[string]any
	Service    string
	InstanceID string
	State      string
}

func (le *LogEntry) MarshalFields() ([]byte, error) {
	return json.Marshal(le.Fields)
}