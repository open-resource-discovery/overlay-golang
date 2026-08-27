//go:build unit

package xml2json

import (
	"strings"
	"testing"
)

func requireConvert(t *testing.T, xml string) Document {
	t.Helper()
	doc, err := Convert(xml)
	if err != nil {
		t.Fatalf("Convert(%q) error: %v", xml, err)
	}
	return doc
}

func requireNodes(t *testing.T, doc Document, wantLen int) []Node {
	t.Helper()
	nodes := doc.Nodes()
	if len(nodes) != wantLen {
		t.Fatalf("nodes len = %d, want %d", len(nodes), wantLen)
	}
	return nodes
}

func TestConvert_InvalidXML(t *testing.T) {
	t.Run("malformed XML returns error", func(t *testing.T) {
		_, err := Convert(`<unclosed>`)
		if err == nil {
			t.Error("expected error for unclosed tag, got nil")
		}
	})

	t.Run("empty string returns error", func(t *testing.T) {
		_, err := Convert(``)
		if err == nil {
			t.Error("expected error for empty input, got nil")
		}
	})
}

func TestConvert_DocumentStructure(t *testing.T) {
	t.Run("empty element document has one node and no notations", func(t *testing.T) {
		// xmlquery always synthesises one implicit declaration node for the document
		// root even when the input has no explicit <?xml?> header, so we only assert
		// on notations being absent.
		doc := requireConvert(t, `<root/>`)
		requireNodes(t, doc, 1)
		if len(doc.Notations()) != 0 {
			t.Errorf("notations len = %d, want 0", len(doc.Notations()))
		}
	})

	t.Run("XML declaration is captured in declarations", func(t *testing.T) {
		doc := requireConvert(t, `<?xml version="1.0" encoding="UTF-8"?><root/>`)
		if len(doc.Declarations()) != 1 {
			t.Fatalf("declarations len = %d, want 1", len(doc.Declarations()))
		}
		if !strings.Contains(doc.Declarations()[0], "xml") {
			t.Errorf("declaration = %q, expected to contain \"xml\"", doc.Declarations()[0])
		}
	})
}

func TestConvert_ElementNodes(t *testing.T) {
	t.Run("self-closing element produces element node", func(t *testing.T) {
		doc := requireConvert(t, `<root/>`)
		nodes := requireNodes(t, doc, 1)
		if nodes[0].Type() != "element" {
			t.Errorf("type = %q, want \"element\"", nodes[0].Type())
		}
		if nodes[0].Name() != "root" {
			t.Errorf("name = %q, want \"root\"", nodes[0].Name())
		}
	})

	t.Run("element with namespace prefix uses colon-joined name", func(t *testing.T) {
		doc := requireConvert(t, `<ns:root xmlns:ns="http://example.com"/>`)
		nodes := requireNodes(t, doc, 1)
		if nodes[0].Name() != "ns:root" {
			t.Errorf("name = %q, want \"ns:root\"", nodes[0].Name())
		}
	})

	t.Run("element without namespace prefix uses local name only", func(t *testing.T) {
		doc := requireConvert(t, `<item/>`)
		nodes := requireNodes(t, doc, 1)
		if nodes[0].Name() != "item" {
			t.Errorf("name = %q, want \"item\"", nodes[0].Name())
		}
	})

	t.Run("sibling elements all appear as top-level nodes in order", func(t *testing.T) {
		doc := requireConvert(t, `<root><a/><b/><c/></root>`)
		nodes := requireNodes(t, doc, 1)
		children := nodes[0].Nodes()
		if len(children) != 3 {
			t.Fatalf("children len = %d, want 3", len(children))
		}
		for i, want := range []string{"a", "b", "c"} {
			if children[i].Name() != want {
				t.Errorf("child[%d] name = %q, want %q", i, children[i].Name(), want)
			}
		}
	})

	t.Run("nested elements produce nested nodes", func(t *testing.T) {
		doc := requireConvert(t, `<outer><inner/></outer>`)
		nodes := requireNodes(t, doc, 1)
		if nodes[0].Name() != "outer" {
			t.Fatalf("root name = %q, want \"outer\"", nodes[0].Name())
		}
		inner := nodes[0].Nodes()
		if len(inner) != 1 || inner[0].Name() != "inner" {
			t.Errorf("inner node: got %v", inner)
		}
	})

	t.Run("empty element has no child nodes", func(t *testing.T) {
		doc := requireConvert(t, `<root/>`)
		nodes := requireNodes(t, doc, 1)
		if len(nodes[0].Nodes()) != 0 {
			t.Errorf("expected no children, got %d", len(nodes[0].Nodes()))
		}
	})
}

