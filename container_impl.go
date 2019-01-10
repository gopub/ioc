package ioc

import (
	"errors"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"os"
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

func (c *containerImpl) RegisterAliases(origin interface{}, aliases ...interface{}) bool {
	name := NameOf(origin)
	logger := log.With("origin", name)
	r := c.getRegistry(name)
	if r == nil {
		logger.Panic("No registry")
	}

	for _, alias := range aliases {
		aliasName := NameOf(alias)
		if c.Contains(aliasName) {
			logger.Panicf("Duplicate registry for alias:%s", aliasName)
		}
		r.AppendAlias(aliasName)
		c.mu.Lock()
		c.nameToRegistryIndex[aliasName] = c.nameToRegistryIndex[r.name]
		c.mu.Unlock()
		logger.Infof("Registered alias:%s", aliasName)
	}

	logger.Info("Succeeded")
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
	logger := log.With("name", name)
	r := c.getRegistry(name)
	if r == nil {
		err := errors.New("no registry")
		if AllowAbsent {
			logger.Error(err)
			return nil
		}
		logger.Panic(err)
	}

	if r.value != nil {
		logger.Info("Succeeded")
		return r.value
	}

	v, err := c.factory.Create(r.name, nil)
	if err != nil {
		if AllowAbsent {
			logger.Error(err)
		} else {
			logger.Panic(err)
		}
		return nil
	}

	//cache value in registry before injecting in case of recursive dependency
	if r.isSingleton {
		r.value = v
	}

	c.Inject(v)

	if initializer, ok := v.(Initializer); ok {
		logger.Infof("Executing %s.Init()", NameOf(v))
		initializer.Init()
		logger.Infof("Finished %s.Init()", NameOf(v))
	}

	logger.Info("Succeeded")
	return v
}

func (c *containerImpl) Inject(ptrToObj interface{}) {
	logger := log.With("ptrToObj", NameOf(ptrToObj))
	v := reflect.ValueOf(ptrToObj)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		err := errors.New("must be a pointer to struct value")
		if AllowAbsent {
			logger.Error(err)
		} else {
			logger.Panic(err)
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
		envVal := os.Getenv(strings.ToUpper(name))
		if len(envVal) > 0 {
			switch f.Kind() {
			case reflect.String:
				f.SetString(envVal)
				continue
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				intVal, err := types.ParseInt(envVal)
				if err == nil {
					f.SetInt(intVal)
					continue
				}
			case reflect.Float32, reflect.Float64:
				floatVal, err := types.ParseFloat(envVal)
				if err == nil {
					f.SetFloat(floatVal)
					continue
				}
			case reflect.Bool:
				boolVal, err := types.ParseBool(envVal)
				if err == nil {
					f.SetBool(boolVal)
					continue
				}
			}
		}
		logger.Errorf("Failed to resolve field=%s, name=%s", f.Type().Name(), nameOfType(t))
	}

	logger.Infof("Succeeded")
}
