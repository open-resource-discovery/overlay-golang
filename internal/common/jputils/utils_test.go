//go:build unit

package jputils

import (
	"slices"
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
)

// fragType returns a string label for the concrete type of a jp.Frag.
func fragType(f jp.Frag) string {
	switch f.(type) {
	case jp.Root:
		return "Root"
	case jp.At:
		return "At"
	case jp.Wildcard:
		return "Wildcard"
	case jp.Child:
		return "Child"
	case jp.Nth:
		return "Nth"
	case *jp.Filter:
		return "Filter"
	default:
		return "unknown"
	}
}

// ---- Parse ------------------------------------------------------------------

func TestParse_RootExpression_ReturnsNoError(t *testing.T) {
	expr, err := Parse("$")
	if err != nil {
		t.Fatalf("Parse(\"$\") error = %v, want nil", err)
	}
	if len(expr) == 0 {
		t.Error("Parse(\"$\"): returned empty expression")
	}
}

func TestParse_RootExpression_EqualsRoot(t *testing.T) {
	expr, _ := Parse("$")
	testutils.AssertExpr(t, expr, "$")
}

func TestParse_SimpleChildPath_ReturnsNoError(t *testing.T) {
	_, err := Parse("$.info")
	if err != nil {
		t.Fatalf("Parse(\"$.info\") error = %v, want nil", err)
	}
}

func TestParse_SimpleChildPath_ExpressionString(t *testing.T) {
	expr, _ := Parse("$.info")
	testutils.AssertExpr(t, expr, "$.info")
}

func TestParse_NestedPath_ReturnsCorrectExpression(t *testing.T) {
	expr, err := Parse("$.info.title")
	if err != nil {
		t.Fatalf("Parse(\"$.info.title\") error = %v, want nil", err)
	}
	testutils.AssertExpr(t, expr, "$.info.title")
}

func TestParse_WildcardPath_ReturnsNoError(t *testing.T) {
	expr, err := Parse("$.*")
	if err != nil {
		t.Fatalf("Parse(\"$.*\") error = %v, want nil", err)
	}
	if len(expr) == 0 {
		t.Error("Parse(\"$.*\"): returned empty expression")
	}
}

func TestParse_ArrayIndexPath_ReturnsNoError(t *testing.T) {
	expr, err := Parse("$.items[0]")
	if err != nil {
		t.Fatalf("Parse(\"$.items[0]\") error = %v, want nil", err)
	}
	if len(expr) == 0 {
		t.Error("Parse(\"$.items[0]\"): returned empty expression")
	}
}

func TestParse_RecursiveDescentPath_ReturnsNoError(t *testing.T) {
	_, err := Parse("$..name")
	if err != nil {
		t.Fatalf("Parse(\"$..name\") error = %v, want nil", err)
	}
}

func TestParse_InvalidSyntax_ReturnsWarningError(t *testing.T) {
	_, err := Parse("$$[invalid")
	if err == nil {
		t.Fatal("expected error for invalid JSONPath syntax, got nil")
	}
	if err.Severity() != 0 { // Severity_Warning == 0
		t.Errorf("Severity() = %v, want Severity_Warning", err.Severity())
	}
}

func TestParse_ValidExpression_ReturnsNilError(t *testing.T) {
	cases := []string{
		"$",
		"$.foo",
		"$.foo.bar",
		"$.*",
		"$.items[0]",
		"$..name",
		"$.paths['/pets']",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(input)
			if err != nil {
				t.Errorf("Parse(%q) error = %v, want nil", input, err)
			}
		})
	}
}

func TestParse_ReturnedExpr_NavigatesDocument(t *testing.T) {
	doc := map[string]any{
		"info": map[string]any{"title": "Petstore"},
	}
	expr, err := Parse("$.info.title")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	result := expr.Get(doc)
	if len(result) != 1 || result[0] != "Petstore" {
		t.Errorf("Get = %v, want [Petstore]", result)
	}
}

