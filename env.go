package envygo

import (
	"fmt"
	"reflect"
)

var registered []interface{}

// Introduce introduces (akin to registers) environments which are
// likely to be mocked in future using MockMany function
func Introduce(envs ...interface{}) {
	registered = append(registered, envs...)
}

// MockMany is a convenience method for mocking many environments
// at once that were previously introduced using Introduce function.
// Each of the envs is matched to a previously introduced environment
// based on type commonality and the introduced environment is
// mocked using the passed environment.
func MockMany(envs ...interface{}) func() {
	var funcs []func()

	for _, unknown := range envs {
		unknownType := reflect.ValueOf(unknown).Type()
		for _, known := range registered {
			if reflect.ValueOf(known).Type() == unknownType {
				funcs = append(funcs, MockOne(known, reflect.ValueOf(unknown).Elem().Interface()))
				unknownType = nil
			}
		}

		if unknownType != nil {
			panic(fmt.Sprintf("Attempt to mock unregistered type via %v", unknown))
		}
	}

	return func() {
		for _, function := range funcs {
			defer function()
		}
	}
}

// MockOne mocks the old environment using the values in the new environment
// If the type for old environment is identical to the type of the new environment
// then any attribute with value identical to default value for its type is
// not mocked. But if types are different then the attribute is mocked regardless
// of its value
func MockOne(old interface{}, new interface{}) func() {
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
		MockOne(old, blank)
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
