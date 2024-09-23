package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	actOnce   sync.Once
	actZapLog *zap.Logger // actZapLog use zap log to log message, to benefit from its buffer.
)

type actLogger struct{}

func (l actLogger) Info(m Map) {
	if !isOff.Load() {
		actZapLog.Info("", zap.Any("payload", m))
	}
}

// Activity start ActLogger.
func Activity(_ context.Context) ActLogger {
	actOnce.Do(func() {
		// setup log writer
		actLogWriter := &writer{wr: setupLog("activity")}
		// setup zap log
		core := zapcore.NewCore(getZapJsonEncoder(), zapcore.AddSync(actLogWriter), zapcore.InfoLevel)
		actZapLog = zap.New(core)
		// add to clean up task
		cleanupTasks = append(cleanupTasks, func() {
			_ = actZapLog.Sync()
			_ = actLogWriter.Close()
		})
	})

	return actLogger{}
}
