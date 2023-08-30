package envygo

import (
	"fmt"
	"reflect"
)

type Env interface {
	Mock(env interface{}) func()
}

var registered []Env

func Publish(envs ...Env) {
	registered = append(registered, envs...)
}

func Mock(envs ...Env) func() {
	var funcs []func()

	for _, unknown := range envs {
		unknownType := reflect.ValueOf(unknown).Type()
		for _, known := range registered {
			if reflect.ValueOf(known).Type() == unknownType {
				funcs = append(funcs, mock(known, reflect.ValueOf(unknown).Elem().Interface()))
				unknownType = nil
			}
		}

		if unknownType != nil {
			panic(fmt.Sprintf("Attempt to mock unregistered type via %v", unknown))
		}
	}

	return func() {
		Unmock(funcs...)
	}
}

func Unmock(unmockers ...func()) {
	for _, function := range unmockers {
		defer function()
	}
}

func mock(old Env, new interface{}) func() {
	valueOfNew := reflect.ValueOf(new)
	typeOfNew := reflect.TypeOf(new)

	blankPtrValue := reflect.New(typeOfNew)
	valueOfBlank := blankPtrValue.Elem()

	oldPtrVal := reflect.New(reflect.TypeOf(old))
	oldPtrVal.Elem().Set(reflect.ValueOf(old))
	valueOfOld := oldPtrVal.Elem().Elem()

	includeUnset := typeOfNew != reflect.ValueOf(old).Elem().Type()

	for i := valueOfNew.NumField(); i > 0; {
		i--
		if typeOfNew.Field(i).IsExported() {
			newField := valueOfNew.Field(i)
			value := newField.Interface()
			if includeUnset || !isZero(reflect.ValueOf(value)) {
				name := typeOfNew.Field(i).Name
				oldField := valueOfOld.FieldByName(name)
				if oldField.CanSet() {
					blankField := valueOfBlank.Field(i)
					blankField.Set(oldField)
					oldField.Set(newField)
				} else {
					panic(fmt.Sprintf("Attempt to mock unregistered field via %s", name))
				}
			}
		}
	}

	blank := valueOfBlank.Interface()

	return func() {
		mock(old, blank)
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}

	// this takes care of the non-exported fields
	return v.IsValid() && v.IsZero()
}
