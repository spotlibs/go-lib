package log

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logDir = "./storage/logs"

var (
	cleanupTasks []func()
	output       io.WriteCloser
)

// SetOutput set the output of the logs to given writer. By default, will print
// to the predefined logs directory ./storage/logs/.
func SetOutput(w io.WriteCloser) {
	output = w
}

// Cleanup do clean up for any writers and SHOULD be called before program
// exit to make sure no lingering resource still opened while program exit and
// will lead to resource leakage.
func Cleanup() {
	for i := range cleanupTasks {
		// execute all the registered cleanup task
		cleanupTasks[i]()
	}
}

type writer struct {
	wr io.WriteCloser
}

func (w *writer) Write(p []byte) (n int, err error) {
	// format incoming message from zap then pass it to the real writer
	var m, mm Map
	_ = sonic.ConfigStd.Unmarshal(p, &m)
	// try grab level from the payload of zap log message
	var lvl string
	if v, ok := m["level"]; ok {
		lvl = v.(string)
	}
	// try grab the time
	var t string
	if v, ok := m["time"]; ok {
		t = v.(string)
	}
	// grab only the payload
	mm = m["payload"].(map[string]any)

	return w.wr.Write([]byte(formatMsg(t, strings.ToUpper(lvl), mm)))
}
func (w *writer) Close() error { return w.wr.Close() }

// ActLogger activity logger and can be accessed with Activity.
//
// This logger may not often be used by most developers since this may only be
// used inside middleware.
type ActLogger interface {
	// Info log given payload in INFO level.
	Info(Map)
}

// WorkLogger worker logger and can be accessed with Worker.
//
// This logger may only be used in worker or job process.
type WorkLogger interface {
	// Info log given payload in INFO level.
	Info(Map)
}

// RunLogger runtime logger and can be accessed with Runtime.
//
// This logger may be the most widely used by developers since this can be used
// in almost anywhere in the system.
type RunLogger interface {
	// Info log given payload in INFO level.
	Info(Map)
	// Warning log given payload in WARNING level.
	Warning(Map)
	// Error log given payload in ERROR level.
	Error(Map)
}

// Map construct map data structure with string as the key type.
type Map map[string]any

// setupLog do setup and return io.WriteCloser that ready to use as target
// output of logs.
func setupLog(logType string) io.WriteCloser {
	if output != nil {
		return output
	}
	checkAndCreateDir(logDir)
	return createOrOpenFile(logDir + "/" + logType + ".log")
}

// checkAndCreateDir make sure the given dir already exist otherwise will try
// to create it with Mkdir in 755 mode.
func checkAndCreateDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Errorf("failed in creating the logs directory: %e", err))
		}
	}
}

// createOrOpenFile create or open file if already exist for the given file
// path with these mode:
//
//   - os.O_APPEND: open in append mode, so we don't overwrite the content.
//   - os.O_CREATE: create the file if it doesn't exist.
//   - os.O_WRONLY: open for writing only.
func createOrOpenFile(logPath string) io.WriteCloser {
	fl, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open log file: %e", err))
	}
	return fl
}

// formatMsg format log message based on the level string and M as the payload.
func formatMsg(t, lvl string, m Map) string {
	payload, _ := sonic.ConfigStd.MarshalToString(m)
	return fmt.Sprintf("[%s] ::%s.%s.%s:: %s\n", t, os.Getenv("APP_NAME"), os.Getenv("APP_ENV"), lvl, payload)
}

// getZapJsonEncoder return common setting for zap json encoder.
func getZapJsonEncoder() zapcore.Encoder {
	jsonEnc := zap.NewProductionEncoderConfig()
	jsonEnc.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	jsonEnc.TimeKey = "time"
	return zapcore.NewJSONEncoder(jsonEnc)
}
