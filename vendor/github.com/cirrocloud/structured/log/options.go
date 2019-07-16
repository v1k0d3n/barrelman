package log

import (
	"github.com/sirupsen/logrus"
)

func OptSetLevel(arg string) func(*Logger) error {
	var level logrus.Level
	switch arg {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	}
	return func(logger *Logger) error {
		logger.Handler.SetLevel(level)
		return nil
	}
}

func OptSetJSON() func(*Logger) error {
	return func(logger *Logger) error {
		logger.Handler.SetFormatter(&logrus.JSONFormatter{})
		return nil
	}
}

func OptForceColor() func(*Logger) error {
	return func(logger *Logger) error {
		logger.Handler.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
		return nil
	}
}
