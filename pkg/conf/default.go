package conf

import (
	"fmt"
	"reflect"
)

func setDefaults(v any) {
	//  Get the element of the pointer
	val := reflect.ValueOf(v).Elem()
	setDefaultsRecursive(val)
}
func setDefaultsRecursive(v reflect.Value) {
	if v.Kind() != reflect.Struct {
		return
	}
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := typ.Field(i)
		// if the field is a struct, set recursively
		if field.Kind() == reflect.Struct {
			setDefaultsRecursive(field)
		}

		// if the field is zero value and has default tag, set the default value
		if isZero(field) {
			defaultValue := fieldType.Tag.Get("default")
			if defaultValue != "" {
				// set the value for the field using reflection
				field.Set(reflect.ValueOf(parseDefaultValue(field.Kind(), defaultValue)))
			}
		}
	}
}
func isZero(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
func parseDefaultValue(kind reflect.Kind, defaultValue string) any {
	switch kind {
	case reflect.String:
		return defaultValue
	case reflect.Int:
		var i int
		_, _ = fmt.Sscanf(defaultValue, "%d", &i)
		return i
	case reflect.Int64:
		var i int64
		_, _ = fmt.Sscanf(defaultValue, "%d", &i)
		return i
	case reflect.Bool:
		var b bool
		_, _ = fmt.Sscanf(defaultValue, "%t", &b)
		return b
	case reflect.Uint32:
		var i uint32
		_, _ = fmt.Sscanf(defaultValue, "%d", &i)
		return i
	default:
		fmt.Printf("类型 %v 没有处理, 值为: %v \n", kind, defaultValue)
		panic("unhandled default case")
	}
}
