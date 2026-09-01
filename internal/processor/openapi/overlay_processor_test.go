//go:build unit

package openapi

import (
	"reflect"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

var petstoreDocJSON = testutils.UnmarshalFixture[map[string]any]("testdata/petstore.json")
var petstoreDocYAML = testutils.UnmarshalFixture[map[string]any]("testdata/petstore.yaml")

// ---- helpers ----------------------------------------------------------------

// newJSONProcessor creates an OverlayProcessor from the petstore JSON fixture.
func newJSONProcessor(t *testing.T) *OverlayProcessor {
	t.Helper()
	raw, err := marshaller.Marshal("application/json", petstoreDocJSON)
	if err != nil {
		t.Fatalf("marshal JSON fixture: %v", err)
	}
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   raw,
		MediaType: "application/json",
	})
	if err != nil {
		t.Fatalf("NewOverlayProcessor(JSON): %v", err)
	}
	return p
}

// newYAMLProcessor creates an OverlayProcessor from the petstore YAML fixture.
func newYAMLProcessor(t *testing.T) *OverlayProcessor {
	t.Helper()
	raw, err := marshaller.Marshal("application/yaml", petstoreDocYAML)
	if err != nil {
		t.Fatalf("marshal YAML fixture: %v", err)
	}
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   raw,
		MediaType: "application/yaml",
	})
	if err != nil {
		t.Fatalf("NewOverlayProcessor(YAML): %v", err)
	}
	return p
}

// applyJSON runs Apply on a JSON processor and returns the result as a map.
func applyJSON(t *testing.T, p *OverlayProcessor, od model.OverlayDefinition) map[string]any {
	t.Helper()
	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	parsed, err := marshaller.Unmarshal("application/json", rd.Content)
	if err != nil {
		t.Fatalf("unmarshal result JSON: %v", err)
	}
	return parsed.(map[string]any)
}

// applyYAML runs Apply on a YAML processor and returns the result as a map.
func applyYAML(t *testing.T, p *OverlayProcessor, od model.OverlayDefinition) map[string]any {
	t.Helper()
	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	parsed, err := marshaller.Unmarshal("application/yaml", rd.Content)
	if err != nil {
		t.Fatalf("unmarshal result YAML: %v", err)
	}
	return parsed.(map[string]any)
}

// operationByID returns the operation map for the given operationId by scanning
// all path items.
func operationByID(t *testing.T, result map[string]any, id string) map[string]any {
	t.Helper()
	paths, ok := result["paths"].(map[string]any)
	if !ok {
		t.Fatalf("operationByID: paths is not map[string]any")
	}
	for _, pathItem := range paths {
		item, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, op := range item {
			opMap, ok := op.(map[string]any)
			if !ok {
				continue
			}
			if opMap["operationId"] == id {
				return opMap
			}
		}
	}
	t.Fatalf("operationByID: no operation with id %q", id)
	return nil
}

// ---- NewOverlayProcessor ----------------------------------------------------

func TestNewOverlayProcessor_ValidJSON_Succeeds(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   `{"openapi":"3.0.3","info":{"title":"T","version":"1"}}`,
		MediaType: "application/json",
	}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNewOverlayProcessor_ValidYAML_Succeeds(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   "openapi: \"3.0.3\"\ninfo:\n  title: T\n  version: \"1\"\n",
		MediaType: "application/yaml",
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

func TestNewOverlayProcessor_InvalidYAML_ReturnsError(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   ":\n  - bad: [unclosed",
		MediaType: "application/yaml",
	}); err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// ---- Apply: output fields ---------------------------------------------------

