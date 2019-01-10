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
	log.With("prototype", name).Info("Succeeded")
	return name
}

func (f *factoryImpl) RegisterCreator(name string, creator Creator, defaultArgs ...interface{}) {
	logger := log.With("name", name)
	if len(name) == 0 {
		logger.Panic("name is empty")
	}

	if creator == nil {
		logger.Panic("creator is nil")
	}

	_, ok := f.nameToCreator.Load(name)
	if ok {
		logger.Warnf("overwrote creator")
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
	logger.Info("Succeeded")
}

func (f *factoryImpl) Create(name string, args ...interface{}) (interface{}, error) {
	logger := log.With("name", name)
	c, ok := f.nameToCreator.Load(name)
	if !ok {
		err := errors.New("no creator")
		if AllowAbsent {
			logger.Error(err)
			return nil, err
		}
		logger.Panic(err)
	}

	ci := c.(*creatorInfo)
	var result interface{}
	if len(args) > 0 {
		result = ci.creator(args...)
	} else {
		result = ci.creator(ci.defaultArgs...)
	}

	logger.Info("Succeeded")
	return result, nil
}

func (f *factoryImpl) Contains(name string) bool {
	_, ok := f.nameToCreator.Load(name)
	return ok
}
