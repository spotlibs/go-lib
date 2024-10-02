package log

import (
	"context"
	"sync"

	"github.com/spotlibs/go-lib/ctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	actOnce   sync.Once
	actZapLog *zap.Logger // actZapLog use zap log to log message, to benefit from its buffer.
)

type actLogger struct {
	trcId      string
	identifier string
}

func (l actLogger) Info(m Map) {
	if !isOff.Load() {
		m["trace-id"] = l.trcId
		m["identifier"] = l.identifier
		actZapLog.Info("", zap.Any("payload", m))
	}
}

// Activity start ActLogger.
func Activity(c context.Context) ActLogger {
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

	// - Start embedding any necessary metadata from context

	trcId := ctx.GetReqId(c)
	id := ctx.Get(c).UrlPath
	// add any other metadata here

	// - End embedding any necessary metadata from context

	return actLogger{trcId: trcId, identifier: id}
}
