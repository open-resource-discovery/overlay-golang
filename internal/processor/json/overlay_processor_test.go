//go:build unit

package json

import (
	"reflect"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

var mcpDoc = testutils.UnmarshalFixture[map[string]any]("testdata/mcp_server_card.json")

// ---- helpers ----------------------------------------------------------------

// newProcessor creates an OverlayProcessor from a raw JSON string.
func newProcessor(t *testing.T, content string) *OverlayProcessor {
	t.Helper()
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   content,
		MediaType: "application/json",
	})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	return p
}

// mcpContent is the MCP Server Card fixture serialised as a string for use
// with NewOverlayProcessor (which requires a string, not a map).
var mcpContent = func() string {
	s, err := marshaller.Marshal("application/json", mcpDoc)
	if err != nil {
		panic(err)
	}
	return s
}()

// ---- NewOverlayProcessor ----------------------------------------------------

func TestNewOverlayProcessor_ValidJSON_Succeeds(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   `{"name":"test"}`,
		MediaType: "application/json",
	}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNewOverlayProcessor_InvalidJSON_ReturnsError(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   `{not valid json`,
		MediaType: "application/json",
	}); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// ---- Apply: output fields ---------------------------------------------------

func TestApply_OutputFields_PurposeAndVisibilitySet(t *testing.T) {
	p := newProcessor(t, mcpContent)
	rd, err := p.Apply(model.OverlayDefinition{
		Purpose: "test-purpose",
		Overlay: model.Overlay{
			Visibility: "public",
			Patches:    []model.Patch{},
		},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rd.Purpose != "test-purpose" {
		t.Errorf("Purpose: got %q, want %q", rd.Purpose, "test-purpose")
	}
	if rd.Visibility != "public" {
		t.Errorf("Visibility: got %q, want %q", rd.Visibility, "public")
	}
}

func TestApply_OutputFields_OriginalDefinitionFieldsPreserved(t *testing.T) {
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:    mcpContent,
		MediaType:  "application/json",
		OrdID:      "sap.sm:resourceDefinition:mcp-server-card:v1",
		URL:        "https://example.com/mcp-server-card.json",
		Visibility: "internal",
	})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	rd, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Visibility: "public", Patches: []model.Patch{}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rd.OrdID != "sap.sm:resourceDefinition:mcp-server-card:v1" {
		t.Errorf("OrdID: got %q, want original value", rd.OrdID)
	}
	if rd.URL != "https://example.com/mcp-server-card.json" {
		t.Errorf("URL: got %q, want original value", rd.URL)
	}
	// Visibility comes from the overlay, not the original definition.
	if rd.Visibility != "public" {
		t.Errorf("Visibility: got %q, want %q", rd.Visibility, "public")
	}
}

// ---- Apply: no patches / immutability ---------------------------------------

func TestApply_NoPatch_ContentUnchanged(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	if !reflect.DeepEqual(result["name"], mcpDoc["name"]) {
		t.Error("content changed without any patches")
	}
	if len(result["tools"].([]any)) != len(mcpDoc["tools"].([]any)) {
		t.Errorf("tools count changed: got %d, want %d",
			len(result["tools"].([]any)), len(mcpDoc["tools"].([]any)))
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	p := newProcessor(t, mcpContent)
	// First Apply modifies _meta.
	testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"mutated": true},
			},
		}},
	})
	// Second Apply with no patches must return the original document — internal
	// state must not have been modified by the first call.
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	meta := testutils.Get(t, result, "_meta").(map[string]any)
	if _, ok := meta["mutated"]; ok {
		t.Error("first Apply mutated the internal document")
	}
}

// ---- Apply: merge action ----------------------------------------------------

func TestApply_Merge_TopLevel_AddsKey(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$"},
				Data:     map[string]any{"overlay": "applied"},
			},
		}},
	})
	// ojg Set on root "$" is a no-op; merge at root uses DeepMerge directly.
	if result["overlay"] != "applied" {
		t.Errorf("overlay: got %v, want %q", result["overlay"], "applied")
	}
}

func TestApply_Merge_NestedObject_AddsKey(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/owner": "team-x"},
			},
		}},
	})
	if got := testutils.Get(t, result, "_meta", "sap/owner"); got != "team-x" {
		t.Errorf("_meta.sap/owner: got %v, want %q", got, "team-x")
	}
}

func TestApply_Merge_NestedObject_PreservesExistingKeys(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/owner": "team-x"},
			},
		}},
	})
	// The pre-existing key must still be present after the merge.
	if got := testutils.Get(t, result, "_meta", "sap/category"); got != "TBD" {
		t.Errorf("_meta.sap/category: got %v, want %q", got, "TBD")
	}
}

