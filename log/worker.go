package log

import (
	"context"
	"sync"

	"github.com/spotlibs/go-lib/ctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	wrkOnce   sync.Once
	wrkZapLog *zap.Logger // wrkZapLog use zap log to log message, to benefit from its buffer.
)

type wrkLogger struct {
	trcId      string
	identifier string
}

func (l wrkLogger) Info(m Map) {
	if !isOff.Load() {
		// add request id
		m["trace-id"] = l.trcId
		m["identifier"] = l.identifier
		wrkZapLog.Info("", zap.Any("payload", m))
	}
}

// Worker start WorkLogger.
func Worker(c context.Context) WorkLogger {
	// prevent panic
	if c == nil {
		c = context.Background()
	}

	wrkOnce.Do(func() {
		// setup log writer
		wrkLogWriter := &writer{wr: setupLog("worker")}
		// setup zap log
		core := zapcore.NewCore(getZapJsonEncoder(), zapcore.AddSync(wrkLogWriter), zapcore.InfoLevel)
		wrkZapLog = zap.New(core)
		// add to clean up task
		cleanupTasks = append(cleanupTasks, func() {
			_ = wrkZapLog.Sync()
			_ = wrkLogWriter.Close()
		})
	})

	// - Start embedding any necessary metadata from context

	trcId := ctx.GetReqId(c)
	id := ctx.Get(c).SignaturePath
	// add any other metadata here

	// - End embedding any necessary metadata from context

	return wrkLogger{trcId: trcId, identifier: id}
}
