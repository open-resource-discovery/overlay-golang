//go:build unit

package utils

import (
	"slices"
	"testing"
)

func TestFlatten(t *testing.T) {
	t.Run("multiple non-empty slices", func(t *testing.T) {
		got := Flatten([][]int{{1, 2}, {3, 4}, {5}})
		want := []int{1, 2, 3, 4, 5}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("single inner slice", func(t *testing.T) {
		got := Flatten([][]string{{"a", "b", "c"}})
		want := []string{"a", "b", "c"}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty outer slice", func(t *testing.T) {
		got := Flatten([][]int{})
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("inner slices are empty", func(t *testing.T) {
		got := Flatten([][]int{{}, {}, {}})
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("mix of empty and non-empty inner slices", func(t *testing.T) {
		got := Flatten([][]int{{}, {1, 2}, {}, {3}})
		want := []int{1, 2, 3}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("preserves order across inner slices", func(t *testing.T) {
		got := Flatten([][]int{{3, 1}, {2}})
		want := []int{3, 1, 2}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestSort(t *testing.T) {
	t.Run("integers unsorted", func(t *testing.T) {
		got := Sort([]int{3, 1, 4, 1, 5, 9, 2, 6})
		want := []int{1, 1, 2, 3, 4, 5, 6, 9}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("strings", func(t *testing.T) {
		got := Sort([]string{"banana", "apple", "cherry"})
		want := []string{"apple", "banana", "cherry"}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("already sorted", func(t *testing.T) {
		input := []int{1, 2, 3}
		got := Sort(input)
		if !slices.Equal(got, input) {
			t.Errorf("got %v, want %v", got, input)
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := Sort([]int{42})
		if !slices.Equal(got, []int{42}) {
			t.Errorf("got %v, want [42]", got)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := Sort([]int{})
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("does not mutate original", func(t *testing.T) {
		original := []int{3, 1, 2}
		_ = Sort(original)
		if !slices.Equal(original, []int{3, 1, 2}) {
			t.Errorf("original was mutated: %v", original)
		}
	})

	t.Run("duplicates preserved", func(t *testing.T) {
		got := Sort([]int{2, 2, 1, 1})
		want := []int{1, 1, 2, 2}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("keeps matching elements", func(t *testing.T) {
		got := Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
		want := []int{2, 4}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no elements match returns empty", func(t *testing.T) {
		got := Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("all elements match", func(t *testing.T) {
		input := []int{2, 4, 6}
		got := Filter(input, func(n int) bool { return n%2 == 0 })
		if !slices.Equal(got, input) {
			t.Errorf("got %v, want %v", got, input)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		got := Filter([]string{}, func(s string) bool { return true })
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("filters strings by length", func(t *testing.T) {
		got := Filter([]string{"a", "bb", "ccc", "dd"}, func(s string) bool { return len(s) > 1 })
		want := []string{"bb", "ccc", "dd"}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("does not mutate original", func(t *testing.T) {
		original := []int{1, 2, 3}
		_ = Filter(original, func(n int) bool { return n > 1 })
		if !slices.Equal(original, []int{1, 2, 3}) {
			t.Errorf("original was mutated: %v", original)
		}
	})
}

func TestMap(t *testing.T) {
	t.Run("doubles each integer", func(t *testing.T) {
		got := Map([]int{1, 2, 3}, func(_ int, v int) int { return v * 2 })
		want := []int{2, 4, 6}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("index is passed correctly", func(t *testing.T) {
		got := Map([]string{"a", "b", "c"}, func(i int, _ string) int { return i })
		want := []int{0, 1, 2}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("type change: bool to string", func(t *testing.T) {
		got := Map([]bool{true, false, true}, func(_ int, v bool) string {
			if v {
				return "yes"
			}
			return "no"
		})
		want := []string{"yes", "no", "yes"}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		got := Map([]int{}, func(_ int, v int) int { return v })
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := Map([]int{7}, func(_ int, v int) int { return v + 1 })
		if !slices.Equal(got, []int{8}) {
			t.Errorf("got %v, want [8]", got)
		}
	})

	t.Run("mapper receives correct index and value for each element", func(t *testing.T) {
		// verifies both idx and v are forwarded correctly by combining them
		got := Map([]int{10, 20, 30}, func(i int, v int) int { return i*100 + v })
		want := []int{10, 120, 230}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum of integers", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4, 5}, 0, func(acc int, v int) int { return acc + v })
		if got != 15 {
			t.Errorf("got %d, want 15", got)
		}
	})

	t.Run("product of integers", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4}, 1, func(acc int, v int) int { return acc * v })
		if got != 24 {
			t.Errorf("got %d, want 24", got)
		}
	})

	t.Run("empty slice returns initial value", func(t *testing.T) {
		got := Reduce([]int{}, 42, func(acc int, v int) int { return acc + v })
		if got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("string concatenation", func(t *testing.T) {
		got := Reduce([]string{"a", "b", "c"}, "", func(acc string, v string) string { return acc + v })
		if got != "abc" {
			t.Errorf("got %q, want \"abc\"", got)
		}
	})

	t.Run("accumulates conditionally: count even elements", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4, 5}, 0, func(acc int, v int) int {
			if v%2 == 0 {
				return acc + 1
			}
			return acc
		})
		if got != 2 {
			t.Errorf("got %d, want 2", got)
		}
	})

	t.Run("single element returns reducer(initial, element)", func(t *testing.T) {
		got := Reduce([]int{10}, 5, func(acc int, v int) int { return acc + v })
		if got != 15 {
			t.Errorf("got %d, want 15", got)
		}
	})

	t.Run("initial value affects result", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3}, 100, func(acc int, v int) int { return acc + v })
		if got != 106 {
			t.Errorf("got %d, want 106", got)
		}
	})

	t.Run("type change: int slice to string", func(t *testing.T) {
		// O (string) differs from I (int) — exercises the generic type parameters
		got := Reduce([]int{1, 2, 3}, "0", func(acc string, v int) string {
			if acc == "0" {
				return acc
			}
			return acc
		})
		// minimal smoke test that mixed-type reduce compiles and runs
		_ = got
		got2 := Reduce([]bool{true, false, true}, 0, func(acc int, v bool) int {
			if v {
				return acc + 1
			}
			return acc
		})
		if got2 != 2 {
			t.Errorf("got %d, want 2", got2)
		}
	})
}

func TestPop(t *testing.T) {
	t.Run("returns last element and remaining slice", func(t *testing.T) {
		elem, rest := Pop([]int{1, 2, 3})
		if elem != 3 {
			t.Errorf("elem: got %d, want 3", elem)
		}
		if !slices.Equal(rest, []int{1, 2}) {
			t.Errorf("rest: got %v, want [1 2]", rest)
		}
	})

	t.Run("single element returns it and empty slice", func(t *testing.T) {
		elem, rest := Pop([]int{42})
		if elem != 42 {
			t.Errorf("elem: got %d, want 42", elem)
		}
		if len(rest) != 0 {
			t.Errorf("rest: expected empty, got %v", rest)
		}
	})

	t.Run("two elements", func(t *testing.T) {
		elem, rest := Pop([]string{"a", "b"})
		if elem != "b" {
			t.Errorf("elem: got %q, want \"b\"", elem)
		}
		if !slices.Equal(rest, []string{"a"}) {
			t.Errorf("rest: got %v, want [a]", rest)
		}
	})

	t.Run("does not mutate original slice", func(t *testing.T) {
		original := []int{10, 20, 30}
		_, _ = Pop(original)
		if !slices.Equal(original, []int{10, 20, 30}) {
			t.Errorf("original was mutated: %v", original)
		}
	})

	t.Run("works with struct elements", func(t *testing.T) {
		type point struct{ x, y int }
		elem, rest := Pop([]point{{1, 2}, {3, 4}, {5, 6}})
		if elem != (point{5, 6}) {
			t.Errorf("elem: got %v, want {5 6}", elem)
		}
		if len(rest) != 2 || rest[0] != (point{1, 2}) || rest[1] != (point{3, 4}) {
			t.Errorf("rest: got %v, want [{1 2} {3 4}]", rest)
		}
	})
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected []int
		want     bool
	}{
		{"value present in list", 2, []int{1, 2, 3}, true},
		{"value not in list", 5, []int{1, 2, 3}, false},
		{"empty expected list returns false", 1, []int{}, false},
		{"single matching expected", 7, []int{7}, true},
		{"single non-matching expected", 7, []int{8}, false},
		{"first element matches", 1, []int{1, 2, 3}, true},
		{"last element matches", 3, []int{1, 2, 3}, true},
		{"zero value not in list", 0, []int{1, 2, 3}, false},
		{"zero value in list", 0, []int{1, 0, 3}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OneOf(tc.value, tc.expected...)
			if got != tc.want {
				t.Errorf("OneOf(%v, %v) = %v, want %v", tc.value, tc.expected, got, tc.want)
			}
		})
	}

	t.Run("string values", func(t *testing.T) {
		if !OneOf("banana", "apple", "banana", "cherry") {
			t.Error("expected true")
		}
	})
}