func TestApply_OutputFields_PurposeAndVisibilitySet_JSON(t *testing.T) {
	rd, err := newJSONProcessor(t).Apply(model.OverlayDefinition{
		Purpose: "test-purpose",
		Overlay: model.Overlay{Visibility: "public", Patches: []model.Patch{}},
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

func TestApply_OutputFields_PurposeAndVisibilitySet_YAML(t *testing.T) {
	rd, err := newYAMLProcessor(t).Apply(model.OverlayDefinition{
		Purpose: "test-purpose",
		Overlay: model.Overlay{Visibility: "public", Patches: []model.Patch{}},
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
	raw, err := marshaller.Marshal("application/json", petstoreDocJSON)
	if err != nil {
		t.Fatalf("marshal JSON fixture: %v", err)
	}
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:    raw,
		MediaType:  "application/json",
		OrdID:      "sap.sm:resourceDefinition:petstore:v1",
		URL:        "https://example.com/petstore.json",
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
	if rd.OrdID != "sap.sm:resourceDefinition:petstore:v1" {
		t.Errorf("OrdID: got %q, want original value", rd.OrdID)
	}
	if rd.URL != "https://example.com/petstore.json" {
		t.Errorf("URL: got %q, want original value", rd.URL)
	}
	if rd.Visibility != "public" {
		t.Errorf("Visibility: got %q, want %q (from overlay)", rd.Visibility, "public")
	}
}

// ---- Apply: no patches / immutability ---------------------------------------

func TestApply_NoPatch_ContentUnchanged_JSON(t *testing.T) {
	p := newJSONProcessor(t)
	result := applyJSON(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	if !reflect.DeepEqual(result["info"], petstoreDocJSON["info"]) {
		t.Error("info changed without any patches (JSON)")
	}
}

func TestApply_NoPatch_ContentUnchanged_YAML(t *testing.T) {
	p := newYAMLProcessor(t)
	result := applyYAML(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	if result["title"] == nil && petstoreDocYAML["title"] != nil {
		t.Error("title changed without any patches (YAML)")
	}
	if len(result["servers"].([]any)) != len(petstoreDocYAML["servers"].([]any)) {
		t.Error("servers count changed without any patches (YAML)")
	}
}

func TestApply_DoesNotMutateOriginal_JSON(t *testing.T) {
	p := newJSONProcessor(t)
	applyJSON(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"mutated": true},
			},
		}},
	})
	result := applyJSON(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	info := testutils.Get(t, result, "info").(map[string]any)
	if _, ok := info["mutated"]; ok {
		t.Error("first Apply mutated the internal document (JSON)")
	}
}

func TestApply_DoesNotMutateOriginal_YAML(t *testing.T) {
	p := newYAMLProcessor(t)
	applyYAML(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"mutated": true},
			},
		}},
	})
	result := applyYAML(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	info := testutils.Get(t, result, "info").(map[string]any)
	if _, ok := info["mutated"]; ok {
		t.Error("first Apply mutated the internal document (YAML)")
	}
}

// ---- Apply: merge action (JSON) ---------------------------------------------

func TestApply_Merge_NestedObject_AddsKey_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"x-owner": "team-x"},
			},
		}},
	})
	if got := testutils.Get(t, result, "info", "x-owner"); got != "team-x" {
		t.Errorf("info.x-owner: got %v, want %q", got, "team-x")
	}
}

func TestApply_Merge_NestedObject_PreservesExistingKeys_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"x-owner": "team-x"},
			},
		}},
	})
	if got := testutils.Get(t, result, "info", "title"); got != "Petstore" {
		t.Errorf("info.title: got %v, want %q", got, "Petstore")
	}
}

func TestApply_Merge_ExistingKey_Overwrites_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"title": "NewTitle"},
			},
		}},
	})
	if got := testutils.Get(t, result, "info", "title"); got != "NewTitle" {
		t.Errorf("info.title: got %v, want %q", got, "NewTitle")
	}
}

func TestApply_Merge_RootSelector_AddsKey_JSON(t *testing.T) {
	p, _ := NewOverlayProcessor(model.ResourceDefinition{
		Content:   `{"openapi":"3.0.3","info":{"title":"T","version":"1"}}`,
		MediaType: "application/json",
	})
	result := applyJSON(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{Root: utils.Ptr(true)},
				Data:     map[string]any{"x-overlay": "applied"},
			},
		}},
	})
	if result["x-overlay"] != "applied" {
		t.Errorf("x-overlay: got %v, want %q", result["x-overlay"], "applied")
	}
	if result["openapi"] != "3.0.3" {
		t.Error("openapi field lost after root merge")
	}
}

func TestApply_Merge_Operation_AddsKey_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{Operation: "listPets"},
				Data:     map[string]any{"x-deprecated": true},
			},
		}},
	})
	op := operationByID(t, result, "listPets")
	if op["x-deprecated"] != true {
		t.Errorf("listPets.x-deprecated: got %v, want true", op["x-deprecated"])
	}
}

func TestApply_Merge_MultiplePatches_AllApplied_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-category": "a"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-owner": "team-y"}},
		}},
	})
	if got := testutils.Get(t, result, "info", "x-category"); got != "a" {
		t.Errorf("x-category: got %v, want %q", got, "a")
	}
	if got := testutils.Get(t, result, "info", "x-owner"); got != "team-y" {
		t.Errorf("x-owner: got %v, want %q", got, "team-y")
	}
}

// ---- Apply: merge action (YAML) ---------------------------------------------

func TestApply_Merge_NestedObject_AddsKey_YAML(t *testing.T) {
	result := applyYAML(t, newYAMLProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"x-owner": "team-x"},
			},
		}},
	})
	if got := testutils.Get(t, result, "info", "x-owner"); got != "team-x" {
		t.Errorf("info.x-owner: got %v, want %q", got, "team-x")
	}
}

