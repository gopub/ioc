package ioc

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

var errNoValue = errors.New("no value")

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

func ParseBool(i interface{}) (bool, error) {
	if i == nil {
		return false, errNoValue
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.String:
		str := strings.ToLower(v.String())
		if str == "true" {
			return true, nil
		}
		if str == "false" {
			return false, nil
		}
		return false, errNoValue
	default:
		b, err := ParseInt(i)
		if err == nil {
			return b != 0, nil
		}

		if v.Kind() == reflect.String {
			if str, ok := i.(string); ok {
				if str == "true" {
					return true, nil
				}

				if str == "false" {
					return false, nil
				}
			}
		}

		return false, err
	}
}

func ParseInt(i interface{}) (int64, error) {
	if i == nil {
		return 0, errNoValue
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		r := int64(v.Uint())
		return r, nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	case reflect.String:
		if num, ok := i.(json.Number); ok {
			n, e := num.Int64()
			if e != nil {
				var f float64
				f, e = num.Float64()
				n = int64(f)
			}
			return n, e
		}

		if n, err := strconv.ParseInt(v.String(), 0, 64); err == nil {
			return n, nil
		}

		if n, err := strconv.ParseFloat(v.String(), 64); err == nil {
			return int64(n), nil
		}
	}
	return 0, errNoValue
}

func ParseFloat(i interface{}) (float64, error) {
	if i == nil {
		return 0, errNoValue
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		if num, ok := i.(json.Number); ok {
			return num.Float64()
		}
		return strconv.ParseFloat(v.String(), 64)
	default:
		return 0, errNoValue
	}
}
