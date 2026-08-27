//go:build unit

package xml2json

import (
	"strings"
	"testing"
)

// helpers

func makeDoc(nodes []Node, notations []string, declarations []string) Document {
	if notations == nil {
		notations = []string{}
	}
	if declarations == nil {
		declarations = []string{}
	}
	return NewDocument(nodes, notations, declarations)
}

func elem(name string, children []Node, attrs ...string) Node {
	return NewElementNode(name, children, NewAttributes(attrs...))
}

func text(value string) Node {
	return NewTextNode(value)
}

func cdata(value string) Node {
	return NewCdataNode(value)
}

func pi(value string) Node {
	return NewProcessingInstructionNode(value)
}

// TestNewDocument

func TestNewDocument(t *testing.T) {
	t.Run("stores and returns nodes", func(t *testing.T) {
		nodes := []Node{elem("root", nil)}
		doc := makeDoc(nodes, nil, nil)
		if len(doc.Nodes()) != 1 || doc.Nodes()[0].Name() != "root" {
			t.Errorf("Nodes() = %v", doc.Nodes())
		}
	})

	t.Run("stores and returns notations", func(t *testing.T) {
		doc := makeDoc(nil, []string{"<!NOTATION foo SYSTEM \"bar\">"}, nil)
		if len(doc.Notations()) != 1 || doc.Notations()[0] != "<!NOTATION foo SYSTEM \"bar\">" {
			t.Errorf("Notations() = %v", doc.Notations())
		}
	})

	t.Run("stores and returns declarations", func(t *testing.T) {
		doc := makeDoc(nil, nil, []string{`<?xml version="1.0"?>`})
		if len(doc.Declarations()) != 1 || doc.Declarations()[0] != `<?xml version="1.0"?>` {
			t.Errorf("Declarations() = %v", doc.Declarations())
		}
	})

	t.Run("empty slices round-trip as empty", func(t *testing.T) {
		doc := makeDoc([]Node{}, []string{}, []string{})
		if len(doc.Nodes()) != 0 {
			t.Errorf("Nodes len = %d, want 0", len(doc.Nodes()))
		}
		if len(doc.Notations()) != 0 {
			t.Errorf("Notations len = %d, want 0", len(doc.Notations()))
		}
		if len(doc.Declarations()) != 0 {
			t.Errorf("Declarations len = %d, want 0", len(doc.Declarations()))
		}
	})

	t.Run("multiple nodes stored in order", func(t *testing.T) {
		nodes := []Node{elem("a", nil), elem("b", nil), elem("c", nil)}
		doc := makeDoc(nodes, nil, nil)
		got := doc.Nodes()
		if len(got) != 3 || got[0].Name() != "a" || got[1].Name() != "b" || got[2].Name() != "c" {
			t.Errorf("Nodes() = %v", got)
		}
	})
}

// TestDocumentToXML_EmptyAndDeclarations

func TestDocumentToXML_EmptyAndDeclarations(t *testing.T) {
	t.Run("empty document produces empty string", func(t *testing.T) {
		doc := makeDoc([]Node{}, nil, nil)
		if got := doc.ToXML(); got != "" {
			t.Errorf("ToXML() = %q, want \"\"", got)
		}
	})

	t.Run("declaration appears before element output", func(t *testing.T) {
		doc := makeDoc(
			[]Node{elem("root", nil)},
			nil,
			[]string{`<?xml version="1.0"?>`},
		)
		out := doc.ToXML()
		declIdx := strings.Index(out, "<?xml")
		rootIdx := strings.Index(out, "<root")
		if declIdx == -1 || rootIdx == -1 {
			t.Fatalf("ToXML() = %q — missing declaration or root", out)
		}
		if declIdx > rootIdx {
			t.Errorf("declaration appears after root element")
		}
	})

	t.Run("multiple declarations joined without separator", func(t *testing.T) {
		doc := makeDoc(
			[]Node{},
			nil,
			[]string{`<?xml version="1.0"?>`, `<?xml-stylesheet type="text/xsl" href="style.xsl"?>`},
		)
		out := doc.ToXML()
		if !strings.Contains(out, "xml-stylesheet") {
			t.Errorf("ToXML() = %q, expected xml-stylesheet", out)
		}
	})
}

// TestDocumentToXML_SelfClosingElement

func TestDocumentToXML_SelfClosingElement(t *testing.T) {
	t.Run("element with no children renders as self-closing", func(t *testing.T) {
		doc := makeDoc([]Node{elem("item", nil)}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "<item/>") {
			t.Errorf("ToXML() = %q, expected self-closing <item/>", out)
		}
	})

	t.Run("element with no children and one attribute renders self-closing with attribute", func(t *testing.T) {
		doc := makeDoc([]Node{elem("item", nil, "id", "1")}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, `id="1"`) {
			t.Errorf("ToXML() = %q, expected id=\"1\"", out)
		}
		if strings.Contains(out, "</item>") {
			t.Errorf("ToXML() = %q, expected self-closing not close tag", out)
		}
	})
}

// TestDocumentToXML_ElementWithText