func TestParse_EmptyString_ReturnsEmptyExprNoError(t *testing.T) {
	// jp.ParseString("") succeeds and returns an empty (zero-length) expression.
	expr, err := Parse("")
	if err != nil {
		t.Fatalf("Parse(\"\") error = %v, want nil", err)
	}
	if len(expr) != 0 {
		t.Errorf("Parse(\"\") len = %d, want 0", len(expr))
	}
}

// ---- Root -------------------------------------------------------------------

func TestRoot(t *testing.T) {
	t.Run("single Root frag that evaluates to the document", func(t *testing.T) {
		got := Root()
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if fragType(got[0]) != "Root" {
			t.Errorf("fragment type = %s, want Root", fragType(got[0]))
		}
		data := map[string]any{"x": 1}
		result := got.Get(data)
		if len(result) != 1 {
			t.Fatalf("Get len = %d, want 1", len(result))
		}
		m, ok := result[0].(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result[0])
		}
		if m["x"] != 1 {
			t.Errorf("x: got %v, want 1", m["x"])
		}
	})
}

func TestFrag(t *testing.T) {
	// Table for the three special-token cases.
	special := []struct {
		input    string
		wantType string
	}{
		{"@", "At"},
		{"$", "Root"},
		{"*", "Wildcard"},
	}
	for _, tc := range special {
		t.Run(tc.input+" returns "+tc.wantType, func(t *testing.T) {
			got := Frag(tc.input)
			if fragType(got) != tc.wantType {
				t.Errorf("Frag(%q): got %s, want %s", tc.input, fragType(got), tc.wantType)
			}
		})
	}

	t.Run("plain name returns Child with correct value", func(t *testing.T) {
		got := Frag("name")
		if fragType(got) != "Child" {
			t.Fatalf("got %s, want Child", fragType(got))
		}
		if string(got.(jp.Child)) != "name" {
			t.Errorf("child value = %q, want \"name\"", string(got.(jp.Child)))
		}
	})

	t.Run("name with hyphens returns Child", func(t *testing.T) {
		got := Frag("some-key")
		if fragType(got) != "Child" {
			t.Errorf("got %s, want Child", fragType(got))
		}
		if string(got.(jp.Child)) != "some-key" {
			t.Errorf("child value = %q, want \"some-key\"", string(got.(jp.Child)))
		}
	})

	t.Run("bare integer string returns Child, not Nth", func(t *testing.T) {
		// "1" does not match ^\[\d+]$ so falls through to Child.
		got := Frag("1")
		if fragType(got) != "Child" {
			t.Errorf("got %s, want Child", fragType(got))
		}
	})

	// Frag("[n]") matches the ^\[\d+]$ regex, but strconv.Atoi("[n]") always
	// fails (brackets are not valid for Atoi) and returns (0, err).
	// utils.First discards the error so jp.Nth(0) is always produced.
	// This is a bug in the implementation; the test acts as a regression guard.
	t.Run("[n] bracket form: regex matches but Atoi fails, always produces Nth(0)", func(t *testing.T) {
		for _, input := range []string{"[0]", "[1]", "[2]", "[99]"} {
			got := Frag(input)
			if fragType(got) != "Nth" {
				t.Errorf("Frag(%q): got %s, want Nth", input, fragType(got))
				continue
			}
			if int(got.(jp.Nth)) != 0 {
				t.Errorf("Frag(%q): nth = %d, want 0 (Atoi fails on bracketed form)", input, int(got.(jp.Nth)))
			}
		}
	})
}

