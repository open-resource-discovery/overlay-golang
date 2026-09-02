//go:build unit

package marshaller

import (
	"strings"
	"testing"
)

func roundtrip(t *testing.T, media, content string) (any, string) {
	t.Helper()
	parsed, err := Unmarshal(media, content)
	if err != nil {
		t.Fatalf("Unmarshal(%q, ...) error: %v", media, err)
	}
	out, err := Marshal(media, parsed)
	if err != nil {
		t.Fatalf("Marshal(%q, ...) error: %v", media, err)
	}
	return parsed, out
}

func TestMarshal(t *testing.T) {
	t.Run("application/json: marshals map to JSON", func(t *testing.T) {
		got, err := Marshal("application/json", map[string]any{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, `"key"`) || !strings.Contains(got, `"value"`) {
			t.Errorf("expected JSON with key/value, got %q", got)
		}
	})

	t.Run("text/json: treated identically to application/json", func(t *testing.T) {
		got, err := Marshal("text/json", map[string]any{"x": 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, `"x"`) {
			t.Errorf("expected JSON with 'x', got %q", got)
		}
	})

	t.Run("application/json: media type with parameters accepted", func(t *testing.T) {
		_, err := Marshal("application/json; charset=utf-8", map[string]any{"a": true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("application/yaml: marshals map to YAML", func(t *testing.T) {
		got, err := Marshal("application/yaml", map[string]any{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "key: value") {
			t.Errorf("expected YAML 'key: value', got %q", got)
		}
	})

	t.Run("application/yaml: empty map produces non-error output", func(t *testing.T) {
		// exercises the utils.Ternary nil-guard on yaml.Marshal result
		got, err := Marshal("application/yaml", map[string]any{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// yaml.Marshal of empty map produces "{}\n"; the nil guard returns "" for a nil result
		_ = got
	})

	t.Run("text/yaml: treated identically to application/yaml", func(t *testing.T) {
		got, err := Marshal("text/yaml", map[string]any{"num": 42})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "num") {
			t.Errorf("expected YAML with 'num', got %q", got)
		}
	})

	t.Run("application/xml: marshals Document back to XML", func(t *testing.T) {
		doc, err := Unmarshal("application/xml", `<root><child>hello</child></root>`)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		got, err := Marshal("application/xml", doc)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if !strings.Contains(got, "root") || !strings.Contains(got, "child") || !strings.Contains(got, "hello") {
			t.Errorf("expected XML with root/child/hello, got %q", got)
		}
	})

	t.Run("text/xml: treated identically to application/xml", func(t *testing.T) {
		doc, err := Unmarshal("text/xml", `<item/>`)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		got, err := Marshal("text/xml", doc)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if !strings.Contains(got, "item") {
			t.Errorf("expected XML with 'item', got %q", got)
		}
	})

	t.Run("unsupported media type returns error", func(t *testing.T) {
		_, err := Marshal("application/octet-stream", map[string]any{})
		if err == nil {
			t.Fatal("expected error for unsupported media type, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported media type") {
			t.Errorf("error message %q does not mention 'unsupported media type'", err.Error())
		}
	})

	t.Run("empty media type returns error", func(t *testing.T) {
		_, err := Marshal("", map[string]any{})
		if err == nil {
			t.Error("expected error for empty media type, got nil")
		}
	})
}

func TestUnmarshal(t *testing.T) {
	t.Run("application/json: parses object fields", func(t *testing.T) {
		got, err := Unmarshal("application/json", `{"name":"Alice","age":30}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if m["name"] != "Alice" {
			t.Errorf("name: got %v, want \"Alice\"", m["name"])
		}
		if m["age"] != int64(30) {
			t.Errorf("age: got %T(%v), want int64(30)", m["age"], m["age"])
		}
	})

	t.Run("application/json: parses array with correct elements", func(t *testing.T) {
		got, err := Unmarshal("application/json", `[1,2,3]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", got)
		}
		if len(arr) != 3 {
			t.Fatalf("len = %d, want 3", len(arr))
		}
		if arr[0] != int64(1) || arr[1] != int64(2) || arr[2] != int64(3) {
			t.Errorf("elements: got %v (types %T), want [int64(1) int64(2) int64(3)]", arr, arr[0])
		}
	})

	t.Run("text/json: routes to JSON parser and returns parsed content", func(t *testing.T) {
		got, err := Unmarshal("text/json", `{"x":1}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if m["x"] != int64(1) {
			t.Errorf("x: got %T(%v), want int64(1)", m["x"], m["x"])
		}
	})

	t.Run("application/json: invalid JSON returns error", func(t *testing.T) {
		_, err := Unmarshal("application/json", `{bad json}`)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("application/json: empty string returns nil without error", func(t *testing.T) {
		// oj.ParseString("") returns (nil, nil) — empty input is not treated as
		// a parse error by the underlying library.
		got, err := Unmarshal("application/json", ``)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil for empty input, got %v", got)
		}
	})

	t.Run("application/yaml: parses string and integer scalars", func(t *testing.T) {
		got, err := Unmarshal("application/yaml", "key: value\nnum: 42\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if m["key"] != "value" {
			t.Errorf("key: got %v, want \"value\"", m["key"])
		}
		if m["num"] != 42 {
			t.Errorf("num: got %v, want 42", m["num"])
		}
	})

	t.Run("application/yaml: empty string returns empty map without error", func(t *testing.T) {
		got, err := Unmarshal("application/yaml", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if len(m) != 0 {
			t.Errorf("expected empty map, got %v", m)
		}
	})

	t.Run("text/yaml: routes to YAML parser and returns parsed content", func(t *testing.T) {
		got, err := Unmarshal("text/yaml", "a: b\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if m["a"] != "b" {
			t.Errorf("a: got %v, want \"b\"", m["a"])
		}
	})

	t.Run("application/yaml: nested map", func(t *testing.T) {
		got, err := Unmarshal("application/yaml", "outer:\n  inner: 1\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := got.(map[string]any)
		inner, ok := m["outer"].(map[string]any)
		if !ok {
			t.Fatalf("expected nested map, got %T", m["outer"])
		}
		if inner["inner"] != 1 {
			t.Errorf("inner: got %v, want 1", inner["inner"])
		}
	})

	t.Run("application/xml: parses element and text content", func(t *testing.T) {
		got, err := Unmarshal("application/xml", `<root><child>hello</child></root>`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out, err := Marshal("application/xml", got)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if !strings.Contains(out, "root") || !strings.Contains(out, "child") || !strings.Contains(out, "hello") {
			t.Errorf("parsed document missing expected elements: %q", out)
		}
	})

	t.Run("text/xml: routes to XML parser and returns parsed content", func(t *testing.T) {
		got, err := Unmarshal("text/xml", `<item attr="x"/>`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out, err := Marshal("text/xml", got)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if !strings.Contains(out, "item") || !strings.Contains(out, `attr="x"`) {
			t.Errorf("parsed document missing expected content: %q", out)
		}
	})

	t.Run("application/xml: invalid XML returns error", func(t *testing.T) {
		_, err := Unmarshal("application/xml", `<unclosed>`)
		if err == nil {
			t.Error("expected error for invalid XML, got nil")
		}
	})

	t.Run("application/xml: media type with parameters accepted", func(t *testing.T) {
		_, err := Unmarshal("application/xml; charset=utf-8", `<r/>`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unsupported media type returns error and nil value", func(t *testing.T) {
		_, err := Unmarshal("application/octet-stream", "data")
		if err == nil {
			t.Fatal("expected error for unsupported media type, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported media type") {
			t.Errorf("error message %q does not mention 'unsupported media type'", err.Error())
		}
	})

	t.Run("empty media type returns error", func(t *testing.T) {
		_, err := Unmarshal("", "{}")
		if err == nil {
			t.Error("expected error for empty media type, got nil")
		}
	})
}

// ---- MustMarshal ------------------------------------------------------------

func TestMustMarshal_JSON_ReturnsString(t *testing.T) {
	got := MustMarshal("application/json", map[string]any{"key": "value"})
	if !strings.Contains(got, `"key"`) || !strings.Contains(got, `"value"`) {
		t.Errorf("expected JSON with key/value, got %q", got)
	}
}

func TestMustMarshal_YAML_ReturnsString(t *testing.T) {
	got := MustMarshal("application/yaml", map[string]any{"key": "value"})
	if !strings.Contains(got, "key: value") {
		t.Errorf("expected YAML 'key: value', got %q", got)
	}
}

func TestMustMarshal_XML_ReturnsString(t *testing.T) {
	doc, err := Unmarshal("application/xml", `<root><child>hello</child></root>`)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	got := MustMarshal("application/xml", doc)
	if !strings.Contains(got, "root") || !strings.Contains(got, "child") {
		t.Errorf("expected XML with root/child, got %q", got)
	}
}

func TestMustMarshal_UnsupportedMediaType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unsupported media type, got none")
		}
	}()
	MustMarshal("application/octet-stream", map[string]any{})
}

func TestMustMarshal_MediaTypeWithParameters_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	MustMarshal("application/json; charset=utf-8", map[string]any{"a": 1})
}

// ---- MustUnmarshal ----------------------------------------------------------

func TestMustUnmarshal_JSON_ReturnsMap(t *testing.T) {
	got := MustUnmarshal("application/json", `{"name":"Alice"}`)
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if m["name"] != "Alice" {
		t.Errorf("name: got %v, want \"Alice\"", m["name"])
	}
}

func TestMustUnmarshal_YAML_ReturnsMap(t *testing.T) {
	got := MustUnmarshal("application/yaml", "key: value\n")
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if m["key"] != "value" {
		t.Errorf("key: got %v, want \"value\"", m["key"])
	}
}

func TestMustUnmarshal_XML_ReturnsDocument(t *testing.T) {
	got := MustUnmarshal("application/xml", `<root><child>hello</child></root>`)
	out := MustMarshal("application/xml", got)
	if !strings.Contains(out, "root") || !strings.Contains(out, "child") {
		t.Errorf("expected document with root/child, got %q", out)
	}
}

func TestMustUnmarshal_InvalidJSON_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid JSON, got none")
		}
	}()
	MustUnmarshal("application/json", `{bad json}`)
}

func TestMustUnmarshal_InvalidYAML_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid YAML, got none")
		}
	}()
	MustUnmarshal("application/yaml", ":\n  - bad: [unclosed")
}

func TestMustUnmarshal_InvalidXML_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid XML, got none")
		}
	}()
	MustUnmarshal("application/xml", `<unclosed>`)
}

func TestMustUnmarshal_UnsupportedMediaType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unsupported media type, got none")
		}
	}()
	MustUnmarshal("application/octet-stream", "data")
}

func TestMustUnmarshal_MediaTypeWithParameters_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	MustUnmarshal("application/json; charset=utf-8", `{"ok":true}`)
}

// ---- MustMarshal / MustUnmarshal roundtrip ----------------------------------

func TestMustMarshalUnmarshalRoundtrip_JSON(t *testing.T) {
	original := map[string]any{"name": "Alice", "score": int64(99)}
	got := MustUnmarshal("application/json", MustMarshal("application/json", original))
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if m["name"] != "Alice" {
		t.Errorf("name: got %v, want \"Alice\"", m["name"])
	}
	if m["score"] != int64(99) {
		t.Errorf("score: got %v (%T), want int64(99)", m["score"], m["score"])
	}
}

func TestMustMarshalUnmarshalRoundtrip_YAML(t *testing.T) {
	out := MustMarshal("application/yaml", map[string]any{"key": "value"})
	got := MustUnmarshal("application/yaml", out)
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if m["key"] != "value" {
		t.Errorf("key: got %v, want \"value\"", m["key"])
	}
}

// ---- TestMarshalUnmarshalRoundtrip ------------------------------------------

func TestMarshalUnmarshalRoundtrip(t *testing.T) {
	t.Run("JSON: flat object fields preserved", func(t *testing.T) {
		_, out := roundtrip(t, "application/json", `{"name":"Alice","score":99}`)
		if !strings.Contains(out, `"name"`) || !strings.Contains(out, `"Alice"`) {
			t.Errorf("round-trip JSON missing name field: %q", out)
		}
		if !strings.Contains(out, `"score"`) || !strings.Contains(out, `99`) {
			t.Errorf("round-trip JSON missing score field: %q", out)
		}
	})

	t.Run("JSON: nested object fields preserved", func(t *testing.T) {
		_, out := roundtrip(t, "application/json", `{"outer":{"inner":42}}`)
		if !strings.Contains(out, `"outer"`) || !strings.Contains(out, `"inner"`) || !strings.Contains(out, `42`) {
			t.Errorf("round-trip JSON missing nested fields: %q", out)
		}
	})

	t.Run("YAML: flat key-value preserved", func(t *testing.T) {
		_, out := roundtrip(t, "application/yaml", "key: value\n")
		if !strings.Contains(out, "key: value") {
			t.Errorf("round-trip YAML missing 'key: value': %q", out)
		}
	})

	t.Run("YAML: nested map preserved", func(t *testing.T) {
		_, out := roundtrip(t, "application/yaml", "outer:\n  inner: 42\n")
		if !strings.Contains(out, "outer") || !strings.Contains(out, "inner") || !strings.Contains(out, "42") {
			t.Errorf("round-trip YAML missing nested fields: %q", out)
		}
	})

	t.Run("XML: element and text content preserved", func(t *testing.T) {
		_, out := roundtrip(t, "application/xml", `<root><child>hello</child></root>`)
		if !strings.Contains(out, "root") || !strings.Contains(out, "child") || !strings.Contains(out, "hello") {
			t.Errorf("round-trip XML missing expected content: %q", out)
		}
	})

	t.Run("XML: all attributes preserved", func(t *testing.T) {
		_, out := roundtrip(t, "application/xml", `<item id="1" name="test"/>`)
		if !strings.Contains(out, "item") {
			t.Errorf("round-trip XML missing 'item': %q", out)
		}
		if !strings.Contains(out, `id="1"`) {
			t.Errorf("round-trip XML missing attribute id: %q", out)
		}
		if !strings.Contains(out, `name="test"`) {
			t.Errorf("round-trip XML missing attribute name: %q", out)
		}
	})
}