func TestConvert_TextNodes(t *testing.T) {
	t.Run("element text content produces a text node child", func(t *testing.T) {
		doc := requireConvert(t, `<root>hello</root>`)
		nodes := requireNodes(t, doc, 1)
		children := nodes[0].Nodes()
		if len(children) != 1 {
			t.Fatalf("children len = %d, want 1", len(children))
		}
		if children[0].Type() != "text" {
			t.Errorf("type = %q, want \"text\"", children[0].Type())
		}
		if children[0].Value() != "hello" {
			t.Errorf("value = %q, want \"hello\"", children[0].Value())
		}
	})

	t.Run("whitespace-only text between elements is ignored", func(t *testing.T) {
		doc := requireConvert(t, `<root>  <child/>  </root>`)
		nodes := requireNodes(t, doc, 1)
		children := nodes[0].Nodes()
		if len(children) != 1 || children[0].Type() != "element" {
			t.Errorf("expected one element child, got %v", children)
		}
	})

	t.Run("non-empty text mixed with whitespace is preserved", func(t *testing.T) {
		doc := requireConvert(t, `<root>  hello  </root>`)
		nodes := requireNodes(t, doc, 1)
		children := nodes[0].Nodes()
		if len(children) != 1 || children[0].Type() != "text" {
			t.Fatalf("expected text node, got %v", children)
		}
		if !strings.Contains(children[0].Value(), "hello") {
			t.Errorf("value = %q, expected to contain \"hello\"", children[0].Value())
		}
	})
}

func TestConvert_CdataNodes(t *testing.T) {
	t.Run("CDATA section produces a cdata node", func(t *testing.T) {
		doc := requireConvert(t, `<root><![CDATA[some <data>]]></root>`)
		nodes := requireNodes(t, doc, 1)
		children := nodes[0].Nodes()
		if len(children) != 1 {
			t.Fatalf("children len = %d, want 1", len(children))
		}
		if children[0].Type() != "cdata" {
			t.Errorf("type = %q, want \"cdata\"", children[0].Type())
		}
		if children[0].Value() != "some <data>" {
			t.Errorf("value = %q, want \"some <data>\"", children[0].Value())
		}
	})

	t.Run("CDATA preserves special characters verbatim", func(t *testing.T) {
		doc := requireConvert(t, `<root><![CDATA[a & b < c > d]]></root>`)
		children := requireNodes(t, doc, 1)[0].Nodes()
		if len(children) != 1 || children[0].Type() != "cdata" {
			t.Fatalf("expected cdata node, got %v", children)
		}
		if children[0].Value() != "a & b < c > d" {
			t.Errorf("value = %q", children[0].Value())
		}
	})
}

func TestConvert_ProcessingInstructions(t *testing.T) {
	t.Run("processing instruction inside element produces processing-instruction node", func(t *testing.T) {
		doc := requireConvert(t, `<root><?target data?></root>`)
		nodes := requireNodes(t, doc, 1)
		children := nodes[0].Nodes()
		if len(children) != 1 {
			t.Fatalf("children len = %d, want 1", len(children))
		}
		if children[0].Type() != "processing-instruction" {
			t.Errorf("type = %q, want \"processing-instruction\"", children[0].Type())
		}
		if !strings.Contains(children[0].Value(), "target") {
			t.Errorf("value = %q, expected to contain \"target\"", children[0].Value())
		}
	})
}

