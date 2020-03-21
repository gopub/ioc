package ioc

import (
	"errors"
	"github.com/gopub/conv"
	"reflect"
	"strings"
	"sync"

	"github.com/gopub/environ"
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
		logger.Warnf("overwrite registry=%s", r.name)
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

	logger.Infof("name=%s", name)
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

func (c *containerImpl) RegisterAliases(origin interface{}, aliases ...interface{}) bool {
	name := NameOf(origin)
	r := c.getRegistry(name)
	if r == nil {
		logger.Panic("No registry")
	}

	for _, alias := range aliases {
		aliasName := NameOf(alias)
		if c.Contains(aliasName) {
			logger.Panicf("Duplicate registry for alias=%s", aliasName)
		}
		r.AppendAlias(aliasName)
		c.mu.Lock()
		c.nameToRegistryIndex[aliasName] = c.nameToRegistryIndex[r.name]
		c.mu.Unlock()
		logger.Infof("Origin=%s, alias=%s", aliasName, name)
	}
	return true
}

func (c *containerImpl) GetAliases(origin interface{}) []string {
	name, ok := origin.(string)
	if !ok {
		name = NameOf(origin)
	}
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

func (c *containerImpl) Resolve(prototype interface{}) interface{} {
	name := NameOf(prototype)
	r := c.getRegistry(name)
	if r == nil {
		err := errors.New("no registry")
		if AllowAbsent {
			logger.Errorf("No registry: name=%s", name)
			return nil
		}
		logger.Panic(err)
	}

	if r.value != nil {
		logger.Infof("Resolved: name=%s", name)
		return r.value
	}

	v, err := c.factory.Create(r.name, nil)
	if err != nil {
		if AllowAbsent {
			logger.Errorf("Failed to instantiate type=%s, err=%v", r.name, err)
		} else {
			logger.Panicf("Failed to instantiate type=%s, err=%v", r.name, err)
		}
		return nil
	}

	//cache value in registry before injecting in case of recursive dependency
	if r.isSingleton {
		r.value = v
	}

	if b, ok := v.(BeforeInjector); ok {
		logger.Debugf("%s.BeforeInject() begin", NameOf(v))
		b.BeforeInject()
		logger.Debugf("%s.BeforeInject() end", NameOf(v))
	}

	c.Inject(v)

	if a, ok := v.(AfterInjector); ok {
		logger.Debugf("%s.AfterInject() begin", NameOf(v))
		a.AfterInject()
		logger.Debugf("%s.AfterInject() end", NameOf(v))
	}

	if initializer, ok := v.(Initializer); ok {
		logger.Debugf("%s.Init() begin", NameOf(v))
		initializer.Init()
		logger.Debugf("%s.Init() end", NameOf(v))
	}

	logger.Infof("Instantiated: name=%s", r.name)
	return v
}

func (c *containerImpl) Inject(ptrToObj interface{}) {
	v := reflect.ValueOf(ptrToObj)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		err := errors.New("must be a pointer to struct value")
		if AllowAbsent {
			logger.Errorf("Failed to inject: %v", err)
		} else {
			logger.Panicf("Failed to inject: %v", err)
		}
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		if !f.CanSet() {
			continue
		}

		name, ok := structField.Tag.Lookup("inject")
		if !ok {
			continue
		}

		name = strings.TrimSpace(name)
		if len(name) == 0 {
			name = nameOfType(f.Type())
		}

		obj := c.Resolve(name)
		if obj != nil {
			f.Set(reflect.ValueOf(obj))
			continue
		}

		// Try to resolve value from environments
		envVal := environ.Get(strings.ToUpper(name))
		if envVal != nil {
			switch f.Kind() {
			case reflect.String:
				if s, err := conv.ToString(envVal); err == nil {
					f.SetString(s)
					continue
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if n, err := conv.ToInt64(envVal); err == nil {
					f.SetInt(n)
					continue
				}
			case reflect.Float32, reflect.Float64:
				if n, err := conv.ToFloat64(envVal); err == nil {
					f.SetFloat(n)
					continue
				}
			case reflect.Bool:
				if b, err := conv.ToBool(envVal); err == nil {
					f.SetBool(b)
					continue
				}
			}
		}
		logger.Errorf("Cannot resolve field=%s, of type=%s", name, nameOfType(t))
	}

	logger.Infof("Injected type=%s", NameOf(ptrToObj))
}
