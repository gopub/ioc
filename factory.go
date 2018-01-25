package ioc

import (
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
