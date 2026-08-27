package utils

import (
	"maps"
	"slices"
)

func AsMap[E comparable](values ...E) map[E]E {
	result := make(map[E]E, len(values)/2)

	for idx := 0; idx < len(values)-1; idx += 2 {
		result[values[idx]] = values[idx+1]
	}

	return result
}

func Keys[K comparable, V any](value map[K]V) []K {
	return slices.Collect(maps.Keys(value))
}

func ContainsKey[K comparable, V any](value map[K]V, key K) bool {
	_, ok := value[key]

	return ok
}

func Projection(data map[string]any, keys []string) map[string]any {
	result := make(map[string]any, len(keys))

	for _, key := range keys {
		result[key] = data[key]
	}

	return result
}

func Remap[IK comparable, IV any, OK comparable, OV any](value map[IK]IV, remapper func(IK, IV) (OK, OV)) map[OK]OV {
	result := make(map[OK]OV, len(value))

	for k, v := range value {
		rk, rv := remapper(k, v)
		result[rk] = rv
	}

	return result
}
