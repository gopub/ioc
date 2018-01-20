package ioc

import (
	"github.com/natande/gox"
	"reflect"
	"sync"
)

type Container interface {
	SetObject(name string, obj interface{})
	GetObject(name string) interface{}

	RegisterSingletonObject(obj interface{})
	RegisterSingletonInterface(ptrToInterface interface{}, obj interface{})

	RegisterTransientObject(obj interface{})
	RegisterTransientInterface(ptrToInterface interface{}, obj interface{})

	Invoke(f interface{}) ([]reflect.Value, error)
	Inject(obj interface{})

	Build()
}

func NewContainer() Container {
	return &containerImpl{
		nameToObject: &sync.Map{},
		factory:      NewFactory(),
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
	}
	return obj
}

func (c *containerImpl) RegisterSingletonObject(obj interface{}) {
	c.factory.RegisterObject(obj)
	c.singletonNames.Add(NameOfObject(obj))
}

func (c *containerImpl) RegisterSingletonInterface(ptrToInterface interface{}, obj interface{}) {
	c.factory.RegisterInterface(ptrToInterface, obj)
	c.singletonNames.Add(NameOfObject(obj))
}

func (c *containerImpl) RegisterTransientObject(obj interface{}) {
	c.factory.RegisterObject(obj)
}

func (c *containerImpl) RegisterTransientInterface(ptrToInterface interface{}, obj interface{}) {
	c.factory.RegisterInterface(ptrToInterface, obj)
}

func (c *containerImpl) Invoke(f interface{}) ([]reflect.Value, error) {
	return nil, nil
}

func (c *containerImpl) Inject(obj interface{}) {

}

func (c *containerImpl) Build() {

}
