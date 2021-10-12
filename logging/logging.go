package logging

import (
	"os"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var (
	globalLogger log.Logger
	loggerInit   sync.Once
)

func GlobalLogger() log.Logger {
	loggerInit.Do(func() {
		logger := NewLogger()
		logger = log.With(logger, "caller", log.DefaultCaller, "ts", log.DefaultTimestamp)
		globalLogger = logger
	})
	return globalLogger
}

func NewLogger() log.Logger {
	logger := log.NewJSONLogger(os.Stderr)

	option := level.AllowInfo()
	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "debug":
		option = level.AllowDebug()
	case "info":
		option = level.AllowInfo()
	case "warn", "warning":
		option = level.AllowWarn()
	case "error":
		option = level.AllowError()
	}
	return level.NewFilter(logger, option)
}
