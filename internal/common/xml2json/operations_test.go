//go:build unit

package xml2json

import (
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
)

// helpers

func docWithNodes(nodes ...Node) Document {
	return NewDocument(nodes, []string{}, []string{})
}

func rootExpr() jp.Expr {
	return jputils.Expr("$")
}

// TestPinpoint

func TestPinpoint(t *testing.T) {
	t.Run("locates existing top-level field", func(t *testing.T) {
		doc := docWithNodes(elem("root", nil))
		expr := jputils.Expr("$", "nodes")
		located, found, err := Pinpoint(doc, expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Error("found = false, want true")
		}
		if located == nil {
			t.Error("located = nil, want non-nil expr")
		}
	})

	t.Run("returns error and false when path does not exist", func(t *testing.T) {
		doc := docWithNodes()
		expr := jputils.Expr("$", "nonexistent", "deep")
		_, found, err := Pinpoint(doc, expr)
		if err == nil {
			t.Error("expected error for missing path, got nil")
		}
		if found {
			t.Error("found = true, want false")
		}
	})

	t.Run("returns true and error when path matches multiple locations", func(t *testing.T) {
		// Wildcard on the nodes array matches each element — more than one result.
		doc := docWithNodes(elem("a", nil), elem("b", nil))
		expr := jputils.Expr("$", "nodes", "*")
		_, found, err := Pinpoint(doc, expr)
		if err == nil {
			t.Error("expected error for ambiguous path, got nil")
		}
		if !found {
			t.Error("found = false, want true for ambiguous match")
		}
	})
}

// TestSetNodes

