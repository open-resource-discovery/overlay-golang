//go:build unit

package utils

import (
	"testing"
)

func TestCanCast(t *testing.T) {
	t.Run("same type returns true (string)", func(t *testing.T) {
		if !CanCast[string]("hello") {
			t.Error("expected true for string -> string")
		}
	})

	t.Run("same type returns true (int)", func(t *testing.T) {
		if !CanCast[int](42) {
			t.Error("expected true for int -> int")
		}
	})

	t.Run("same type returns true (bool)", func(t *testing.T) {
		if !CanCast[bool](true) {
			t.Error("expected true for bool -> bool")
		}
	})

	t.Run("numeric conversion returns true (int -> float64)", func(t *testing.T) {
		if !CanCast[float64](42) {
			t.Error("expected true for int -> float64")
		}
	})

	t.Run("numeric conversion returns true (float32 -> float64)", func(t *testing.T) {
		if !CanCast[float64](float32(3.14)) {
			t.Error("expected true for float32 -> float64")
		}
	})

	t.Run("numeric conversion returns true (int -> int64)", func(t *testing.T) {
		if !CanCast[int64](42) {
			t.Error("expected true for int -> int64")
		}
	})

	t.Run("nil input returns false", func(t *testing.T) {
		if CanCast[string](nil) {
			t.Error("expected false for nil input")
		}
	})

	t.Run("non-convertible type returns false (string -> map)", func(t *testing.T) {
		if CanCast[map[string]any]("not a map") {
			t.Error("expected false for string -> map[string]any")
		}
	})

	t.Run("non-convertible type returns false (bool -> slice)", func(t *testing.T) {
		if CanCast[[]any](true) {
			t.Error("expected false for bool -> []any")
		}
	})

	t.Run("non-convertible type returns false (int -> map)", func(t *testing.T) {
		if CanCast[map[string]any](42) {
			t.Error("expected false for int -> map[string]any")
		}
	})

	t.Run("same type returns true (map)", func(t *testing.T) {
		if !CanCast[map[string]any](map[string]any{"k": "v"}) {
			t.Error("expected true for map[string]any -> map[string]any")
		}
	})
}

func TestSafeCast(t *testing.T) {
	t.Run("same type returns value unchanged (string)", func(t *testing.T) {
		if got := SafeCast[string]("hello"); got != "hello" {
			t.Errorf("got %q, want \"hello\"", got)
		}
	})

	t.Run("same type returns value unchanged (int)", func(t *testing.T) {
		if got := SafeCast[int](42); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("same type returns value unchanged (bool)", func(t *testing.T) {
		if got := SafeCast[bool](true); got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("same type returns value unchanged (map)", func(t *testing.T) {
		input := map[string]any{"key": "val"}
		got := SafeCast[map[string]any](input)
		if got == nil || got["key"] != "val" {
			t.Errorf("got %v, want map with key=val", got)
		}
	})

	t.Run("nil input returns zero value (string)", func(t *testing.T) {
		if got := SafeCast[string](nil); got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	t.Run("nil input returns zero value (int)", func(t *testing.T) {
		if got := SafeCast[int](nil); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("nil input returns zero value (map)", func(t *testing.T) {
		if got := SafeCast[map[string]any](nil); got != nil {
			t.Errorf("got %v, want nil map", got)
		}
	})

	t.Run("non-convertible type returns zero value (map target, string input)", func(t *testing.T) {
		// string is not ConvertibleTo map[string]any, so SafeCast returns nil.
		if got := SafeCast[map[string]any]("not a map"); got != nil {
			t.Errorf("got %v, want nil for non-convertible type", got)
		}
	})

	t.Run("non-convertible type returns zero value (slice target, bool input)", func(t *testing.T) {
		// bool is not ConvertibleTo []any, so SafeCast returns nil.
		if got := SafeCast[[]any](true); got != nil {
			t.Errorf("got %v, want nil for non-convertible type", got)
		}
	})

	t.Run("non-convertible type returns zero value (map target, int input)", func(t *testing.T) {
		// int is not ConvertibleTo map[string]any, so SafeCast returns nil.
		if got := SafeCast[map[string]any](42); got != nil {
			t.Errorf("got %v, want nil for non-convertible type", got)
		}
	})
}