func TestExpr(t *testing.T) {
	data := map[string]any{
		"name":  "Alice",
		"score": int64(99),
		"addr":  map[string]any{"city": "Berlin"},
		"items": []any{"a", "b", "c"},
	}

	t.Run("empty args returns empty Expr", func(t *testing.T) {
		if len(Expr()) != 0 {
			t.Errorf("len = %d, want 0", len(Expr()))
		}
	})

	t.Run("string fragments navigate to top-level field", func(t *testing.T) {
		result := Expr("$", "name").Get(data)
		if len(result) != 1 || result[0] != "Alice" {
			t.Errorf("got %v, want [Alice]", result)
		}
	})

	t.Run("string fragments navigate nested field", func(t *testing.T) {
		result := Expr("$", "addr", "city").Get(data)
		if len(result) != 1 || result[0] != "Berlin" {
			t.Errorf("got %v, want [Berlin]", result)
		}
	})

	t.Run("wildcard returns all top-level values", func(t *testing.T) {
		result := Expr("$", "*").Get(data)
		if len(result) != 4 {
			t.Errorf("len = %d, want 4", len(result))
		}
		// Confirm known scalar values are present (maps/slices are not hashable).
		foundAlice, foundScore := false, false
		for _, v := range result {
			if v == "Alice" {
				foundAlice = true
			}
			if v == int64(99) {
				foundScore = true
			}
		}
		if !foundAlice {
			t.Error("wildcard result missing \"Alice\"")
		}
		if !foundScore {
			t.Error("wildcard result missing int64(99)")
		}
	})

	t.Run("jp.Frag argument appended directly", func(t *testing.T) {
		result := Expr(jp.Root('$'), jp.Child("score")).Get(data)
		if len(result) != 1 || result[0] != int64(99) {
			t.Errorf("got %v, want [99]", result)
		}
	})

	t.Run("jp.Expr argument is flattened into result", func(t *testing.T) {
		sub := jp.Expr{jp.Root('$'), jp.Child("name")}
		result := Expr(sub).Get(data)
		if len(result) != 1 || result[0] != "Alice" {
			t.Errorf("got %v, want [Alice]", result)
		}
	})

	t.Run("jp.Expr argument combined with trailing string fragments", func(t *testing.T) {
		// Demonstrates that a jp.Expr is flattened and string frags appended after.
		sub := jp.Expr{jp.Root('$'), jp.Child("addr")}
		result := Expr(sub, "city").Get(data)
		if len(result) != 1 || result[0] != "Berlin" {
			t.Errorf("got %v, want [Berlin]", result)
		}
	})

	t.Run("*jp.Equation argument becomes Filter frag on array root", func(t *testing.T) {
		arr := []any{
			map[string]any{"name": "Alice"},
			map[string]any{"name": "Bob"},
		}
		eq := Eq("@.name", "Alice")
		result := Expr("$", eq).Get(arr)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		if result[0].(map[string]any)["name"] != "Alice" {
			t.Errorf("name = %v, want Alice", result[0].(map[string]any)["name"])
		}
	})

	t.Run("fragment types assembled in correct order", func(t *testing.T) {
		e := Expr("$", "items", "*")
		if len(e) != 3 {
			t.Fatalf("len = %d, want 3", len(e))
		}
		want := []string{"Root", "Child", "Wildcard"}
		got := make([]string, len(e))
		for i, f := range e {
			got[i] = fragType(f)
		}
		if !slices.Equal(got, want) {
			t.Errorf("fragment types = %v, want %v", got, want)
		}
	})
}

func TestConst(t *testing.T) {
	t.Run("all supported types return non-nil Equation", func(t *testing.T) {
		cases := []struct {
			name  string
			value any
		}{
			{"bool true", true},
			{"bool false", false},
			{"int64", int64(42)},
			{"string", "hello"},
			{"float64", float64(3.14)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if Const(tc.value) == nil {
					t.Errorf("Const(%v): got nil, want non-nil Equation", tc.value)
				}
			})
		}
	})

	t.Run("string constant evaluates correctly in a filter", func(t *testing.T) {
		// Verify the Equation is not just non-nil but holds the right constant.
		arr := []any{
			map[string]any{"tag": "yes"},
			map[string]any{"tag": "no"},
		}
		filter := jp.Eq(jp.Get(Expr("@", "tag")), Const("yes"))
		result := Expr("$", filter).Get(arr)
		if len(result) != 1 || result[0].(map[string]any)["tag"] != "yes" {
			t.Errorf("got %v, want [{tag:yes}]", result)
		}
	})

	t.Run("bool constant evaluates correctly in a filter", func(t *testing.T) {
		arr := []any{
			map[string]any{"active": true},
			map[string]any{"active": false},
		}
		filter := jp.Eq(jp.Get(Expr("@", "active")), Const(true))
		result := Expr("$", filter).Get(arr)
		if len(result) != 1 || result[0].(map[string]any)["active"] != true {
			t.Errorf("got %v, want [{active:true}]", result)
		}
	})

	t.Run("unsupported type panics", func(t *testing.T) {
		defer testutils.AssertPanics(t, "expected panic due to unsupported type")

		Const([]int{1, 2, 3})
	})
}

