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
	Handler *logrus.Logger
}

type Fields map[string]interface{}

type errorer interface {
	Error() string
}

func New(funcArgs ...func(*Logger) error) *Logger {
	logger := &Logger{
		Handler: logrus.New(),
	}
	for _, f := range funcArgs {
		if err := f(logger); err != nil {
			//Returning an error here becomes too much to handle down the stack
			panic(err)
		}
	}
	return logger
}

func (l *Logger) Debug(args ...interface{}) {
	l.Handler.Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Handler.Info(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.Handler.Warn(args...)
}

func (l *Logger) Error(args ...interface{}) {
	msg, fields := getFields(args)
	l.Handler.WithFields(logrus.Fields(fields)).Error(msg)
}

func (l *Logger) WithFields(f Fields) *Entry {
	return &Entry{
		Logger: l,
		Data:   f,
	}
}

// *************
func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Error(args ...interface{}) {
	msg, fields := getFields(args)
	logrus.WithFields(logrus.Fields(fields)).Error(msg)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func (e *Entry) Info(args ...interface{}) {
	e.Logger.Handler.WithFields(logrus.Fields(e.getFields())).Info(args...)
}

func (e *Entry) Warn(args ...interface{}) {
	e.Logger.Handler.WithFields(logrus.Fields(e.getFields())).Warn(args...)
}

func (e *Entry) Debug(args ...interface{}) {
	e.Logger.Handler.WithFields(logrus.Fields(e.getFields())).Debug(args...)
}

func (e *Entry) Error(args ...interface{}) {
	e.Logger.Handler.WithFields(logrus.Fields(e.getFields())).Error(args...)
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
	return &Entry{
		Logger: New(),
		Data:   f,
	}
}

type keyvalser interface {
	Keyvals() []interface{}
}

func recurseFields(args ...interface{}) Fields {
	fields := Fields{}
	for _, v := range args {
		switch v.(type) {
		case keyvalser:
			vals := v.(keyvalser).Keyvals()
			for i := 0; i < len(vals); i++ {
				switch vals[i].(type) {
				case string:
					fields[vals[i].(string)] = vals[i+1]
					i++
					continue
				default:
					for k, v := range recurseFields(vals[i]) {
						fields[k] = v
					}
				}
			}
		case []interface{}:
			for k, v := range recurseFields(v.([]interface{})...) {
				fields[k] = v
			}
		}
	}
	return fields
}

func getFields(args ...interface{}) (interface{}, Fields) {
	fields := Fields{}
	var msg interface{}
	for _, v := range args {
		switch v.(type) {
		case keyvalser:
			vals := v.(keyvalser).Keyvals()
			for i := 0; i < len(vals); i += 2 {
				fields[vals[i].(string)] = vals[i+1]
			}
		case []interface{}:
			for k, v := range recurseFields(v) {
				fields[k] = v
			}
		}
	}
	if v, ok := fields["msg"]; ok {
		msg = v
		delete(fields, "msg")
	}
	return msg, fields
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
