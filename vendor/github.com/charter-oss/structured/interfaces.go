package structured

import "github.com/charter-oss/structured/log"

type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	WithFields(log.Fields) *log.Entry
}
