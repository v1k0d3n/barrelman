package yamlpack

import (
	"fmt"
	"sync"
)

//Yp is a yamlpack instance
type Yp struct {
	sync.RWMutex
	Files    map[string][]*YamlSection
	Handlers map[string]func(string) error
}

//New returns a newly created *Yp
func New() *Yp {
	yp := &Yp{}
	yp.Handlers = make(map[string]func(string) error)
	yp.Files = make(map[string][]*YamlSection)
	return yp
}

//GetAllSections returns an array containing all yaml sections
func (yp *Yp) GetAllSections() []*YamlSection {
	yp.Lock()
	defer func() {
		yp.Unlock()
	}()
	ret := []*YamlSection{}
	for _, f := range yp.Files {
		for _, ys := range f {
			ret = append(ret, ys)
		}
	}
	return ret
}

//ListYamls returns a list of yaml section names as defined by metadata.name
func (yp *Yp) ListYamls() []string {
	ret := []string{}
	for _, ys := range yp.GetAllSections() {
		ret = append(ret, ys.Viper.Get("metadata.name").(string))
	}
	return ret
}

//RegisterHandler adds a handler to this instance
func (yp *Yp) RegisterHandler(s string, f func(string) error) error {
	yp.Lock()
	defer func() {
		yp.Unlock()
	}()
	if _, exists := yp.Handlers[s]; exists {
		return fmt.Errorf("handler \"%v\" already exists", s)
	}
	yp.Handlers[s] = f
	return nil
}

//DeregisterHandler removed a previously registered handler if it exists
func (yp *Yp) DeregisterHandler(s string) {
	yp.Lock()
	defer func() {
		yp.Unlock()
	}()
	if _, exists := yp.Handlers[s]; exists {
		delete(yp.Handlers, s)
	}
	return
}
