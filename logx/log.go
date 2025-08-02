package logx

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A Logger provides fast, leveled, structured logging. All methods are safe
// for concurrent use.
// The Logger extends the zap.Logger with the ability to send all logs to a sing
type Logger struct {
	logger *zap.Logger

	cfg Config

	cancel func()
	logQueue chan LogEntry
	wg sync.WaitGroup
	droppedCount int64
	closed atomic.Bool
}

// New constructs a new Logger from the provided context, config, transport. The provided options are optional and for the internal zap.Logger.
//
// The Logger is designed fast and structured logging while also being able to 

// For sample code, see the package-level AdvancedConfiguration example.
func New(ctx context.Context, config Config, transport Transport, options ...zap.Option) *Logger {
	if !config.valid() {
		log.Fatal("logx.Config: missing service name or instance ID")
	}

	cfg := zap.NewProductionConfig()
	if config.State == Development {
		cfg = zap.NewDevelopmentConfig()
	}

	// Create encoder + output sink
	base, err := cfg.Build(options...)

	if err != nil {
		log.Fatal("could not build logging core: " + err.Error())
	}

	queue := make(chan LogEntry, 1000)

	var dropped int64
	intercept := newInterceptCore(base.Core(), queue, &dropped, config.Service, config.InstanceID, config.State)

	ctx, cancel := context.WithCancel(ctx)

	l := zap.New(intercept, options...)

	structuredL := l.With(zap.String("service", config.Service), zap.String("instance ID", config.InstanceID), zap.String("state", config.State.String()))

	lg := &Logger{
		logger: structuredL,
		cancel: cancel,
		logQueue: queue,
		droppedCount: dropped,
		cfg: config,
	}

	for range config.NumWorkers {
		lg.startLogWorker(ctx, transport)
	}

	return lg
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lg *Logger) Info(msg string, fields ...zapcore.Field) {
	lg.logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lg *Logger) Warn(msg string, fields ...zapcore.Field) {
	lg.logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lg *Logger) Error(msg string, fields ...zapcore.Field) {
	lg.logger.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (lg *Logger) Fatal(msg string, fields ...zapcore.Field) {
	lg.logger.Fatal(msg, fields...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lg *Logger) Debug(msg string, fields ...zapcore.Field) {
	lg.logger.Debug(msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (lg *Logger) DPanic(msg string, fields ...zapcore.Field) {
	lg.logger.DPanic(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (lg *Logger) Panic(msg string, fields ...zapcore.Field) {
	lg.logger.Panic(msg, fields...)
}

// Log logs a message at the specified level. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
// Any Fields that require  evaluation (such as Objects) are evaluated upon
// invocation of Log.
func (lg *Logger) Log(lvl zapcore.Level, msg string, fields ...zapcore.Field) {
	lg.logger.Log(lvl, msg, fields...)
}

// Level reports the minimum enabled level for this logger.
//
// For NopLoggers, this is [zapcore.InvalidLevel].
func (lg *Logger) Level() zapcore.Level  {
	return lg.logger.Level()
}

func (lg *Logger) With(fields ...zap.Field) *Logger {
	childLogger := lg.logger.With(fields...)
	return &Logger{
		logger:        childLogger,
		cfg:           lg.cfg,
		cancel:        lg.cancel,      // share same cancel function
		logQueue:      lg.logQueue,    // share same queue
		wg:            lg.wg,          // wait group not copied, since workers are shared
		droppedCount:  lg.droppedCount, // atomic counter is shared
		closed:        lg.closed,      // shared closed state
	}
}



// Close syncs and flushes any buffered log entries while also closing all log workers.
func (lg *Logger) Close() error {
	if lg.closed.CompareAndSwap(false, true) {
		lg.cancel()
		close(lg.logQueue)
		lg.wg.Wait()
	}
	return  lg.logger.Sync()
}

func (lg *Logger) startLogWorker(ctx context.Context, transport Transport) {
	lg.wg.Add(1)
	go func() {
		defer lg.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case log, ok := <-lg.logQueue:
				if !ok {
					return // channel closed
				}
				_ = transport.Send(log)
			}
		}
	}()
}

// DroppedCount returns the amount of log entries that were not sent with the transport
func (lg *Logger) DroppedCount() int64 {
	return atomic.LoadInt64(&lg.droppedCount)
}


// Config returns the underlying config
func (lg *Logger) Config() Config {
	return lg.cfg
}