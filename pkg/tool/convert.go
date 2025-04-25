package tool

import (
	"fmt"
	"reflect"
	"strconv"
)

// ConvertValueToString converts the value to string
func ConvertValueToString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Ptr:
		switch value.Type().Elem().Kind() {
		case reflect.Bool:
			return fmt.Sprintf("%v", value.Elem().Bool())
		default:
			return ""
		}
	default:
		return ""
	}
}
