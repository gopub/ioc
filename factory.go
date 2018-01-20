package ioc

import (
	"fmt"
	"github.com/natande/gox"
	"reflect"
	"sync"
)

type Creator func(args ...interface{}) interface{}

type Factory interface {
	RegisterObject(obj interface{})
	RegisterInterface(ptrToInterface interface{}, obj interface{})
	RegisterCreator(name string, creator Creator, defaultArgs ...interface{})
	CreateObject(name string, args ...interface{}) (interface{}, error)
}

func NewFactory() Factory {
	f := &factoryImpl{
		nameToCreator: &sync.Map{},
	}

	return f
}

type creatorInfo struct {
	creator     Creator
	defaultArgs []interface{}
}

type factoryImpl struct {
	nameToCreator *sync.Map
}

func (f *factoryImpl) RegisterObject(obj interface{}) {
	var creator Creator = func(args ...interface{}) interface{} {
		p := reflect.New(reflect.TypeOf(obj))
		result := p.Interface()
		p = p.Elem()
		for p.Kind() == reflect.Ptr {
			p.Set(reflect.New(p.Type().Elem()))
			p = p.Elem()
		}
		return result
	}
	f.RegisterCreator(NameOfObject(obj), creator, nil)
}

func (f *factoryImpl) RegisterInterface(ptrToInterface interface{}, obj interface{}) {
	interfaceType := InterfaceOf(ptrToInterface)
	t := reflect.TypeOf(obj)
	if !t.Implements(interfaceType) {
		panic(t.Name() + " doesn't implement interface " + interfaceType.Name())
	}

	var creator Creator = func(args ...interface{}) interface{} {
		p := reflect.New(t)
		result := p.Interface()
		p = p.Elem()
		for p.Kind() == reflect.Ptr {
			p.Set(reflect.New(p.Type().Elem()))
			p = p.Elem()
		}
		return result
	}
	f.RegisterCreator(NameOfInterface(ptrToInterface), creator, nil)
}

func (f *factoryImpl) RegisterCreator(name string, creator Creator, defaultArgs ...interface{}) {
	if len(name) == 0 {
		panic("name is empty")
	}

	if creator == nil {
		panic("creator is nil")
	}

	_, ok := f.nameToCreator.Load(name)
	if ok {
		panic("duplicate creator for name: " + name)
		return
	}

	if len(defaultArgs) > 0 {
		t := reflect.TypeOf(creator)
		if len(defaultArgs) != t.NumIn() {
			panic("defaultArgs doesn't match creator's arguments")
		}

		for i := 0; i < t.NumIn(); i++ {
			if !reflect.TypeOf(defaultArgs[i]).AssignableTo(t.In(i)) {
				panic("defaultArgs doesn't match creator's arguments at: " + fmt.Sprint(i))
			}
		}
	}

	info := &creatorInfo{
		creator:     creator,
		defaultArgs: defaultArgs,
	}

	f.nameToCreator.Store(name, info)
}

func (f *factoryImpl) CreateObject(name string, args ...interface{}) (interface{}, error) {
	c, ok := f.nameToCreator.Load(name)
	if !ok {
		return nil, gox.ErrNotFound("name")
	}

	ci := c.(*creatorInfo)
	if len(args) > 0 {
		return ci.creator(args...), nil
	}

	return ci.creator(ci.defaultArgs...), nil
}
