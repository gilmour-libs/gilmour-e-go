package gilmour

import (
	"errors"
	"fmt"
	"reflect"
)

func reflectVal(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr:
		v = reflectVal(reflect.Indirect(v))
	case reflect.Interface:
		v = reflectVal(v.Elem())
	}
	return v
}

func typeNval(s interface{}) (reflect.Type, reflect.Value) {
	sval := reflectVal(reflect.ValueOf(s))
	return sval.Type(), sval
}

// merges the value from one interface to another.
// Currently only map is supported, structs coming in later.
//
// Example:
// t := map[string]interface{}{"b": 2}
//
// m := Message{Data: map[string]interface{}{"a": 1, "c": 3}}
// compositionMerge(m.Data, &t)
// Yields: map[b:2 a:1 c:3]
//
// m2 := Message{Data: []interface{}{"a", "b"}}
// compositionMerge(m2.Data, &t)
// Yields: Cannot convert map[string]interface {} to []interface {}
//
// m3 := Message{Data: &map[string]interface{}{"a": 1}}
// compositionMerge(m3.Data, &t)
// Yields: &map[a:1 b:2]

func compositionMerge(s interface{}, m interface{}) error {
	source_type, source_val := typeNval(s)
	merge_type, merge_val := typeNval(m)

	if !merge_type.ConvertibleTo(source_type) {
		return errors.New(fmt.Sprintf(
			"Cannot convert %v to %v", merge_type, source_type,
		))
	}

	t := source_type.Kind()
	switch t {
	case reflect.Struct:
		return errors.New("Structs are not supported yet. Coming Soon!")
	case reflect.Map:
		for _, key := range merge_val.MapKeys() {
			val := merge_val.MapIndex(key)
			source_val.SetMapIndex(key, val)
		}
	case reflect.Ptr:
		return errors.New("Pointer should have been converted to its Elem")
	default:
		return errors.New(fmt.Sprintf(
			"Can only merge Maps at the moment. Found %v", t,
		))
	}

	return nil
}