func TestSetNodes(t *testing.T) {
	t.Run("replaces all nodes with new set", func(t *testing.T) {
		doc := docWithNodes(elem("old", nil))
		updated, err := SetNodes(doc, rootExpr(), elem("new", nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		nodes := updated.Nodes()
		if len(nodes) != 1 {
			t.Fatalf("len = %d, want 1", len(nodes))
		}
		if nodes[0].Name() != "new" {
			t.Errorf("name = %q, want \"new\"", nodes[0].Name())
		}
	})

	t.Run("replaces nodes with multiple new nodes", func(t *testing.T) {
		doc := docWithNodes(elem("old", nil))
		updated, err := SetNodes(doc, rootExpr(), elem("a", nil), elem("b", nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		nodes := updated.Nodes()
		if len(nodes) != 2 {
			t.Fatalf("len = %d, want 2", len(nodes))
		}
		if nodes[0].Name() != "a" || nodes[1].Name() != "b" {
			t.Errorf("names = %q %q, want \"a\" \"b\"", nodes[0].Name(), nodes[1].Name())
		}
	})

	t.Run("replaces nodes with empty set", func(t *testing.T) {
		doc := docWithNodes(elem("old", nil))
		updated, err := SetNodes(doc, rootExpr())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.Nodes()) != 0 {
			t.Errorf("len = %d, want 0", len(updated.Nodes()))
		}
	})

	t.Run("returns original document and error for invalid expression", func(t *testing.T) {
		doc := docWithNodes(elem("root", nil))
		_, err := SetNodes(doc, jputils.Expr("$", "nonexistent"))
		if err == nil {
			t.Error("expected error for invalid expression, got nil")
		}
	})

	t.Run("sets nodes on nested element via child expression", func(t *testing.T) {
		child := elem("child", []Node{elem("old-leaf", nil)})
		doc := docWithNodes(elem("root", []Node{child}))

		// Target the nodes array of the nested child element.
		childExpr := jputils.Expr("$", "nodes", jputils.Frag("[0]"), "nodes", jputils.Frag("[0]"))
		updated, err := SetNodes(doc, childExpr, elem("new-leaf", nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		root := updated.Nodes()[0]
		nested := root.Nodes()[0].Nodes()
		if len(nested) != 1 || nested[0].Name() != "new-leaf" {
			t.Errorf("nested nodes = %v", nested)
		}
	})
}

// TestAppendNodes

func TestAppendNodes(t *testing.T) {
	t.Run("appends a node to existing nodes", func(t *testing.T) {
		doc := docWithNodes(elem("existing", nil))
		updated, err := AppendNodes(doc, rootExpr(), elem("appended", nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		nodes := updated.Nodes()
		if len(nodes) != 2 {
			t.Fatalf("len = %d, want 2", len(nodes))
		}
		if nodes[0].Name() != "existing" {
			t.Errorf("nodes[0] = %q, want \"existing\"", nodes[0].Name())
		}
		if nodes[1].Name() != "appended" {
			t.Errorf("nodes[1] = %q, want \"appended\"", nodes[1].Name())
		}
	})

	t.Run("appends multiple nodes preserving original order", func(t *testing.T) {
		doc := docWithNodes(elem("first", nil))
		updated, err := AppendNodes(doc, rootExpr(), elem("second", nil), elem("third", nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		nodes := updated.Nodes()
		if len(nodes) != 3 {
			t.Fatalf("len = %d, want 3", len(nodes))
		}
		names := []string{nodes[0].Name(), nodes[1].Name(), nodes[2].Name()}
		want := []string{"first", "second", "third"}
		for i, n := range want {
			if names[i] != n {
				t.Errorf("nodes[%d] = %q, want %q", i, names[i], n)
			}
		}
	})

	t.Run("appends to empty nodes slice", func(t *testing.T) {
		doc := docWithNodes()
		updated, err := AppendNodes(doc, rootExpr(), elem("new", nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		nodes := updated.Nodes()
		if len(nodes) != 1 || nodes[0].Name() != "new" {
			t.Errorf("nodes = %v", nodes)
		}
	})

	t.Run("returns error for invalid expression", func(t *testing.T) {
		doc := docWithNodes()
		_, err := AppendNodes(doc, jputils.Expr("$", "nonexistent"), elem("x", nil))
		if err == nil {
			t.Error("expected error for invalid expression, got nil")
		}
	})
}

// TestPruneNodes

func TestPruneNodes(t *testing.T) {
	t.Run("keeps nodes matching predicate", func(t *testing.T) {
		doc := docWithNodes(elem("keep", nil), elem("remove", nil), elem("keep", nil))
		updated, err := PruneNodes(doc, rootExpr(), func(n Node) bool {
			return n.Name() == "keep"
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		nodes := updated.Nodes()
		if len(nodes) != 2 {
			t.Fatalf("len = %d, want 2", len(nodes))
		}
		for _, n := range nodes {
			if n.Name() != "keep" {
				t.Errorf("unexpected node %q", n.Name())
			}
		}
	})

	t.Run("all nodes removed leaves empty nodes slice when deleteIfEmpty is false", func(t *testing.T) {
		doc := docWithNodes(elem("a", nil), elem("b", nil))
		updated, err := PruneNodes(doc, rootExpr(), func(n Node) bool {
			return false // remove all
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.Nodes()) != 0 {
			t.Errorf("len = %d, want 0", len(updated.Nodes()))
		}
	})

	t.Run("predicate keeps all nodes unchanged", func(t *testing.T) {
		doc := docWithNodes(elem("a", nil), elem("b", nil))
		updated, err := PruneNodes(doc, rootExpr(), func(n Node) bool {
			return true // keep all
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.Nodes()) != 2 {
			t.Errorf("len = %d, want 2", len(updated.Nodes()))
		}
	})

	t.Run("prune on empty nodes returns empty slice", func(t *testing.T) {
		doc := docWithNodes()
		updated, err := PruneNodes(doc, rootExpr(), func(n Node) bool {
			return true
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.Nodes()) != 0 {
			t.Errorf("len = %d, want 0", len(updated.Nodes()))
		}
	})

	t.Run("deleteIfEmpty=true removes the element itself when its nodes become empty", func(t *testing.T) {
		// PruneNodes with deleteIfEmpty removes the parent element from its
		// containing nodes array when the pruned result is empty.
		leaf := elem("leaf", []Node{elem("target", nil)})
		doc := docWithNodes(leaf)

		childExpr := jputils.Expr("$", "nodes", jputils.Frag("[0]"))
		updated, err := PruneNodes(doc, childExpr, func(n Node) bool {
			return false // prune all children of leaf → empty → deleteIfEmpty removes leaf
		}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The leaf element itself is removed from the document's top-level nodes.
		if len(updated.Nodes()) != 0 {
			t.Errorf("top-level nodes len = %d, want 0 (leaf removed)", len(updated.Nodes()))
		}
	})

	t.Run("deleteIfEmpty=false keeps empty nodes slice in place", func(t *testing.T) {
		doc := docWithNodes(elem("a", nil))
		updated, err := PruneNodes(doc, rootExpr(), func(n Node) bool {
			return false
		}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.Nodes()) != 0 {
			t.Errorf("len = %d, want 0 (empty kept)", len(updated.Nodes()))
		}
	})

	t.Run("returns error for invalid expression", func(t *testing.T) {
		doc := docWithNodes()
		_, err := PruneNodes(doc, jputils.Expr("$", "nonexistent"), func(n Node) bool { return true })
		if err == nil {
			t.Error("expected error for invalid expression, got nil")
		}
	})
}
