//go:build unit

package xml2json

import (
	"testing"
)

func TestNewAttributes(t *testing.T) {
	t.Run("even pairs create expected map", func(t *testing.T) {
		a := NewAttributes("key1", "val1", "key2", "val2")
		if len(a) != 2 {
			t.Fatalf("len = %d, want 2", len(a))
		}
		if a["key1"] != "val1" {
			t.Errorf("key1 = %q, want \"val1\"", a["key1"])
		}
		if a["key2"] != "val2" {
			t.Errorf("key2 = %q, want \"val2\"", a["key2"])
		}
	})

	t.Run("no arguments returns empty Attributes", func(t *testing.T) {
		a := NewAttributes()
		if len(a) != 0 {
			t.Errorf("expected empty, got %v", a)
		}
	})

	t.Run("odd number of values drops the unpaired last element", func(t *testing.T) {
		a := NewAttributes("key1", "val1", "orphan")
		if len(a) != 1 {
			t.Fatalf("len = %d, want 1", len(a))
		}
		if a["key1"] != "val1" {
			t.Errorf("key1 = %q, want \"val1\"", a["key1"])
		}
		if _, ok := a["orphan"]; ok {
			t.Error("orphan key should not be present")
		}
	})

	t.Run("single pair", func(t *testing.T) {
		a := NewAttributes("only", "value")
		if len(a) != 1 || a["only"] != "value" {
			t.Errorf("got %v, want {only:value}", a)
		}
	})

	t.Run("duplicate keys: last value wins", func(t *testing.T) {
		a := NewAttributes("k", "first", "k", "second")
		if a["k"] != "second" {
			t.Errorf("k = %q, want \"second\"", a["k"])
		}
		if len(a) != 1 {
			t.Errorf("len = %d, want 1", len(a))
		}
	})

	t.Run("empty string key and value are valid", func(t *testing.T) {
		a := NewAttributes("", "")
		if len(a) != 1 {
			t.Fatalf("len = %d, want 1", len(a))
		}
		if a[""] != "" {
			t.Errorf("empty key: got %q, want \"\"", a[""])
		}
	})
}

func TestAttributesHas(t *testing.T) {
	a := NewAttributes("name", "Alice", "age", "30")

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"present key", "name", true},
		{"another present key", "age", true},
		{"absent key", "city", false},
		{"empty string key absent", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := a.Has(tc.key)
			if got != tc.want {
				t.Errorf("Has(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}

	t.Run("empty Attributes always returns false", func(t *testing.T) {
		empty := NewAttributes()
		if empty.Has("anything") {
			t.Error("expected false for empty Attributes")
		}
	})

	t.Run("empty string key is found when present", func(t *testing.T) {
		a2 := NewAttributes("", "emptykey")
		if !a2.Has("") {
			t.Error("expected true for empty-string key that was inserted")
		}
	})
}

func TestAttributesGet(t *testing.T) {
	a := NewAttributes("color", "blue", "size", "large")

	t.Run("returns value for present key", func(t *testing.T) {
		if got := a.Get("color"); got != "blue" {
			t.Errorf("Get(\"color\") = %q, want \"blue\"", got)
		}
	})

	t.Run("absent key returns empty string", func(t *testing.T) {
		if got := a.Get("missing"); got != "" {
			t.Errorf("Get(\"missing\") = %q, want \"\"", got)
		}
	})

	t.Run("key present with empty value returns empty string", func(t *testing.T) {
		// Distinguishes the present-but-empty case from the absent case;
		// both return "" but Has() would differ.
		a2 := NewAttributes("blank", "")
		if got := a2.Get("blank"); got != "" {
			t.Errorf("Get(\"blank\") = %q, want \"\"", got)
		}
		if !a2.Has("blank") {
			t.Error("Has(\"blank\") should be true for a key that was inserted with empty value")
		}
	})

	t.Run("empty string key returns stored value when present", func(t *testing.T) {
		a2 := NewAttributes("", "sentinel")
		if got := a2.Get(""); got != "sentinel" {
			t.Errorf("Get(\"\") = %q, want \"sentinel\"", got)
		}
	})
}
