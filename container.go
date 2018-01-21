package ioc

import (
	"github.com/natande/gox"
	"reflect"
	"sync"
	"fmt"
)

type Container interface {
	SetObject(name string, obj interface{})
	GetObject(name string) interface{}

	RegisterSingleton(prototype interface{})
	RegisterSingletonInterface(ptrToInterface interface{}, prototype interface{})

	RegisterTransient(prototype interface{})
	RegisterTransientInterface(ptrToInterface interface{}, prototype interface{})

	Invoke(f interface{}) ([]reflect.Value, error)
	Inject(obj interface{})

	Build()
}

var RootContainer = NewContainer()

func NewContainer() Container {
	return &containerImpl{
		nameToObject:   &sync.Map{},
		factory:        NewFactory(),
		singletonNames: gox.NewConcurrentSet(10),
	}
}

type containerImpl struct {
	nameToObject   *sync.Map
	singletonNames *gox.ConcurrentSet
	factory        Factory
}

func (c *containerImpl) SetObject(name string, obj interface{}) {
	if len(name) == 0 {
		panic("name is empty")
	}

	if obj == nil {
		panic("obj is nil")
	}
	c.nameToObject.Store(name, obj)
}

func (c *containerImpl) GetObject(name string) interface{} {
	v, ok := c.nameToObject.Load(name)
	if ok {
		return v
	}

	obj, err := c.factory.CreateObject(name, nil)
	if err != nil {
		return nil
	}

	if c.singletonNames.Contains(name) {
		c.nameToObject.Store(name, obj)
	} else {
		c.Inject(obj)
	}
	return obj
}

func (c *containerImpl) RegisterSingleton(prototype interface{}) {
	c.factory.Register(prototype)
	gox.LogInfo(prototype, NameOfObject(prototype))
	c.singletonNames.Add(NameOfObject(prototype))
}

func (c *containerImpl) RegisterSingletonInterface(ptrToInterface interface{}, prototype interface{}) {
	c.factory.RegisterInterface(ptrToInterface, prototype)
	gox.LogInfo(prototype, NameOfInterface(ptrToInterface))
	c.singletonNames.Add(NameOfInterface(ptrToInterface))
}

func (c *containerImpl) RegisterTransient(prototype interface{}) {
	c.factory.Register(prototype)
}

func (c *containerImpl) RegisterTransientInterface(ptrToInterface interface{}, prototype interface{}) {
	c.factory.RegisterInterface(ptrToInterface, prototype)
}

func (c *containerImpl) Invoke(f interface{}) ([]reflect.Value, error) {
	return nil, nil
}

func (c *containerImpl) Inject(ptrToObj interface{}) {
	v := reflect.ValueOf(ptrToObj)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		gox.LogError("Failed to inject into non-struct object: " + NameOfObject(ptrToObj))
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		if f.CanSet() && structField.Tag.Get("inject") != "" {
			ft := f.Type()
			obj := c.GetObject(NameOfType(ft))
			if obj == nil {
				panic("Failed to find object for type:" + NameOfType(ft))
			}
			f.Set(reflect.ValueOf(obj))
		}
	}

	gox.LogInfo("inject object for type: ", NameOfType(t))
}

func (c *containerImpl) Build() {
	names := c.singletonNames.Slice()
	for _, name := range names {
		obj := c.GetObject(name.(string))
		if obj != nil {
			gox.LogInfo("Initialized singleton object for", name)
		} else {
			gox.LogError("Failed to initialize singleton object for", name)
		}
	}

	initType := reflect.TypeOf((*Initializer)(nil)).Elem()
	for _, name := range names {
		obj := c.GetObject(name.(string))
		if obj != nil {
			c.Inject(obj)
			if reflect.TypeOf(obj).Implements(initType) {
				obj.(Initializer).Init()
			}
		}
	}
}
