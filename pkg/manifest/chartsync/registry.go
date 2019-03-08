package chartsync

//registry.go provides a mechanism for chart source handlers to self register
//and for consumers to find and use those handlers

import (
	"sync"
)

type Registration struct {
	Name    string
	New     regFunc
	Control Controller
}

type regFunc func(string, *ChartMeta, AccountTable) (Archiver, error)
type registrationList map[string]*Registration

type reg struct {
	sync.RWMutex
	list registrationList
}

//registry is in the non-exported global scope so source modules can self register
var registry *reg

func Register(r *Registration) {
	if registry == nil {
		registry = &reg{
			list: make(registrationList),
		}
	}
	registry.Lock()
	defer func() {
		registry.Unlock()
	}()
	registry.add(r.Name, r)
}

func (r *reg) add(name string, registration *Registration) {
	r.list[name] = registration
}

func (r *reg) Lookup(name string) (*Registration, bool) {
	if _, ok := r.list[name]; ok {
		return r.list[name], true
	}
	return nil, false
}

func (r *reg) AllControllers() []Controller {
	c := []Controller{}
	for _, v := range r.list {
		c = append(c, v.Control)
	}
	return c
}

func Reset() {
	for _, v := range registry.AllControllers() {
		v.Reset()
	}
}
