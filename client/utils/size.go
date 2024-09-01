package utils

import "reflect"

func PackedSizeOf(v interface{}) uint {
	value := reflect.ValueOf(v)
	var totalSize uint
	for i := 0; i < value.NumField(); i++ {
		totalSize += uint(value.Field(i).Type().Size())
	}
	return totalSize
}
