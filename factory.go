package ioc

import (
	"fmt"
	"github.com/natande/gox"
	"reflect"
	"sync"
)

type Creator func(args ...interface{}) interface{}

type Factory interface {
	Register(prototype interface{})
	RegisterInterface(ptrToInterface interface{}, prototype interface{})
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

func (f *factoryImpl) Register(prototype interface{}) {
	var creator Creator = func(args ...interface{}) interface{} {
		v := reflect.New(reflect.TypeOf(prototype)).Elem()
		result := v
		for v.Kind() == reflect.Ptr {
			v.Set(reflect.New(v.Type().Elem()))
			v = v.Elem()
		}
		return result.Interface()
	}
	f.RegisterCreator(NameOfObject(prototype), creator)
}

func (f *factoryImpl) RegisterInterface(ptrToInterface interface{}, prototype interface{}) {
	interfaceType := InterfaceOf(ptrToInterface)
	t := reflect.TypeOf(prototype)
	if !t.Implements(interfaceType) {
		panic(t.Name() + " doesn't implement interface " + interfaceType.Name())
	}

	var creator Creator = func(args ...interface{}) interface{} {
		v := reflect.New(reflect.TypeOf(prototype)).Elem()
		result := v
		for v.Kind() == reflect.Ptr {
			v.Set(reflect.New(v.Type().Elem()))
			v = v.Elem()
		}
		return result.Interface()
	}
	f.RegisterCreator(NameOfInterface(ptrToInterface), creator)
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
		fmt.Println(defaultArgs...)
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
