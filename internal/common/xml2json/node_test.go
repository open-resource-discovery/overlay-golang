//go:build unit

package xml2json

import (
	"testing"
)

func TestNewTextNode(t *testing.T) {
	t.Run("type and value stored correctly", func(t *testing.T) {
		n := NewTextNode("hello world")
		if n.Type() != "text" {
			t.Errorf("Type() = %q, want \"text\"", n.Type())
		}
		if n.Value() != "hello world" {
			t.Errorf("Value() = %q, want \"hello world\"", n.Value())
		}
	})

	t.Run("empty string value is stored", func(t *testing.T) {
		n := NewTextNode("")
		if n.Value() != "" {
			t.Errorf("Value() = %q, want \"\"", n.Value())
		}
	})

	t.Run("value with special characters stored verbatim (no escaping at storage)", func(t *testing.T) {
		n := NewTextNode("a & b < c > d")
		if n.Value() != "a & b < c > d" {
			t.Errorf("Value() = %q", n.Value())
		}
	})
}

func TestNewCdataNode(t *testing.T) {
	t.Run("type and value stored correctly", func(t *testing.T) {
		n := NewCdataNode("raw <data>")
		if n.Type() != "cdata" {
			t.Errorf("Type() = %q, want \"cdata\"", n.Type())
		}
		if n.Value() != "raw <data>" {
			t.Errorf("Value() = %q, want \"raw <data>\"", n.Value())
		}
	})

	t.Run("empty string value is stored", func(t *testing.T) {
		n := NewCdataNode("")
		if n.Value() != "" {
			t.Errorf("Value() = %q, want \"\"", n.Value())
		}
	})
}

func TestNewProcessingInstructionNode(t *testing.T) {
	t.Run("type and value stored correctly", func(t *testing.T) {
		n := NewProcessingInstructionNode("<?target data?>")
		if n.Type() != "processing-instruction" {
			t.Errorf("Type() = %q, want \"processing-instruction\"", n.Type())
		}
		if n.Value() != "<?target data?>" {
			t.Errorf("Value() = %q, want \"<?target data?>\"", n.Value())
		}
	})

	t.Run("empty string value is stored", func(t *testing.T) {
		n := NewProcessingInstructionNode("")
		if n.Value() != "" {
			t.Errorf("Value() = %q, want \"\"", n.Value())
		}
	})
}

func TestNewElementNode(t *testing.T) {
	t.Run("type is element", func(t *testing.T) {
		n := NewElementNode("root", nil, NewAttributes())
		if n.Type() != "element" {
			t.Errorf("Type() = %q, want \"element\"", n.Type())
		}
	})

	t.Run("name is stored correctly", func(t *testing.T) {
		n := NewElementNode("my-element", nil, NewAttributes())
		if n.Name() != "my-element" {
			t.Errorf("Name() = %q, want \"my-element\"", n.Name())
		}
	})

	t.Run("namespaced name stored as-is", func(t *testing.T) {
		n := NewElementNode("ns:tag", nil, NewAttributes())
		if n.Name() != "ns:tag" {
			t.Errorf("Name() = %q, want \"ns:tag\"", n.Name())
		}
	})

	t.Run("nil nodes slice stored as nil", func(t *testing.T) {
		n := NewElementNode("root", nil, NewAttributes())
		if n.Nodes() != nil {
			t.Errorf("Nodes() = %v, want nil", n.Nodes())
		}
	})

	t.Run("empty nodes slice stored as empty", func(t *testing.T) {
		n := NewElementNode("root", []Node{}, NewAttributes())
		if len(n.Nodes()) != 0 {
			t.Errorf("Nodes() len = %d, want 0", len(n.Nodes()))
		}
	})

	t.Run("child nodes stored in order", func(t *testing.T) {
		children := []Node{NewTextNode("a"), NewTextNode("b"), NewTextNode("c")}
		n := NewElementNode("root", children, NewAttributes())
		got := n.Nodes()
		if len(got) != 3 {
			t.Fatalf("Nodes() len = %d, want 3", len(got))
		}
		for i, want := range []string{"a", "b", "c"} {
			if got[i].Value() != want {
				t.Errorf("Nodes()[%d].Value() = %q, want %q", i, got[i].Value(), want)
			}
		}
	})

	t.Run("attributes stored and retrievable via Attributes()", func(t *testing.T) {
		attrs := NewAttributes("id", "42", "class", "main")
		n := NewElementNode("root", nil, attrs)
		if n.Attributes().Get("id") != "42" {
			t.Errorf("id = %q, want \"42\"", n.Attributes().Get("id"))
		}
		if n.Attributes().Get("class") != "main" {
			t.Errorf("class = %q, want \"main\"", n.Attributes().Get("class"))
		}
	})

	t.Run("empty attributes stored as empty", func(t *testing.T) {
		n := NewElementNode("root", nil, NewAttributes())
		if len(n.Attributes()) != 0 {
			t.Errorf("Attributes() len = %d, want 0", len(n.Attributes()))
		}
	})
}

func TestNodeAttribute(t *testing.T) {
	t.Run("returns value for present attribute", func(t *testing.T) {
		n := NewElementNode("root", nil, NewAttributes("key", "value"))
		if got := n.Attribute("key"); got != "value" {
			t.Errorf("Attribute(\"key\") = %q, want \"value\"", got)
		}
	})

	t.Run("absent attribute returns empty string", func(t *testing.T) {
		n := NewElementNode("root", nil, NewAttributes())
		if got := n.Attribute("missing"); got != "" {
			t.Errorf("Attribute(\"missing\") = %q, want \"\"", got)
		}
	})

	t.Run("attribute present with empty value returns empty string", func(t *testing.T) {
		// Differs from absent: Has() would return true, but Attribute() still returns "".
		n := NewElementNode("root", nil, NewAttributes("blank", ""))
		if got := n.Attribute("blank"); got != "" {
			t.Errorf("Attribute(\"blank\") = %q, want \"\"", got)
		}
		if !n.Attributes().Has("blank") {
			t.Error("Has(\"blank\") should be true")
		}
	})
}

func TestNodeTypeDistinctness(t *testing.T) {
	cases := []struct {
		name string
		node Node
		want string
	}{
		{"text", NewTextNode("x"), "text"},
		{"cdata", NewCdataNode("x"), "cdata"},
		{"processing-instruction", NewProcessingInstructionNode("x"), "processing-instruction"},
		{"element", NewElementNode("x", nil, NewAttributes()), "element"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.node.Type() != tc.want {
				t.Errorf("Type() = %q, want %q", tc.node.Type(), tc.want)
			}
		})
	}
}
