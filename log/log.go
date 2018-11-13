package log

import (
	"github.com/sirupsen/logrus"
)

type Logger struct {
	Logger *logrus.Logger
}

type Fields map[string]interface{}

func Info(s string) {
	logrus.Info(s)
}

func Warn(s string) {
	logrus.Warn(s)
}

func Error(s string) {
	logrus.Error(s)
}

func Errorf(s string, v ...interface{}) {
	logrus.Error(s, v)
}

func WithField(k string, v interface{}) *logrus.Entry {
	return logrus.WithField(k, v)
}

func WithFields(f Fields) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(f))
}
