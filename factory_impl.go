package ioc

import (
	"errors"
	"fmt"
	"github.com/gopub/log"
	"reflect"
	"sync"
)

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
	log.Infof("name=%s", name)
	return name
}

func (f *factoryImpl) RegisterCreator(name string, creator Creator, defaultArgs ...interface{}) {
	if len(name) == 0 {
		log.Panic("name is empty")
	}

	if creator == nil {
		log.Panic("creator is nil")
	}

	_, ok := f.nameToCreator.Load(name)
	if ok {
		log.Warnf("overwrite creator of name=%s", name)
		return
	}

	if len(defaultArgs) > 0 {
		fmt.Println(defaultArgs...)
		t := reflect.TypeOf(creator)
		if len(defaultArgs) != t.NumIn() {
			log.Panicf("name=%s, num=%d, argument number doesn't match", name, len(defaultArgs))
		}

		for i := 0; i < t.NumIn(); i++ {
			if !reflect.TypeOf(defaultArgs[i]).AssignableTo(t.In(i)) {
				log.Panicf("name=%s, index=%d, type=%v, requiredType=%v, type doesn't match",
					name, i, reflect.TypeOf(defaultArgs[i]), t.In(i))
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
		return nil, errors.New("no creator for name:" + name)
	}

	defer func() {
		log.Infof("Created %s", name)
	}()

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
