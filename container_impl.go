package ioc

import (
	"github.com/gopub/log"
	"reflect"
	"strings"
	"sync"
)

type registryInfo struct {
	name        string
	aliasList   []string
	isSingleton bool
	value       interface{}
	mu          sync.RWMutex
}

func (r *registryInfo) AppendAlias(alias string) {
	r.mu.Lock()
	r.aliasList = append(r.aliasList, alias)
	r.mu.Unlock()
}

type containerImpl struct {
	mu                  sync.RWMutex
	registries          []*registryInfo
	nameToRegistryIndex map[string]int

	factory Factory

	parent *containerImpl
}

func (c *containerImpl) getRegistry(name string) *registryInfo {
	c.mu.RLock()
	var r *registryInfo
	if i, ok := c.nameToRegistryIndex[name]; ok {
		r = c.registries[i]
	}
	c.mu.RUnlock()
	return r
}

func (c *containerImpl) addRegistry(r *registryInfo) {
	c.mu.Lock()
	index, ok := c.nameToRegistryIndex[r.name]
	if ok {
		log.Warnf("overwrite registry=%s", r.name)
	}

	if len(r.aliasList) == 0 {
		r.aliasList = []string{r.name}
	}

	if ok {
		c.registries[index] = r
	} else {
		c.registries = append(c.registries, r)
		c.nameToRegistryIndex[r.name] = len(c.registries) - 1
	}
	c.mu.Unlock()
}

func (c *containerImpl) RegisterValue(name string, value interface{}) bool {
	if len(name) == 0 {
		panic("name is empty")
	}

	if value == nil {
		panic("value is nil")
	}

	log.Infof("name=%s", name)
	r := &registryInfo{
		name:        name,
		isSingleton: true,
		value:       value,
	}
	c.addRegistry(r)
	return true
}

func (c *containerImpl) RegisterSingleton(prototype interface{}) string {
	name := c.factory.RegisterType(prototype)
	r := &registryInfo{
		name:        name,
		isSingleton: true,
	}
	c.addRegistry(r)
	return name
}

func (c *containerImpl) RegisterTransient(prototype interface{}) string {
	name := c.factory.RegisterType(prototype)
	r := &registryInfo{
		name: name,
	}
	c.addRegistry(r)
	return name
}

func (c *containerImpl) RegisterSingletonCreator(name string, creator Creator) bool {
	c.factory.RegisterCreator(name, creator)
	r := &registryInfo{
		name:        name,
		isSingleton: true,
	}
	c.addRegistry(r)
	return true
}

func (c *containerImpl) RegisterTransientCreator(name string, creator Creator) bool {
	c.factory.RegisterCreator(name, creator)
	r := &registryInfo{
		name: name,
	}
	c.addRegistry(r)
	return true
}

func (c *containerImpl) RegisterAliases(name string, aliases ...string) bool {
	logger := log.With("name", name)
	r := c.getRegistry(name)
	if r == nil {
		logger.Panic("no registry")
	}

	for _, alias := range aliases {
		if c.Contains(alias) {
			logger.Panicf("duplicated registry for alias:%s", alias)
		}
		r.AppendAlias(alias)
		c.mu.Lock()
		c.nameToRegistryIndex[alias] = c.nameToRegistryIndex[r.name]
		c.mu.Unlock()
		logger.Infof("registered alias=%s", alias)
	}

	return true
}

func (c *containerImpl) GetAliases(name string) []string {
	r := c.getRegistry(name)
	if r == nil {
		return nil
	}
	return r.aliasList
}

func (c *containerImpl) Contains(name string) bool {
	r := c.getRegistry(name)
	return r != nil
}

func (c *containerImpl) resolveByName(name string) interface{} {
	logger := log.With("name", name)
	r := c.getRegistry(name)
	if r == nil {
		logger.Error("no registry")
		return nil
	}

	if r.value != nil {
		logger.Infof("resolved %s", name)
		return r.value
	}

	v, err := c.factory.Create(r.name, nil)
	if err != nil {
		logger.Error(err)
		return nil
	}

	//cache value in registry before injecting in case of recursive dependency
	if r.isSingleton {
		r.value = v
	}

	c.Inject(v)

	if initializer, ok := v.(Initializer); ok {
		initializer.Init()
		logger.Infof("called %s.Init()", NameOf(v))
	}

	logger.Infof("resolved %s", name)
	return v
}

func (c *containerImpl) Resolve(prototype interface{}) interface{} {
	if s, ok := prototype.(string); ok {
		return c.resolveByName(s)
	}
	return c.resolveByName(NameOf(prototype))
}

func (c *containerImpl) Inject(ptrToObj interface{}) {
	log.Infof("start injecting %s", NameOf(ptrToObj))
	v := reflect.ValueOf(ptrToObj)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		log.Infof("name=%s, non-struct value", NameOf(ptrToObj))
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		if !f.CanSet() {
			continue
		}

		if name, ok := structField.Tag.Lookup("inject"); ok {
			name = strings.TrimSpace(name)
			if len(name) == 0 {
				name = nameOfType(f.Type())
			}

			obj := c.resolveByName(name)
			if obj == nil {
				log.Errorf("field=%s, name=%s, failed to resolve value", f.Type().Name(), nameOfType(t))
			} else {
				f.Set(reflect.ValueOf(obj))
			}
		}
	}

	log.Infof("injected %s", nameOfType(t))
}
