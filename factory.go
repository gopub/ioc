package ioc

import (
	"fmt"
	"github.com/gopub/types"
	"log"
	"reflect"
	"sync"
)

type Creator func(args ...interface{}) interface{}

type Factory interface {
	RegisterType(prototype interface{}) string
	RegisterCreator(name string, creator Creator, defaultArgs ...interface{})
	Create(name string, args ...interface{}) (interface{}, error)
	Contains(name string) bool
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

func (f *factoryImpl) RegisterType(prototype interface{}) string {
	var creator Creator = func(args ...interface{}) interface{} {
		v := reflect.New(reflect.TypeOf(prototype)).Elem()
		result := v
		for v.Kind() == reflect.Ptr {
			v.Set(reflect.New(v.Type().Elem()))
			v = v.Elem()
		}
		return result.Interface()
	}
	name := NameOf(prototype)
	f.RegisterCreator(name, creator)
	log.Printf("RegisterType:name=%s", name)
	return name
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

func (f *factoryImpl) Create(name string, args ...interface{}) (interface{}, error) {
	c, ok := f.nameToCreator.Load(name)
	if !ok {
		return nil, types.ErrNotFound("name")
	}

	ci := c.(*creatorInfo)
	if len(args) > 0 {
		return ci.creator(args...), nil
	}

	return ci.creator(ci.defaultArgs...), nil
}

func (f *factoryImpl) Contains(name string) bool {
	_, ok := f.nameToCreator.Load(name)
	return ok
}
