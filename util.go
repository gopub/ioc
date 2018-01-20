package ioc

import "reflect"

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

func NameOfObject(obj interface{}) string {
	t := reflect.TypeOf(obj)
	return t.PkgPath() + "/" + t.Name()
}

func NameOfInterface(ptrToInterface interface{}) string {
	t := InterfaceOf(ptrToInterface)
	return t.PkgPath() + "/" + t.Name()
}
