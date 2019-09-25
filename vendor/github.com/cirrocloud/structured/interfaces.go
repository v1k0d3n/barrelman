package structured

import "github.com/cirrocloud/structured/log"

type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	WithFields(log.Fields) *log.Entry
}
