package ioc

import (
	"fmt"
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
	if _, ok := c.nameToRegistryIndex[r.name]; ok {
		panic("duplicate name")
	}

	if len(r.aliasList) == 0 {
		r.aliasList = []string{r.name}
	}
	c.registries = append(c.registries, r)
	c.nameToRegistryIndex[r.name] = len(c.registries) - 1
	c.mu.Unlock()
}

func (c *containerImpl) RegisterValue(name string, value interface{}) bool {
	if len(name) == 0 {
		panic("name is empty")
	}

	if value == nil {
		panic("value is nil")
	}

	log.Infof("RegisterValue: name=%s", name)
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
	r := c.getRegistry(name)
	if r == nil {
		log.Panicf("name=%s, not found" + name)
	}

	for _, alias := range aliases {
		if c.Contains(alias) {
			log.Panicf("name=%s, duplicate registry", alias)
		}
		r.AppendAlias(alias)
		c.mu.Lock()
		c.nameToRegistryIndex[alias] = c.nameToRegistryIndex[r.name]
		c.mu.Unlock()
		log.Infof("name=%s, alias=%s", name, alias)
	}

	return true
}

func (c *containerImpl) GetAlias(name string) []string {
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

func (c *containerImpl) Resolve(name string) interface{} {
	r := c.getRegistry(name)
	if r == nil {
		log.Errorf("name=%s, no registry", name)
		return nil
	}

	if r.value != nil {
		return r.value
	}

	v, err := c.factory.Create(r.name, nil)
	if err != nil {
		log.Errorf("name=%s, error=%s", name, err)
		return nil
	}

	c.Inject(v)
	if r.isSingleton {
		r.value = v
	}

	if initializer, ok := v.(Initializer); ok {
		initializer.Init()
	}
	return v
}

func (c *containerImpl) Inject(ptrToObj interface{}) {
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

			obj := c.Resolve(name)
			if obj == nil {
				panic(fmt.Sprintf("Inject: failed to resolve value for field=%s, name=%s", f.Type().Name(), nameOfType(t)))
			}
			f.Set(reflect.ValueOf(obj))
		}
	}

	log.Infof("name=%s, injected", nameOfType(t))
}