func TestApply_Merge_ExistingKey_Overwrites(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/category": "AI"},
			},
		}},
	})
	if got := testutils.Get(t, result, "_meta", "sap/category"); got != "AI" {
		t.Errorf("_meta.sap/category: got %v, want %q", got, "AI")
	}
}

func TestApply_Merge_ArrayElement_AddsKey(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.tools[0]"},
				Data:     map[string]any{"deprecated": true},
			},
		}},
	})
	tool := testutils.FindByName(t, result["tools"].([]any), "catalogservice-books-read")
	if tool["deprecated"] != true {
		t.Errorf("tools[0].deprecated: got %v, want true", tool["deprecated"])
	}
}

func TestApply_Merge_RootSelector_AddsKey(t *testing.T) {
	p := newProcessor(t, `{"name":"test","version":"1.0"}`)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{Root: utils.Ptr(true)},
				Data:     map[string]any{"overlay": "applied"},
			},
		}},
	})
	// Root merge uses DeepMerge directly on the root map.
	if result["overlay"] != "applied" {
		t.Errorf("overlay: got %v, want %q", result["overlay"], "applied")
	}
	if result["name"] != "test" {
		t.Error("name field lost after root merge")
	}
}

func TestApply_Merge_MultiplePatches_AllApplied(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/category": "AI"},
			},
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/owner": "team-x"},
			},
		}},
	})
	if got := testutils.Get(t, result, "_meta", "sap/category"); got != "AI" {
		t.Errorf("sap/category: got %v, want %q", got, "AI")
	}
	if got := testutils.Get(t, result, "_meta", "sap/owner"); got != "team-x" {
		t.Errorf("sap/owner: got %v, want %q", got, "team-x")
	}
}

// ---- Apply: update action ---------------------------------------------------

func TestApply_Update_NestedObject_ReplacesNode(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/category": "replaced"},
			},
		}},
	})
	meta := testutils.Get(t, result, "_meta").(map[string]any)
	// update replaces the entire node — the original key must be gone.
	if _, exists := meta["sap/category"]; !exists {
		// sap/category is in Data so it survives
	}
	if got := meta["sap/category"]; got != "replaced" {
		t.Errorf("_meta.sap/category: got %v, want %q", got, "replaced")
	}
	// The old key not present in Data must be gone after update.
	if _, exists := meta["sap/owner"]; exists {
		t.Error("unexpected key sap/owner survived update — update should replace the whole node")
	}
}

func TestApply_Update_Root_ReplacesEntireDocument(t *testing.T) {
	p := newProcessor(t, `{"name":"old","version":"1.0"}`)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{Root: utils.Ptr(true)},
				Data:     map[string]any{"name": "new"},
			},
		}},
	})
	if result["name"] != "new" {
		t.Errorf("name: got %v, want %q", result["name"], "new")
	}
	if _, exists := result["version"]; exists {
		t.Error("version should be gone after root update")
	}
}

// ---- Apply: remove action ---------------------------------------------------

func TestApply_Remove_LeafKey_DeletesKey(t *testing.T) {
	// A remove patch deletes a key by specifying the parent selector and a nil
	// data value for the target key. PatchDecomposer converts the nil value into
	// a leaf patch whose selector points directly at the key.
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/category": nil},
			},
		}},
	})
	meta := testutils.Get(t, result, "_meta").(map[string]any)
	if _, exists := meta["sap/category"]; exists {
		t.Error("sap/category should have been removed but still exists")
	}
}

func TestApply_Remove_EmptyData_IsNoOp(t *testing.T) {
	// PatchDecomposer produces zero patches for a remove with empty Data, so the
	// node is left unchanged. This is the current documented behavior.
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{},
			},
		}},
	})
	if _, exists := result["_meta"]; !exists {
		t.Error("_meta was unexpectedly removed (remove with empty Data should be a no-op)")
	}
}

// ---- Apply: patch ordering --------------------------------------------------

func TestApply_PatchesAppliedInOrder(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$._meta"}, Data: map[string]any{"sap/category": "a"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$._meta"}, Data: map[string]any{"sap/category": "b"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$._meta"}, Data: map[string]any{"sap/category": "c"}},
		}},
	})
	if got := testutils.Get(t, result, "_meta", "sap/category"); got != "c" {
		t.Errorf("sap/category: got %v, want %q", got, "c")
	}
}

