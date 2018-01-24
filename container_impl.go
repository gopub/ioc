package ioc

import (
	"fmt"
	"github.com/natande/gox"
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

	gox.LogInfo("register value for", name)
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

func (c *containerImpl) RegisterAlias(name string, aliasList ...string) bool {
	r := c.getRegistry(name)
	if r == nil {
		panic("not found: " + name)
	}

	for _, alias := range aliasList {
		if c.Contains(alias) {
			panic("duplicate registry for name: " + alias)
		}
		r.AppendAlias(alias)
		c.mu.Lock()
		c.nameToRegistryIndex[alias] = c.nameToRegistryIndex[r.name]
		c.mu.Unlock()
		gox.LogInfo(fmt.Sprintf("Name=%s, Alias=%s", name, alias))
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
		gox.LogWarn("not found:", name)
		return nil
	}

	if r.value != nil {
		return r.value
	}

	v, err := c.factory.Create(r.name, nil)
	if err != nil {
		gox.LogError(err, name)
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
		gox.LogError("Failed to inject into non-struct object: " + NameOfValue(ptrToObj))
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
				name = NameOfType(f.Type())
			}

			obj := c.Resolve(name)
			if obj == nil {
				panic("Failed to inject field:" + NameOfType(t) + "." + f.Type().Name())
			}
			f.Set(reflect.ValueOf(obj))
		}
	}

	gox.LogInfo("inject object for type: ", NameOfType(t))
}
