package ioc

import (
	"reflect"
)

func InterfaceOf(ptrToInterface interface{}) reflect.Type {
	t := reflect.TypeOf(ptrToInterface)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Interface {
		panic("v is not a pointer to an interface. (*MyInterface)(nil)")
	}

	return t
}

func nameOfType(t reflect.Type) string {
	stars := ""
	if t == nil {
		return "nil"
	}

	for t.Kind() == reflect.Ptr {
		stars += "*"
		t = t.Elem()
	}

	if t.Kind() == reflect.Interface {
		stars = ""
	}
	return t.PkgPath() + "/" + stars + t.Name()
}

func NameOf(obj interface{}) string {
	name, ok := obj.(string)
	if ok {
		return name
	}
	return nameOfType(reflect.TypeOf(obj))
}
