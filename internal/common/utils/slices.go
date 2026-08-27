package utils

import (
	"cmp"
	"slices"
)

func Pop[E any](elements []E) (E, []E) {
	return elements[len(elements)-1], elements[:len(elements)-1]
}

func Flatten[E any](elements [][]E) []E {
	result := make([]E, 0, len(elements))

	for idx, _ := range elements {
		result = append(result, elements[idx]...)
	}

	return result
}

func Sort[T cmp.Ordered](elements []T) []T {
	result := slices.Clone(elements)

	slices.Sort(result)

	return result
}

func Filter[T any](elements []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(elements))

	for _, element := range elements {
		if predicate(element) {
			result = append(result, element)
		}
	}

	return slices.Clip(result)
}

func Map[I any, O any](elements []I, mapper func(int, I) O) []O {
	result := make([]O, 0, len(elements))

	for idx, element := range elements {
		result = append(result, mapper(idx, element))
	}

	return result
}

func Reduce[I any, O any](elements []I, initial O, reducer func(O, I) O) O {
	result := initial

	for _, element := range elements {
		result = reducer(result, element)
	}

	return result
}

func OneOf[T comparable](value T, expected ...T) bool {
	return slices.Index(expected, value) >= 0
}
