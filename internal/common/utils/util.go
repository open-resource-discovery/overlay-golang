package utils

import (
	"fmt"

	"github.com/huandu/go-clone"
)

func Ptr[T any](value T) *T {
	return &value
}

func First[S any](s S, _ ...any) S {
	return s
}

func Second[S any](_ any, s S, _ ...any) S {
	return s
}

func Third[S any](_ any, _ any, s S, _ ...any) S {
	return s
}

func Clone[V any](value V, mutations ...func(*V)) V {
	result := clone.Clone(value).(V)

	for _, mutation := range mutations {
		mutation(&result)
	}

	return result
}

func Ternary[V any](condition bool, first V, second V) V {
	if condition {
		return first
	}

	return second
}

func DeepMerge(destination any, source any) any {
	if destination == nil || IsScalar(destination) {
		return clone.Clone(source)
	}

	if CanCast[[]any](destination) {
		return append(clone.Clone(destination).([]any), SafeCast[[]any](clone.Clone(source))...)
	}

	if CanCast[map[string]any](destination) {
		out := clone.Clone(destination).(map[string]any)

		for k, v := range SafeCast[map[string]any](source) {
			if dstVal, ok := out[k]; ok {
				dstMap, dstIsMap := dstVal.(map[string]any)
				srcMap, srcIsMap := v.(map[string]any)
				if dstIsMap && srcIsMap {
					out[k] = DeepMerge(dstMap, srcMap)
					continue
				}
				dstArr, dstIsArr := dstVal.([]any)
				srcArr, srcIsArr := v.([]any)
				if dstIsArr && srcIsArr {
					out[k] = append(dstArr, srcArr...)
					continue
				}
			}
			out[k] = v
		}

		return out
	}

	return DeepMerge(SafeCast[map[string]any](destination), SafeCast[map[string]any](source))
}

func ToString(value any) string {
	return fmt.Sprintf("%v", value)
}