func TestApply_MixedActions_AppliedInOrder(t *testing.T) {
	// merge overwrites a key, remove deletes it, merge sets a new value.
	// The remove target (sap/category) must exist in the original document
	// because Decompose resolves selectors against the original content.
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$._meta"}, Data: map[string]any{"sap/category": "first"}},
			{Action: "remove", Selector: &model.Selector{JSONPath: "$._meta"}, Data: map[string]any{"sap/category": nil}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$._meta"}, Data: map[string]any{"sap/category": "final"}},
		}},
	})
	if got := testutils.Get(t, result, "_meta", "sap/category"); got != "final" {
		t.Errorf("sap/category: got %v, want %q", got, "final")
	}
}

// ---- Apply: error paths -----------------------------------------------------

func TestApply_UnknownAction_ReturnsError(t *testing.T) {
	p := newProcessor(t, mcpContent)
	_, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "upsert",
				Selector: &model.Selector{JSONPath: "$._meta"},
				Data:     map[string]any{"sap/category": nil},
			},
		}},
	})
	if err == nil {
		t.Fatal("expected error for unknown action, got nil")
	}
}

func TestApply_JSONPath_NotFound_Merge_IsNoOp(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.nonexistent"},
				Data:     map[string]any{"key": "created"},
			},
		}},
	})
	if _, exists := result["nonexistent"]; exists {
		t.Fatal("zero-match JSONPath merge must not create a node")
	}
}

func TestApply_JSONPath_NotFound_Update_IsNoOp(t *testing.T) {
	p := newProcessor(t, mcpContent)
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{JSONPath: "$.newnode"},
				Data:     map[string]any{"key": "value"},
			},
		}},
	})
	if _, exists := result["newnode"]; exists {
		t.Fatal("zero-match JSONPath update must not create a node")
	}
}

func TestApply_InvalidJSONPath_ReturnsError(t *testing.T) {
	p := newProcessor(t, mcpContent)
	_, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$$[invalid"},
				Data:     map[string]any{"key": "val"},
			},
		}},
	})
	if err == nil {
		t.Fatal("expected error for invalid JSONPath syntax, got nil")
	}
}

func TestApply_UnsupportedSelector_ReturnsError(t *testing.T) {
	p := newProcessor(t, mcpContent)
	_, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{},
				Data:     map[string]any{"key": "val"},
			},
		}},
	})
	if err == nil {
		t.Fatal("expected error for unsupported (empty) selector, got nil")
	}
}

// ---- decompose --------------------------------------------------------------

func TestDecompose_NilData_ReturnsPatchAsIs(t *testing.T) {
	p := newProcessor(t, mcpContent)
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$._meta"},
		Data:     nil,
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Action != "remove" {
		t.Errorf("action = %q, want \"remove\"", result[0].Action)
	}
}

func TestDecompose_MergeAction_ReturnsPatchAsIs(t *testing.T) {
	p := newProcessor(t, mcpContent)
	patch := model.Patch{
		Action:   "merge",
		Selector: &model.Selector{JSONPath: "$._meta"},
		Data:     map[string]any{"sap/owner": "team"},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestDecompose_UpdateAction_ReturnsPatchAsIs(t *testing.T) {
	p := newProcessor(t, mcpContent)
	patch := model.Patch{
		Action:   "update",
		Selector: &model.Selector{JSONPath: "$._meta"},
		Data:     map[string]any{"sap/category": "AI"},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestDecompose_RemoveWithNilValueKey_ProducesChildPatch(t *testing.T) {
	p := newProcessor(t, mcpContent)
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$._meta"},
		Data:     map[string]any{"sap/category": nil},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Selector.JSONPath == "" {
		t.Error("child patch selector JSONPath should not be empty")
	}
}

func TestDecompose_RemoveWithScalarValueKey_IsSkipped(t *testing.T) {
	p := newProcessor(t, mcpContent)
	// Scalar value (non-nil, non-map) is skipped — produces no patches.
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$._meta"},
		Data: map[string]any{
			"sap/category": nil,
			"title":        "scalar — should be skipped",
		},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the nil key produces a patch; scalar key is skipped.
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1 (scalar key skipped)", len(result))
	}
}

func TestDecompose_InvalidSelector_ReturnsError(t *testing.T) {
	p := newProcessor(t, mcpContent)
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{}, // empty selector — unsupported
		Data:     map[string]any{"key": nil},
	}
	_, err := p.decompose(p.content, patch)
	if err == nil {
		t.Fatal("expected error for unsupported selector, got nil")
	}
}

func TestDecompose_RecursiveError_Propagates(t *testing.T) {
	p := newProcessor(t, mcpContent)
	// JSONPath selectors no longer validate existence — a non-existent recursive
	// child path is valid and produces a leaf patch rather than an error.
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$._meta"},
		Data: map[string]any{
			"nonexistent": map[string]any{"child": nil},
		},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error for non-existent recursive child path: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one leaf patch to be produced")
	}
}
