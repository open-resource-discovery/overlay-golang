package utils

import (
	"reflect"
)

func CanCast[T any](value any) bool {
	return value != nil && reflect.ValueOf(value).CanConvert(reflect.TypeOf((*T)(nil)).Elem())
}

func SafeCast[T any](value any) (result T) {
	if rtype, rvalue := reflect.TypeOf(result), reflect.ValueOf(value); value != nil && rvalue.CanConvert(rtype) {
		result = rvalue.Convert(rtype).Interface().(T)
	}

	return result
}

func IsScalar(value any) bool {
	return value != nil && OneOf(
		reflect.ValueOf(value).Kind(),
		reflect.Bool,
		reflect.String,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
	)
}
