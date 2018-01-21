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

func NameOfType(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() + "/" + t.Name()
}

func NameOfObject(obj interface{}) string {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() + "/" + t.Name()
}

func NameOfInterface(ptrToInterface interface{}) string {
	t := InterfaceOf(ptrToInterface)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() + "/" + t.Name()
}
