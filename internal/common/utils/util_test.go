//go:build unit

package utils

import (
	"reflect"
	"testing"
)

func TestFirst(t *testing.T) {
	t.Run("returns first of two args", func(t *testing.T) {
		if got := First("hello", "world"); got != "hello" {
			t.Errorf("got %q, want \"hello\"", got)
		}
	})

	t.Run("returns first int", func(t *testing.T) {
		if got := First(42, "ignored"); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("single argument", func(t *testing.T) {
		if got := First(true); got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("works with multi-return function result", func(t *testing.T) {
		multiReturn := func() (int, error) { return 7, nil }
		if got := First(multiReturn()); got != 7 {
			t.Errorf("got %d, want 7", got)
		}
	})
}

func TestSecond(t *testing.T) {
	t.Run("returns second of two args", func(t *testing.T) {
		if got := Second("ignored", "world"); got != "world" {
			t.Errorf("got %q, want \"world\"", got)
		}
	})

	t.Run("returns second int", func(t *testing.T) {
		if got := Second("ignored", 99); got != 99 {
			t.Errorf("got %d, want 99", got)
		}
	})

	t.Run("works with multi-return function result", func(t *testing.T) {
		multiReturn := func() (string, int) { return "skip", 42 }
		if got := Second(multiReturn()); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("extra variadic args are ignored", func(t *testing.T) {
		if got := Second(0, "pick", "extra1", "extra2"); got != "pick" {
			t.Errorf("got %q, want \"pick\"", got)
		}
	})
}

func TestThird(t *testing.T) {
	t.Run("returns third of three args", func(t *testing.T) {
		if got := Third("a", "b", "c"); got != "c" {
			t.Errorf("got %q, want \"c\"", got)
		}
	})

	t.Run("returns third int", func(t *testing.T) {
		if got := Third(1, 2, 42); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("extra variadic args are ignored", func(t *testing.T) {
		if got := Third(0, 0, "pick", "extra"); got != "pick" {
			t.Errorf("got %q, want \"pick\"", got)
		}
	})
}

func TestClone(t *testing.T) {
	t.Run("returns deep-equal value", func(t *testing.T) {
		type S struct{ X, Y int }
		original := S{X: 10, Y: 20}
		got := Clone(original)
		if !reflect.DeepEqual(got, original) {
			t.Errorf("got %v, want %v", got, original)
		}
	})

	t.Run("is independent copy for pointer field", func(t *testing.T) {
		type S struct{ Ptr *int }
		n := 5
		original := S{Ptr: &n}
		got := Clone(original)
		*got.Ptr = 99
		if *original.Ptr != 5 {
			t.Error("Clone mutated original via pointer field")
		}
	})

	t.Run("no mutations returns unchanged clone", func(t *testing.T) {
		got := Clone(42)
		if got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("single mutation applied", func(t *testing.T) {
		type S struct{ X int }
		got := Clone(S{X: 1}, func(s *S) { s.X = 99 })
		if got.X != 99 {
			t.Errorf("got %v, want {99}", got)
		}
	})

	t.Run("multiple mutations applied in order", func(t *testing.T) {
		type S struct{ X int }
		got := Clone(S{X: 0},
			func(s *S) { s.X += 10 },
			func(s *S) { s.X *= 3 },
		)
		if got.X != 30 {
			t.Errorf("got %v, want {30}", got)
		}
	})

	t.Run("mutation does not affect original", func(t *testing.T) {
		type S struct{ X int }
		original := S{X: 5}
		_ = Clone(original, func(s *S) { s.X = 999 })
		if original.X != 5 {
			t.Errorf("original was mutated: got %v", original)
		}
	})

	t.Run("clone of map is independent", func(t *testing.T) {
		original := map[string]int{"a": 1}
		got := Clone(original)
		got["a"] = 999
		if original["a"] != 1 {
			t.Error("Clone mutated original map")
		}
	})
}

func TestTernary(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		first     string
		second    string
		want      string
	}{
		{"true returns first", true, "yes", "no", "yes"},
		{"false returns second", false, "yes", "no", "no"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Ternary(tc.condition, tc.first, tc.second)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("integer values", func(t *testing.T) {
		if got := Ternary(1 > 2, 100, 200); got != 200 {
			t.Errorf("got %d, want 200", got)
		}
	})

	t.Run("bool values", func(t *testing.T) {
		if got := Ternary(true, true, false); got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("pointer values", func(t *testing.T) {
		a, b := 1, 2
		got := Ternary(false, &a, &b)
		if got != &b {
			t.Error("expected pointer to b")
		}
	})
}

func TestPtr(t *testing.T) {
	t.Run("returns non-nil pointer", func(t *testing.T) {
		if got := Ptr(42); got == nil {
			t.Error("expected non-nil pointer, got nil")
		}
	})

	t.Run("dereferenced value equals input (int)", func(t *testing.T) {
		if got := *Ptr(42); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("dereferenced value equals input (string)", func(t *testing.T) {
		if got := *Ptr("hello"); got != "hello" {
			t.Errorf("got %q, want \"hello\"", got)
		}
	})

	t.Run("dereferenced value equals input (bool)", func(t *testing.T) {
		if got := *Ptr(true); got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("zero value input produces valid pointer", func(t *testing.T) {
		if got := *Ptr(0); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("works with struct type", func(t *testing.T) {
		type S struct{ X int }
		got := Ptr(S{X: 7})
		if got == nil || got.X != 7 {
			t.Errorf("got %v, want &{7}", got)
		}
	})

	t.Run("returns independent pointer — modifying one does not affect another", func(t *testing.T) {
		p1 := Ptr(10)
		p2 := Ptr(10)
		*p1 = 99
		if *p2 != 10 {
			t.Errorf("p2 was affected by mutation of p1: got %d, want 10", *p2)
		}
	})
}

func TestDeepMerge(t *testing.T) {
	t.Run("nil destination returns clone of source", func(t *testing.T) {
		src := map[string]any{"x": 1}
		got := DeepMerge(nil, src).(map[string]any)
		if got["x"] != 1 {
			t.Errorf("got %v, want 1", got["x"])
		}
	})

	t.Run("scalar destination returns clone of source", func(t *testing.T) {
		got := DeepMerge("old", "new")
		if got != "new" {
			t.Errorf("got %v, want \"new\"", got)
		}
	})

	t.Run("source scalar overwrites destination scalar", func(t *testing.T) {
		dst := map[string]any{"k": "old"}
		src := map[string]any{"k": "new"}
		got := DeepMerge(dst, src).(map[string]any)
		if got["k"] != "new" {
			t.Errorf("got %v, want \"new\"", got["k"])
		}
	})

	t.Run("destination-only key is preserved in result", func(t *testing.T) {
		dst := map[string]any{"a": 1, "b": 2}
		src := map[string]any{"b": 99}
		got := DeepMerge(dst, src).(map[string]any)
		if got["a"] != 1 {
			t.Errorf("key \"a\": got %v, want 1", got["a"])
		}
		if got["b"] != 99 {
			t.Errorf("key \"b\": got %v, want 99", got["b"])
		}
	})

	t.Run("source-only key is added; destination key is preserved", func(t *testing.T) {
		dst := map[string]any{"a": 1}
		src := map[string]any{"b": 2}
		got := DeepMerge(dst, src).(map[string]any)
		if got["a"] != 1 {
			t.Errorf("key \"a\": got %v, want 1", got["a"])
		}
		if got["b"] != 2 {
			t.Errorf("key \"b\": got %v, want 2", got["b"])
		}
	})

	t.Run("nested maps merged recursively", func(t *testing.T) {
		dst := map[string]any{"nested": map[string]any{"x": 1, "y": 2}}
		src := map[string]any{"nested": map[string]any{"y": 99, "z": 3}}
		got := DeepMerge(dst, src).(map[string]any)
		nested, ok := got["nested"].(map[string]any)
		if !ok {
			t.Fatal("nested is not map[string]any")
		}
		if nested["x"] != 1 {
			t.Errorf("nested.x: got %v, want 1", nested["x"])
		}
		if nested["y"] != 99 {
			t.Errorf("nested.y: got %v, want 99", nested["y"])
		}
		if nested["z"] != 3 {
			t.Errorf("nested.z: got %v, want 3", nested["z"])
		}
	})

	t.Run("arrays appended: source items after destination items", func(t *testing.T) {
		dst := map[string]any{"arr": []any{1, 2}}
		src := map[string]any{"arr": []any{3, 4}}
		got := DeepMerge(dst, src).(map[string]any)
		arr, ok := got["arr"].([]any)
		if !ok {
			t.Fatal("arr is not []any")
		}
		want := []any{1, 2, 3, 4}
		if len(arr) != len(want) {
			t.Fatalf("len = %d, want %d", len(arr), len(want))
		}
		for i, v := range want {
			if arr[i] != v {
				t.Errorf("[%d] got %v, want %v", i, arr[i], v)
			}
		}
	})

	t.Run("empty source returns clone of destination", func(t *testing.T) {
		dst := map[string]any{"a": 1}
		got := DeepMerge(dst, map[string]any{}).(map[string]any)
		if got["a"] != 1 {
			t.Errorf("got %v, want 1", got["a"])
		}
	})

	t.Run("empty destination returns clone of source", func(t *testing.T) {
		src := map[string]any{"x": 42}
		got := DeepMerge(map[string]any{}, src).(map[string]any)
		if got["x"] != 42 {
			t.Errorf("got %v, want 42", got["x"])
		}
	})

	t.Run("does not mutate destination", func(t *testing.T) {
		dst := map[string]any{"k": "original"}
		_ = DeepMerge(dst, map[string]any{"k": "changed"})
		if dst["k"] != "original" {
			t.Errorf("destination was mutated: got %v", dst["k"])
		}
	})

	t.Run("does not mutate source", func(t *testing.T) {
		src := map[string]any{"k": "src"}
		_ = DeepMerge(map[string]any{"k": "dst"}, src)
		if src["k"] != "src" {
			t.Errorf("source was mutated: got %v", src["k"])
		}
	})

	t.Run("scalar in dst, map in src: source wins", func(t *testing.T) {
		dst := map[string]any{"k": "scalar"}
		src := map[string]any{"k": map[string]any{"nested": true}}
		got := DeepMerge(dst, src).(map[string]any)
		if _, ok := got["k"].(map[string]any); !ok {
			t.Errorf("expected map, got %T", got["k"])
		}
	})

	t.Run("map in dst, scalar in src: source wins", func(t *testing.T) {
		dst := map[string]any{"k": map[string]any{"nested": true}}
		src := map[string]any{"k": "scalar"}
		got := DeepMerge(dst, src).(map[string]any)
		if got["k"] != "scalar" {
			t.Errorf("got %v, want \"scalar\"", got["k"])
		}
	})

	t.Run("deeply nested merge (3 levels)", func(t *testing.T) {
		dst := map[string]any{
			"l1": map[string]any{
				"l2": map[string]any{"a": 1, "b": 2},
			},
		}
		src := map[string]any{
			"l1": map[string]any{
				"l2": map[string]any{"b": 99, "c": 3},
			},
		}
		got := DeepMerge(dst, src).(map[string]any)
		l2 := got["l1"].(map[string]any)["l2"].(map[string]any)
		if l2["a"] != 1 {
			t.Errorf("l2.a: got %v, want 1", l2["a"])
		}
		if l2["b"] != 99 {
			t.Errorf("l2.b: got %v, want 99", l2["b"])
		}
		if l2["c"] != 3 {
			t.Errorf("l2.c: got %v, want 3", l2["c"])
		}
	})

	t.Run("mutating destination array after merge does not affect result", func(t *testing.T) {
		dst := map[string]any{"arr": []any{1, 2}}
		src := map[string]any{"arr": []any{3}}
		got := DeepMerge(dst, src).(map[string]any)
		// Modify dst's original slice; result must remain independent.
		dst["arr"] = []any{99, 99}
		arr := got["arr"].([]any)
		if len(arr) != 3 || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
			t.Errorf("result was affected by mutation of destination: %v", arr)
		}
	})
}
