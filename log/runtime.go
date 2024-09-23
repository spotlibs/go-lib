package log

import (
	"context"
	"sync"

	"github.com/spotlibs/go-lib/ctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	runOnce   sync.Once
	runZapLog *zap.Logger // runZapLog use zap log to log message, to benefit from its buffer.
)

type runLogger struct {
	reqId string
}

func (l runLogger) Info(m M) {
	if !isOff.Load() {
		// add request id
		m["request-id"] = l.reqId
		runZapLog.Info("", zap.Any("payload", m))
	}
}
func (l runLogger) Warning(m M) {
	if !isOff.Load() {
		// add request id
		m["request-id"] = l.reqId
		runZapLog.Warn("", zap.Any("payload", m))
	}
}
func (l runLogger) Error(m M) {
	if !isOff.Load() {
		// add request id
		m["request-id"] = l.reqId
		runZapLog.Error("", zap.Any("payload", m))
	}
}

// Runtime start RunLogger.
func Runtime(c ...context.Context) RunLogger {
	runOnce.Do(func() {
		// setup log writer
		runLogWriter := &writer{wr: setupLog("runtime")}
		// setup zap log
		core := zapcore.NewCore(getZapJsonEncoder(), zapcore.AddSync(runLogWriter), zapcore.InfoLevel)
		runZapLog = zap.New(core)
		// add to clean up task
		cleanupTasks = append(cleanupTasks, func() {
			_ = runZapLog.Sync()
			_ = runLogWriter.Close()
		})
	})

	// - Start embedding any necessary metadata from context

	var reqId string
	if len(c) > 0 {
		reqId = ctx.GetReqId(c[0])

		// add any other metadata here
	}

	// - End embedding any necessary metadata from context

	return runLogger{reqId: reqId}
}