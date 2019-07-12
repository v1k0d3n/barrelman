package errors

//Hyper experimental polyfill, not for public consumption

import (
	"fmt"

	"github.com/cirrocloud/structured/report"
	jerror "github.com/jjeffery/errors"
)

type Error = jerror.Error
type Context = jerror.Context

type Fields map[string]interface{}

type keyvalser interface {
	Keyvals() []interface{}
}

func New(s string) Error {
	return jerror.New(s)
}

func Rep(rep report.Reportables) Context {
	return WithReport(rep)
}

func WithReport(rep report.Reportables) Context {
	return WithFields(Fields(rep.DetailedReport()))
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

	msgNum := 0
	msgFill := func(value string, f Fields) Fields {
		for {
			msgString := fmt.Sprintf("msg-%v", msgNum)
			if _, exist := f[msgString]; !exist {
				f[msgString] = value
				return f
			}
			msgNum++
		}
	}
	if v, ok := err.(keyvalser); ok {
		fields := make(Fields)
		vals := v.(keyvalser).Keyvals()
		for i := 0; i < len(vals); i += 2 {
			fields[vals[i].(string)] = vals[i+1]
		}
		var cause string

		if m, ok := fields["cause"]; ok {
			cause = m.(string)
			delete(fields, "cause")
			err = New(cause)
		} else {
			if m, ok := fields["msg"]; ok {
				fields["cause"] = m
				delete(fields, "msg")
				err = New(m.(string))
			}
		}
		if m, ok := fields["msg"]; ok {
			fields = msgFill(m.(string), fields)
			delete(fields, "msg")
		}

		return WithFields(fields).Wrap(err, message...)
	}
	return jerror.Wrap(err, message...)
}

func Cause(err error) error {
	return jerror.Cause(err)
}
