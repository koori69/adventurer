// Package adventurer @author K·J Create at 2019-01-09 11:10
package adventurer

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// ContextHook for log the call context
type ContextHook struct {
	Field  string
	Skip   int
	levels []logrus.Level
}

// DefaultLog default log
func DefaultLog() *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.AddHook(NewContextHook())
	return logger
}

// NewContextHook use to make an hook
func NewContextHook(levels ...logrus.Level) logrus.Hook {
	hook := ContextHook{
		Field:  "source",
		Skip:   5,
		levels: levels,
	}
	if len(hook.levels) == 0 {
		hook.levels = logrus.AllLevels
	}
	return &hook
}

// Levels implement levels
func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire implement fire
func (hook ContextHook) Fire(entry *logrus.Entry) error {
	entry.Data[hook.Field] = findCaller(hook.Skip)
	return nil
}

// findCaller caller file path
func findCaller(skip int) string {
	file := ""
	line := 0
	for i := 0; i < 10; i++ {
		file, line = getCaller(skip + i)
		if !strings.HasPrefix(file, "logrus") {
			break
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// getCaller get path
func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line
}