func TestApply_Merge_Operation_AddsKey_YAML(t *testing.T) {
	result := applyYAML(t, newYAMLProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{Operation: "createPets"},
				Data:     map[string]any{"x-deprecated": true},
			},
		}},
	})
	op := operationByID(t, result, "createPets")
	if op["x-deprecated"] != true {
		t.Errorf("createPets.x-deprecated: got %v, want true", op["x-deprecated"])
	}
}

// ---- Apply: update action ---------------------------------------------------

func TestApply_Update_NestedObject_ReplacesNode_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"title": "Replaced"},
			},
		}},
	})
	info := testutils.Get(t, result, "info").(map[string]any)
	if info["title"] != "Replaced" {
		t.Errorf("info.title: got %v, want %q", info["title"], "Replaced")
	}
	if _, exists := info["version"]; exists {
		t.Error("info.version should be gone after update — update replaces the whole node")
	}
}

func TestApply_Update_Root_ReplacesEntireDocument_JSON(t *testing.T) {
	p, _ := NewOverlayProcessor(model.ResourceDefinition{
		Content:   `{"openapi":"3.0.3","info":{"title":"T","version":"1"}}`,
		MediaType: "application/json",
	})
	result := applyJSON(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{Root: utils.Ptr(true)},
				Data:     map[string]any{"openapi": "3.1.0"},
			},
		}},
	})
	if result["openapi"] != "3.1.0" {
		t.Errorf("openapi: got %v, want %q", result["openapi"], "3.1.0")
	}
	if _, exists := result["info"]; exists {
		t.Error("info should be gone after root update")
	}
}

func TestApply_Update_NestedObject_ReplacesNode_YAML(t *testing.T) {
	result := applyYAML(t, newYAMLProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"title": "Replaced"},
			},
		}},
	})
	info := testutils.Get(t, result, "info").(map[string]any)
	if info["title"] != "Replaced" {
		t.Errorf("info.title: got %v, want %q", info["title"], "Replaced")
	}
	if _, exists := info["version"]; exists {
		t.Error("info.version should be gone after update (YAML)")
	}
}

// ---- Apply: remove action ---------------------------------------------------

func TestApply_Remove_TargetedNode_DeletesNode_JSON(t *testing.T) {
	// The openapi overlay processor applies patches directly without decomposition.
	// A remove selector targets the node itself — here we remove $.externalDocs.
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$.externalDocs"},
				Data:     nil,
			},
		}},
	})
	if _, exists := result["externalDocs"]; exists {
		t.Error("externalDocs should have been removed but still exists")
	}
	// Sibling keys must be untouched.
	if _, exists := result["info"]; !exists {
		t.Error("info was unexpectedly removed")
	}
}

func TestApply_Remove_TargetedNode_DeletesNode_YAML(t *testing.T) {
	result := applyYAML(t, newYAMLProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$.externalDocs"},
				Data:     nil,
			},
		}},
	})
	if _, exists := result["externalDocs"]; exists {
		t.Error("externalDocs should have been removed but still exists (YAML)")
	}
	if _, exists := result["info"]; !exists {
		t.Error("info was unexpectedly removed (YAML)")
	}
}

func TestApply_Remove_EmptyData_StillRemovesTargetedNode_JSON(t *testing.T) {
	// The openapi overlay processor does NOT use PatchDecomposer — patches are
	// applied directly. A remove selector always targets and deletes the matched
	// node regardless of what Data contains (Data is ignored for remove).
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$.externalDocs"},
				Data:     map[string]any{},
			},
		}},
	})
	if _, exists := result["externalDocs"]; exists {
		t.Error("externalDocs should have been removed even when Data is empty")
	}
}

func TestApply_Remove_NilData_RemovesTargetedNode_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$.externalDocs"},
				Data:     nil,
			},
		}},
	})
	if _, exists := result["externalDocs"]; exists {
		t.Error("externalDocs should have been removed when Data is nil")
	}
}

func TestApply_Remove_NilData_RemovesTargetedNode_YAML(t *testing.T) {
	result := applyYAML(t, newYAMLProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{JSONPath: "$.externalDocs"},
				Data:     nil,
			},
		}},
	})
	if _, exists := result["externalDocs"]; exists {
		t.Error("externalDocs should have been removed when Data is nil (YAML)")
	}
}

func TestApply_RootSelector_Remove_ClearsDocument_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{Root: utils.Ptr(true)},
				Data:     nil,
			},
		}},
	})
	if len(result) != 0 {
		t.Errorf("expected empty document after root remove, got %d keys", len(result))
	}
}

func TestApply_Remove_Operation_DeletesOperation_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "remove",
				Selector: &model.Selector{Operation: "createPets"},
				Data:     nil,
			},
		}},
	})
	paths := result["paths"].(map[string]any)
	pets := paths["/pets"].(map[string]any)
	if _, exists := pets["post"]; exists {
		t.Error("createPets operation should have been removed but still exists")
	}
}

