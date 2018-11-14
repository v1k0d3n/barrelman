package log

import (
	"bytes"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// Level type
type Level uint32

type Entry struct {
	Logger *Logger

	// Contains all the fields set by the user.
	Data Fields

	// Time at which the log entry was created
	Time time.Time

	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
	Level Level

	// Calling method, with package name
	Caller *runtime.Frame

	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
	Message string

	// When formatter is called in entry.log(), a Buffer may be set to entry
	Buffer *bytes.Buffer

	// err may contain a field formatting error
	err string
}

type Logger struct {
	Logger *logrus.Logger
}

type Fields map[string]interface{}

type keyvalser interface {
	Keyvals() []interface{}
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func (e *Entry) Info(args ...interface{}) {
	logrus.WithFields(logrus.Fields(e.getFields())).Info(args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func (e *Entry) Warn(args ...interface{}) {
	logrus.WithFields(logrus.Fields(e.getFields())).Warn(args...)
}

func (e *Entry) Debug(args ...interface{}) {
	logrus.WithFields(logrus.Fields(e.getFields())).Debug(args...)
}

func (e *Entry) Error(args ...interface{}) {
	logrus.WithFields(logrus.Fields(e.getFields())).Error(args...)
}

func Errorf(s string, v ...interface{}) {
	logrus.Errorf(s, v)
}

func WithField(k string, v interface{}) *Entry {
	return &Entry{Data: Fields{k: v}}
}

func (e *Entry) WithFields(f Fields) *Entry {
	if e.Data == nil {
		e.Data = f
		return e
	}
	for k, v := range f {
		e.Data[k] = v
	}
	return e
}

func WithFields(f Fields) *Entry {
	return &Entry{Data: f}
}

func (e *Entry) getFields() Fields {
	fields := Fields{}
	for k, v := range e.Data {
		if kv, ok := v.(keyvalser); ok {
			vals := kv.Keyvals()
			for i := 0; i < len(vals); i += 2 {
				fields[vals[i].(string)] = vals[i+1]
			}
		} else {
			fields[k] = v
		}
	}
	return fields
}
