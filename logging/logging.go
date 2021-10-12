package logging

import (
	"os"
	"sync"

	"github.com/go-kit/log"
)

var (
	globalLogger log.Logger
	loggerInit   sync.Once
)

func GlobalLogger() log.Logger {
	loggerInit.Do(func() {
		globalLogger = NewLogger()
	})
	return globalLogger
}

func NewLogger() log.Logger {
	logger := log.NewJSONLogger(os.Stderr)
	logger = log.With(logger, "caller", log.DefaultCaller, "ts", log.DefaultTimestamp)
	return logger
}