// ---- Apply: patch ordering --------------------------------------------------

func TestApply_PatchesAppliedInOrder_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-category": "a"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-category": "b"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-category": "c"}},
		}},
	})
	if got := testutils.Get(t, result, "info", "x-category"); got != "c" {
		t.Errorf("x-category: got %v, want %q", got, "c")
	}
}

func TestApply_MixedActions_AppliedInOrder_JSON(t *testing.T) {
	// merge adds x-sap-ext, remove deletes externalDocs, merge adds a top-level key.
	// No decomposer: each selector targets the node that will be acted on directly.
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-category": "first"}},
			{Action: "remove", Selector: &model.Selector{JSONPath: "$.externalDocs"}, Data: nil},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.info"}, Data: map[string]any{"x-category": "final"}},
		}},
	})
	if got := testutils.Get(t, result, "info", "x-category"); got != "final" {
		t.Errorf("x-category: got %v, want %q", got, "final")
	}
	if _, exists := result["externalDocs"]; exists {
		t.Error("externalDocs should have been removed")
	}
}

// ---- Apply: error paths -----------------------------------------------------

func TestApply_UnknownAction_ReturnsError_JSON(t *testing.T) {
	_, err := newJSONProcessor(t).Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "upsert",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"title": nil},
			},
		}},
	})
	if err == nil {
		t.Fatal("expected error for unknown action, got nil")
	}
}

func TestApply_JSONPath_NotFound_Merge_IsNoOp_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
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

func TestApply_JSONPath_NotFound_Update_IsNoOp_JSON(t *testing.T) {
	result := applyJSON(t, newJSONProcessor(t), model.OverlayDefinition{
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

func TestApply_JSONPath_NotFound_Merge_IsNoOp_YAML(t *testing.T) {
	result := applyYAML(t, newYAMLProcessor(t), model.OverlayDefinition{
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

func TestApply_InvalidJSONPath_ReturnsError_JSON(t *testing.T) {
	_, err := newJSONProcessor(t).Apply(model.OverlayDefinition{
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

func TestApply_UnsupportedSelector_ReturnsError_JSON(t *testing.T) {
	_, err := newJSONProcessor(t).Apply(model.OverlayDefinition{
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

func TestApply_UnknownAction_ReturnsError_YAML(t *testing.T) {
	_, err := newYAMLProcessor(t).Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "upsert",
				Selector: &model.Selector{JSONPath: "$.info"},
				Data:     map[string]any{"title": nil},
			},
		}},
	})
	if err == nil {
		t.Fatal("expected error for unknown action (YAML), got nil")
	}
}

// ---- decompose --------------------------------------------------------------

func TestDecompose_NilData_ReturnsPatchAsIs(t *testing.T) {
	p := newJSONProcessor(t)
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$.info"},
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
	p := newJSONProcessor(t)
	patch := model.Patch{
		Action:   "merge",
		Selector: &model.Selector{JSONPath: "$.info"},
		Data:     map[string]any{"x-owner": "team"},
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
	p := newJSONProcessor(t)
	patch := model.Patch{
		Action:   "update",
		Selector: &model.Selector{JSONPath: "$.info"},
		Data:     map[string]any{"title": "New"},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestDecompose_RemoveWithMapData_DecomposesIntoChildPatches(t *testing.T) {
	p := newJSONProcessor(t)
	// Data is a map with two keys: one has a nil value (produces a child patch),
	// one has a scalar value (skipped by the "ignore non-map non-nil values" guard).
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$.info"},
		Data: map[string]any{
			"contact": nil,
			"title":   "scalar — should be skipped",
		},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the "contact" key (nil value) is decomposed; "title" (scalar) is skipped.
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1 (scalar key skipped)", len(result))
	}
	if result[0].Selector.JSONPath == "" {
		t.Error("child patch selector JSONPath should not be empty")
	}
}

func TestDecompose_RemoveWithNilValueKey_ProducesChildPatch(t *testing.T) {
	p := newJSONProcessor(t)
	// A nil-value key is treated as a leaf and produces a child patch.
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$.info"},
		Data: map[string]any{
			"contact": nil,
		},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestDecompose_InvalidSelector_ReturnsError(t *testing.T) {
	p := newJSONProcessor(t)
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
	p := newJSONProcessor(t)
	// contact->email->nil leaf: exercises the map-value recursion path successfully.
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$.info"},
		Data: map[string]any{
			"contact": map[string]any{"email": nil},
		},
	}
	result, err := p.decompose(p.content, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// contact->email->nil leaf produces one patch
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
}

func TestDecompose_RecursiveResolveError_Propagates(t *testing.T) {
	p := newJSONProcessor(t)
	// JSONPath selectors no longer validate existence — a non-existent recursive
	// child path is valid and produces a leaf patch rather than an error.
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$.info"},
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
