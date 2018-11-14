package errors

import (
	jerror "github.com/jjeffery/errors"
)

type Error = jerror.Error
type Context = jerror.Context

type Fields map[string]interface{}

func New(s string) Error {
	return jerror.New(s)
}

func With(keyvals ...interface{}) Context {
	return jerror.With(keyvals)
}

func WithFields(fields Fields) Context {
	keyvals := []interface{}{}
	for k, v := range fields {
		keyvals = append(keyvals, k, v)
	}
	return jerror.With(keyvals...)
}

func Wrap(err error, message ...string) error {
	return jerror.Wrap(err, message...)
}

func Cause(err error) error {
	return jerror.Cause(err)
}
