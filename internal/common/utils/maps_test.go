//go:build unit

package utils

import (
	"maps"
	"slices"
	"testing"
)

func TestAsMap(t *testing.T) {
	t.Run("even number of values", func(t *testing.T) {
		got := AsMap("a", "1", "b", "2")
		want := map[string]string{"a": "1", "b": "2"}
		if !maps.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("odd number of values drops last unpaired element", func(t *testing.T) {
		got := AsMap("a", "1", "orphan")
		want := map[string]string{"a": "1"}
		if !maps.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no values returns empty map", func(t *testing.T) {
		got := AsMap[string]()
		if len(got) != 0 {
			t.Fatalf("expected empty map, got %v", got)
		}
	})

	t.Run("integer keys and values", func(t *testing.T) {
		got := AsMap(1, 10, 2, 20, 3, 30)
		want := map[int]int{1: 10, 2: 20, 3: 30}
		if !maps.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("duplicate keys: last write wins", func(t *testing.T) {
		got := AsMap("x", "first", "x", "second")
		if got["x"] != "second" {
			t.Errorf("got %q, want \"second\"", got["x"])
		}
		if len(got) != 1 {
			t.Errorf("len = %d, want 1", len(got))
		}
	})
}

func TestKeys(t *testing.T) {
	t.Run("returns all keys", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := Keys(m)
		slices.Sort(got)
		want := []string{"a", "b", "c"}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty map returns empty slice", func(t *testing.T) {
		got := Keys(map[string]int{})
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %v", got)
		}
	})

	t.Run("single entry", func(t *testing.T) {
		got := Keys(map[int]string{42: "hello"})
		if len(got) != 1 || got[0] != 42 {
			t.Errorf("got %v, want [42]", got)
		}
	})
}

func TestContainsKey(t *testing.T) {
	m := map[string]int{"x": 1, "y": 2}

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"present key", "x", true},
		{"another present key", "y", true},
		{"absent key", "z", false},
		{"empty string key absent", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsKey(m, tc.key)
			if got != tc.want {
				t.Errorf("ContainsKey(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}

	t.Run("empty map always false", func(t *testing.T) {
		if ContainsKey(map[string]int{}, "anything") {
			t.Error("expected false for empty map")
		}
	})

	t.Run("zero-value key present", func(t *testing.T) {
		m2 := map[string]int{"": 99}
		if !ContainsKey(m2, "") {
			t.Error("expected true for zero-value key")
		}
	})
}

func TestProjection(t *testing.T) {
	data := map[string]any{
		"name": "Alice",
		"age":  30,
		"city": "Berlin",
	}

	t.Run("subset of keys", func(t *testing.T) {
		got := Projection(data, []string{"name", "age"})
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got["name"] != "Alice" {
			t.Errorf("name: got %v, want \"Alice\"", got["name"])
		}
		if got["age"] != 30 {
			t.Errorf("age: got %v, want 30", got["age"])
		}
	})

	t.Run("missing keys produce nil values", func(t *testing.T) {
		got := Projection(data, []string{"name", "missing"})
		if got["name"] != "Alice" {
			t.Errorf("name: got %v, want \"Alice\"", got["name"])
		}
		if got["missing"] != nil {
			t.Errorf("missing: got %v, want nil", got["missing"])
		}
	})

	t.Run("empty key list returns empty map", func(t *testing.T) {
		got := Projection(data, []string{})
		if len(got) != 0 {
			t.Fatalf("expected empty map, got %v", got)
		}
	})

	t.Run("all keys", func(t *testing.T) {
		got := Projection(data, []string{"name", "age", "city"})
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
	})

	t.Run("empty source map with keys returns nil values", func(t *testing.T) {
		got := Projection(map[string]any{}, []string{"a", "b"})
		if got["a"] != nil || got["b"] != nil {
			t.Errorf("expected nil values, got %v", got)
		}
	})
}

func TestRemap(t *testing.T) {
	t.Run("transforms keys and values", func(t *testing.T) {
		input := map[int]int{1: 10, 2: 20, 3: 30}
		got := Remap(input, func(k int, v int) (int, int) {
			return k * 2, v + 1
		})
		want := map[int]int{2: 11, 4: 21, 6: 31}
		if !maps.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		got := Remap(map[string]int{}, func(k string, v int) (string, int) {
			return k, v
		})
		if len(got) != 0 {
			t.Fatalf("expected empty map, got %v", got)
		}
	})

	t.Run("identity remapper preserves all entries", func(t *testing.T) {
		input := map[string]string{"x": "1", "y": "2"}
		got := Remap(input, func(k string, v string) (string, string) {
			return k, v
		})
		if !maps.Equal(got, input) {
			t.Errorf("got %v, want %v", got, input)
		}
	})

	t.Run("change value type: int to bool", func(t *testing.T) {
		input := map[string]int{"a": 1, "b": 0, "c": -1}
		got := Remap(input, func(k string, v int) (string, bool) {
			return k, v > 0
		})
		want := map[string]bool{"a": true, "b": false, "c": false}
		if !maps.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("colliding remapped keys: one entry survives", func(t *testing.T) {
		// map iteration order is non-deterministic, so either value is valid
		input := map[int]int{1: 10, 2: 20}
		got := Remap(input, func(k int, v int) (string, int) {
			return "same", v
		})
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got["same"] != 10 && got["same"] != 20 {
			t.Errorf("unexpected value %d, want 10 or 20", got["same"])
		}
	})
}
