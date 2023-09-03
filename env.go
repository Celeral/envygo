package envygo

import (
	"fmt"
	"reflect"
	"unsafe"
)

var registered []interface{}

// Introduce introduces (akin to registers) environments which are
// likely to be mocked in future using MockMany function
// Deprecated Introduce is being considered for deletion to
// reduce the footprint of the API
func Introduce(envs ...interface{}) {
	registered = append(registered, envs...)
}

// MockMany is a convenience method for mocking many environments
// at once that were previously introduced using Introduce function.
// Each of the envs is matched to a previously introduced environment
// based on type commonality and the introduced environment is
// mocked using the passed environment.
// Deprecated MockMany is being considered for deletion to remove
// the linkage between Introduce and MockMany. It simplifies the
// API by reducing complexity and footprint both at the same time.
func MockMany(envs ...interface{}) func() {
	var funcs []func()

	for _, unknown := range envs {
		unknownType := reflect.ValueOf(unknown).Type()
		for _, known := range registered {
			if reflect.ValueOf(known).Type() == unknownType {
				funcs = append(funcs, Mock(known, unknown))
				unknownType = nil
			}
		}

		if unknownType != nil {
			panic(fmt.Sprintf("Attempt to mock unregistered type via %v", unknown))
		}
	}

	return Unmock(funcs...)
}

// Unmock is handy to invoke the return values of Mock family of methods
func Unmock(funcs ...func()) func() {
	return func() {
		for _, function := range funcs {
			defer function()
		}
	}
}

func getNonExportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

type field struct {
	name     string
	value    interface{}
	exported bool
}

func toPairs(new interface{}, includeZeros bool) []field {
	var fields []field = nil

	valueOfNew := reflect.ValueOf(new).Elem()
	typeOfNew := reflect.TypeOf(new).Elem()

	for i := typeOfNew.NumField(); i > 0; {
		i--

		newField := valueOfNew.Field(i)

		var value interface{}

		exported := typeOfNew.Field(i).IsExported()
		if exported {
			value = newField.Interface()
		} else {
			value = getNonExportedField(newField).Interface()
		}

		if includeZeros || !isZero(reflect.ValueOf(value)) {
			fields = append(fields, field{typeOfNew.Field(i).Name, value, exported})
		}
	}

	return fields
}

// Mock mocks the old environment using the values in the new environment
// If the type for old environment is identical to the type of the new environment
// then any attribute with value identical to default value for its type is
// not mocked. But if types are different then the attribute is mocked regardless
// of its value
func Mock(old interface{}, new interface{}) func() {
	array := toPairs(new, reflect.TypeOf(new).Elem() != reflect.ValueOf(old).Elem().Type())
	if array == nil {
		return func() {}
	}

	array = mockHelper(old, array)
	return func() {
		mockHelper(old, array)
	}
}

func mockHelper(old interface{}, fields []field) []field {
	oldPtrVal := reflect.New(reflect.TypeOf(old))
	oldPtrVal.Elem().Set(reflect.ValueOf(old))
	valueOfOld := oldPtrVal.Elem().Elem()

	for i, f := range fields {
		fields[i].value = mockField(valueOfOld, f.name, f.value, f.exported)
	}

	return fields
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
	return !v.IsValid() || v.IsZero()
}

// MockField replaces the value of the field identified by `name` with `value`
func MockField(old interface{}, name string, value any) func() {
	typeOf := reflect.TypeOf(old).Elem()
	if typeField, found := typeOf.FieldByName(name); found {
		exported := typeField.IsExported()
		valueOfStruct := reflect.ValueOf(old).Elem()
		old := mockField(valueOfStruct, name, value, exported)
		return func() {
			mockField(valueOfStruct, name, old, exported)
		}
	}

	panic("no property with name " + name + " found")
}

// Fields allows specifying multiple mappings by using field names.
type Fields map[string]interface{}

// MockFields mocks many fields of the struct pointed to by old
func MockFields(old interface{}, fields Fields) func() {
	typeOf := reflect.TypeOf(old).Elem()
	valueOfStruct := reflect.ValueOf(old).Elem()

	var array []field
	for name, value := range fields {
		if typeField, found := typeOf.FieldByName(name); found {
			exported := typeField.IsExported()
			old := mockField(valueOfStruct, name, value, exported)
			array = append(array, field{name, old, exported})
		}
	}

	if array == nil {
		return func() {}
	}

	return func() {
		mockHelper(old, array)
	}
}

func mockField(valueOfStruct reflect.Value, name string, new any, exported bool) interface{} {
	field := valueOfStruct.FieldByName(name)
	if !exported {
		field = getNonExportedField(field)
	}
	old := field.Interface()
	field.Set(reflect.ValueOf(new))
	return old
}