func TestEq(t *testing.T) {
	arr := []any{
		map[string]any{"name": "Alice", "age": int64(30)},
		map[string]any{"name": "Bob", "age": int64(25)},
		map[string]any{"name": "Alice", "age": int64(40)},
	}

	t.Run("filters array by string equality", func(t *testing.T) {
		result := Expr("$", Eq("@.name", "Alice")).Get(arr)
		if len(result) != 2 {
			t.Fatalf("len = %d, want 2", len(result))
		}
		for i, r := range result {
			if r.(map[string]any)["name"] != "Alice" {
				t.Errorf("result[%d]: name = %v, want Alice", i, r.(map[string]any)["name"])
			}
		}
	})

	t.Run("filters array by int64 equality", func(t *testing.T) {
		result := Expr("$", Eq("@.age", int64(25))).Get(arr)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		m := result[0].(map[string]any)
		if m["name"] != "Bob" || m["age"] != int64(25) {
			t.Errorf("got %v, want {name:Bob age:25}", m)
		}
	})

	t.Run("no match returns empty slice", func(t *testing.T) {
		result := Expr("$", Eq("@.name", "Charlie")).Get(arr)
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("dot-separated path navigates nested field", func(t *testing.T) {
		nested := []any{
			map[string]any{"addr": map[string]any{"city": "Berlin"}},
			map[string]any{"addr": map[string]any{"city": "Paris"}},
		}
		result := Expr("$", Eq("@.addr.city", "Berlin")).Get(nested)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		city := result[0].(map[string]any)["addr"].(map[string]any)["city"]
		if city != "Berlin" {
			t.Errorf("city = %v, want Berlin", city)
		}
	})

	t.Run("single-segment path (no dot) works", func(t *testing.T) {
		// strings.Split("@", ".") produces ["@"], a one-element path.
		flat := []any{"Alice", "Bob"}
		result := Expr("$", Eq("@", "Alice")).Get(flat)
		if len(result) != 1 || result[0] != "Alice" {
			t.Errorf("got %v, want [Alice]", result)
		}
	})

	t.Run("returns non-nil Equation", func(t *testing.T) {
		if Eq("@.x", "v") == nil {
			t.Error("expected non-nil equation")
		}
	})
}

func TestIsRoot(t *testing.T) {
	t.Run("Root() expression is root", func(t *testing.T) {
		if !IsRoot(Root()) {
			t.Error("IsRoot(Root()): got false, want true")
		}
	})

	t.Run("single Root frag is root", func(t *testing.T) {
		if !IsRoot(jp.Expr{jp.Root('$')}) {
			t.Error("single Root frag: got false, want true")
		}
	})

	t.Run("empty expression is not root", func(t *testing.T) {
		if IsRoot(jp.Expr{}) {
			t.Error("empty expr: got true, want false")
		}
	})

	t.Run("multi-frag expression is not root", func(t *testing.T) {
		if IsRoot(jp.Expr{jp.Root('$'), jp.Child("x")}) {
			t.Error("multi-frag expr: got true, want false")
		}
	})

	t.Run("single non-Root frag is not root", func(t *testing.T) {
		if IsRoot(jp.Expr{jp.Child("x")}) {
			t.Error("Child frag: got true, want false")
		}
	})
}

func TestPinpoint(t *testing.T) {
	data := map[string]any{
		"name": "Alice",
		"addr": map[string]any{"city": "Berlin"},
	}

	t.Run("root expression returns expression as-is", func(t *testing.T) {
		expr := Root()
		loc, exists, err := Pinpoint(data, expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("exists: got false, want true")
		}
		if loc.String() != expr.String() {
			t.Errorf("loc = %q, want %q", loc.String(), expr.String())
		}
	})

	t.Run("expression matching exactly one element returns location", func(t *testing.T) {
		expr := Expr("$", "name")
		loc, exists, err := Pinpoint(data, expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("exists: got false, want true")
		}
		if loc == nil {
			t.Error("loc: got nil, want non-nil")
		}
	})

	t.Run("expression matching nothing returns error and exists=false", func(t *testing.T) {
		expr := Expr("$", "missing")
		_, exists, err := Pinpoint(data, expr)
		if err == nil {
			t.Error("expected error for no-match, got nil")
		}
		if exists {
			t.Error("exists: got true, want false")
		}
	})

	t.Run("ambiguous expression returns error and exists=true", func(t *testing.T) {
		// Wildcard matches multiple elements.
		expr := Expr("$", "*")
		_, exists, err := Pinpoint(data, expr)
		if err == nil {
			t.Error("expected error for ambiguous match, got nil")
		}
		if !exists {
			t.Error("exists: got false, want true")
		}
	})
}

func TestAnd(t *testing.T) {
	arr := []any{
		map[string]any{"name": "Alice", "age": int64(30)},
		map[string]any{"name": "Alice", "age": int64(25)},
		map[string]any{"name": "Bob", "age": int64(30)},
	}

	t.Run("two conditions both satisfied", func(t *testing.T) {
		result := Expr("$", And(Eq("@.name", "Alice"), Eq("@.age", int64(30)))).Get(arr)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		m := result[0].(map[string]any)
		if m["name"] != "Alice" || m["age"] != int64(30) {
			t.Errorf("got %v, want {name:Alice age:30}", m)
		}
	})

	t.Run("only first condition satisfied — no match", func(t *testing.T) {
		result := Expr("$", And(Eq("@.name", "Alice"), Eq("@.age", int64(99)))).Get(arr)
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("neither condition satisfied — no match", func(t *testing.T) {
		result := Expr("$", And(Eq("@.name", "Charlie"), Eq("@.age", int64(30)))).Get(arr)
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("three conditions via rest variadics", func(t *testing.T) {
		data := []any{
			map[string]any{"a": "x", "b": "y", "c": "z"},
			map[string]any{"a": "x", "b": "y", "c": "w"},
			map[string]any{"a": "x", "b": "n", "c": "z"},
		}
		result := Expr("$", And(Eq("@.a", "x"), Eq("@.b", "y"), Eq("@.c", "z"))).Get(data)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		m := result[0].(map[string]any)
		if m["a"] != "x" || m["b"] != "y" || m["c"] != "z" {
			t.Errorf("got %v, want {a:x b:y c:z}", m)
		}
	})

	t.Run("four conditions via rest variadics", func(t *testing.T) {
		data := []any{
			map[string]any{"a": "1", "b": "2", "c": "3", "d": "4"},
			map[string]any{"a": "1", "b": "2", "c": "3", "d": "X"},
		}
		result := Expr("$", And(Eq("@.a", "1"), Eq("@.b", "2"), Eq("@.c", "3"), Eq("@.d", "4"))).Get(data)
		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		m := result[0].(map[string]any)
		if m["a"] != "1" || m["b"] != "2" || m["c"] != "3" || m["d"] != "4" {
			t.Errorf("got %v, want {a:1 b:2 c:3 d:4}", m)
		}
	})

	t.Run("returns non-nil Equation", func(t *testing.T) {
		if And(Eq("@.x", "a"), Eq("@.y", "b")) == nil {
			t.Error("expected non-nil equation")
		}
	})
}