func TestConvert_Attributes(t *testing.T) {
	t.Run("element attributes are captured", func(t *testing.T) {
		doc := requireConvert(t, `<root id="1" class="main"/>`)
		nodes := requireNodes(t, doc, 1)
		attrs := nodes[0].Attributes()
		if attrs.Get("id") != "1" {
			t.Errorf("id = %q, want \"1\"", attrs.Get("id"))
		}
		if attrs.Get("class") != "main" {
			t.Errorf("class = %q, want \"main\"", attrs.Get("class"))
		}
	})

	t.Run("element with no attributes has empty Attributes", func(t *testing.T) {
		doc := requireConvert(t, `<root/>`)
		nodes := requireNodes(t, doc, 1)
		if len(nodes[0].Attributes()) != 0 {
			t.Errorf("expected no attributes, got %v", nodes[0].Attributes())
		}
	})

	t.Run("namespaced attribute uses colon-joined name", func(t *testing.T) {
		doc := requireConvert(t, `<root xmlns:ns="http://example.com" ns:attr="val"/>`)
		nodes := requireNodes(t, doc, 1)
		attrs := nodes[0].Attributes()
		if attrs.Get("ns:attr") != "val" {
			t.Errorf("ns:attr = %q, want \"val\"", attrs.Get("ns:attr"))
		}
	})

	t.Run("attribute entity references are decoded by xmlquery", func(t *testing.T) {
		// xmlquery decodes &amp; to & before handing the value to Convert.
		doc := requireConvert(t, `<root title="a &amp; b"/>`)
		nodes := requireNodes(t, doc, 1)
		title := nodes[0].Attributes().Get("title")
		if title != "a & b" {
			t.Errorf("title = %q, want \"a & b\"", title)
		}
	})

	t.Run("Attribute helper on Node returns correct value", func(t *testing.T) {
		doc := requireConvert(t, `<root key="value"/>`)
		nodes := requireNodes(t, doc, 1)
		if got := nodes[0].Attribute("key"); got != "value" {
			t.Errorf("Attribute(\"key\") = %q, want \"value\"", got)
		}
	})
}

func TestConvert_Notations(t *testing.T) {
	t.Run("DOCTYPE directive is captured in notations", func(t *testing.T) {
		// xmlquery produces a NotationNode for <!DOCTYPE ...> only when an explicit
		// <?xml?> declaration precedes it (otherwise the element node is lost).
		doc := requireConvert(t, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE root>\n<root/>")
		if len(doc.Notations()) != 1 {
			t.Fatalf("notations len = %d, want 1", len(doc.Notations()))
		}
		if !strings.Contains(doc.Notations()[0], "DOCTYPE") {
			t.Errorf("notation = %q, expected to contain \"DOCTYPE\"", doc.Notations()[0])
		}
	})
}

func TestConvert_RoundTrip(t *testing.T) {
	t.Run("simple element round-trips with correct name", func(t *testing.T) {
		doc := requireConvert(t, `<root/>`)
		out := doc.ToXML()
		// Re-parse to verify structural fidelity, not just substring presence.
		doc2 := requireConvert(t, out)
		nodes := requireNodes(t, doc2, 1)
		if nodes[0].Name() != "root" {
			t.Errorf("round-trip name = %q, want \"root\"", nodes[0].Name())
		}
	})

	t.Run("element with text round-trips with correct child text", func(t *testing.T) {
		doc := requireConvert(t, `<root><child>hello</child></root>`)
		out := doc.ToXML()
		doc2 := requireConvert(t, out)
		nodes := requireNodes(t, doc2, 1)
		children := nodes[0].Nodes()
		if len(children) != 1 || children[0].Name() != "child" {
			t.Fatalf("child element not preserved: %v", children)
		}
		grandchildren := children[0].Nodes()
		if len(grandchildren) != 1 || grandchildren[0].Type() != "text" {
			t.Fatalf("text node not preserved: %v", grandchildren)
		}
		if grandchildren[0].Value() != "hello" {
			t.Errorf("text value = %q, want \"hello\"", grandchildren[0].Value())
		}
	})

	t.Run("element with attribute round-trips with correct attribute value", func(t *testing.T) {
		doc := requireConvert(t, `<root id="42"/>`)
		out := doc.ToXML()
		doc2 := requireConvert(t, out)
		nodes := requireNodes(t, doc2, 1)
		if got := nodes[0].Attribute("id"); got != "42" {
			t.Errorf("Attribute(\"id\") = %q, want \"42\"", got)
		}
	})
}