func TestDocumentToXML_ElementWithText(t *testing.T) {
	t.Run("element with single text child renders inline", func(t *testing.T) {
		doc := makeDoc([]Node{elem("title", []Node{text("hello")})}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "<title") || !strings.Contains(out, "hello") || !strings.Contains(out, "</title>") {
			t.Errorf("ToXML() = %q", out)
		}
		// inline: open tag, text, close tag should all be on one line
		if strings.Count(out, "title") < 2 {
			t.Errorf("expected open and close title tags, got %q", out)
		}
	})

	t.Run("text value has special characters escaped", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", []Node{text("a & b < c")})}, nil, nil)
		out := doc.ToXML()
		if strings.Contains(out, "a & b") {
			t.Errorf("ToXML() = %q — unescaped & found", out)
		}
		if !strings.Contains(out, "a &amp; b") {
			t.Errorf("ToXML() = %q — expected &amp;", out)
		}
		if !strings.Contains(out, "&lt;") {
			t.Errorf("ToXML() = %q — expected &lt;", out)
		}
	})

	t.Run("text value is trimmed of surrounding whitespace in output", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", []Node{text("  hello  ")})}, nil, nil)
		out := doc.ToXML()
		if strings.Contains(out, "  hello  ") {
			t.Errorf("ToXML() = %q — expected trimmed text", out)
		}
		if !strings.Contains(out, "hello") {
			t.Errorf("ToXML() = %q — expected hello", out)
		}
	})
}

// TestDocumentToXML_NestedElements

func TestDocumentToXML_NestedElements(t *testing.T) {
	t.Run("element with element child renders with open and close tags", func(t *testing.T) {
		doc := makeDoc([]Node{
			elem("outer", []Node{elem("inner", nil)}),
		}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "<outer>") || !strings.Contains(out, "</outer>") {
			t.Errorf("ToXML() = %q — missing outer tags", out)
		}
		if !strings.Contains(out, "<inner/>") {
			t.Errorf("ToXML() = %q — missing inner self-closing", out)
		}
	})

	t.Run("multiple children render in order", func(t *testing.T) {
		doc := makeDoc([]Node{
			elem("root", []Node{elem("first", nil), elem("second", nil)}),
		}, nil, nil)
		out := doc.ToXML()
		firstIdx := strings.Index(out, "first")
		secondIdx := strings.Index(out, "second")
		if firstIdx == -1 || secondIdx == -1 || firstIdx > secondIdx {
			t.Errorf("ToXML() = %q — first not before second", out)
		}
	})

	t.Run("three levels of nesting render correctly", func(t *testing.T) {
		doc := makeDoc([]Node{
			elem("l1", []Node{
				elem("l2", []Node{
					elem("l3", nil),
				}),
			}),
		}, nil, nil)
		out := doc.ToXML()
		for _, want := range []string{"l1", "l2", "l3"} {
			if !strings.Contains(out, want) {
				t.Errorf("ToXML() = %q — missing %q", out, want)
			}
		}
	})
}

// TestDocumentToXML_Attributes

func TestDocumentToXML_Attributes(t *testing.T) {
	t.Run("attributes appear in output", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", nil, "id", "42", "name", "test")}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, `id="42"`) {
			t.Errorf("ToXML() = %q — missing id", out)
		}
		if !strings.Contains(out, `name="test"`) {
			t.Errorf("ToXML() = %q — missing name", out)
		}
	})

	t.Run("attribute value with & is escaped to &amp;", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", nil, "title", "a & b")}, nil, nil)
		out := doc.ToXML()
		if strings.Contains(out, `title="a & b"`) {
			t.Errorf("ToXML() = %q — unescaped & in attribute", out)
		}
		if !strings.Contains(out, `title="a &amp; b"`) {
			t.Errorf("ToXML() = %q — expected &amp; in attribute", out)
		}
	})

	t.Run("attribute value with < is escaped to &lt;", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", nil, "val", "a < b")}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "&lt;") {
			t.Errorf("ToXML() = %q — expected &lt;", out)
		}
	})

	t.Run("attributes are rendered in sorted key order", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", nil, "z", "last", "a", "first")}, nil, nil)
		out := doc.ToXML()
		aIdx := strings.Index(out, `a="first"`)
		zIdx := strings.Index(out, `z="last"`)
		if aIdx == -1 || zIdx == -1 || aIdx > zIdx {
			t.Errorf("ToXML() = %q — attributes not sorted: a should precede z", out)
		}
	})

	t.Run("attribute value with quote is escaped to &quot;", func(t *testing.T) {
		doc := makeDoc([]Node{elem("root", nil, "msg", `say "hi"`)}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "&quot;") {
			t.Errorf("ToXML() = %q — expected &quot;", out)
		}
	})
}

// TestDocumentToXML_CdataAndPI

func TestDocumentToXML_CdataAndPI(t *testing.T) {
	t.Run("cdata node renders as CDATA section", func(t *testing.T) {
		doc := makeDoc([]Node{
			elem("root", []Node{cdata("raw <content>")}),
		}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "<![CDATA[raw <content>]]>") {
			t.Errorf("ToXML() = %q — expected CDATA section", out)
		}
	})

	t.Run("processing-instruction node renders verbatim", func(t *testing.T) {
		doc := makeDoc([]Node{
			elem("root", []Node{pi("<?target data?>")}),
		}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "<?target data?>") {
			t.Errorf("ToXML() = %q — expected PI", out)
		}
	})
}

// TestDocumentToXML_TextAtDocumentLevel

func TestDocumentToXML_TextAtDocumentLevel(t *testing.T) {
	t.Run("top-level text node renders escaped", func(t *testing.T) {
		doc := makeDoc([]Node{text("hello & world")}, nil, nil)
		out := doc.ToXML()
		if !strings.Contains(out, "&amp;") {
			t.Errorf("ToXML() = %q — expected &amp;", out)
		}
		if !strings.Contains(out, "hello") {
			t.Errorf("ToXML() = %q — expected hello", out)
		}
	})
}
