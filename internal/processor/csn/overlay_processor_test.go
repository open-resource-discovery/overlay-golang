//go:build unit

package csn

import (
	"reflect"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

var flightContent = testutils.LoadFixture("testdata/flight_model.json")
var flightDoc = testutils.UnmarshalFixture[map[string]any]("testdata/flight_model.json")

// ---- helpers ----------------------------------------------------------------

// makeDefinition builds a ResourceDefinition with the given content and
// common test metadata filled in.
func makeDefinition(content string) model.ResourceDefinition {
	return model.ResourceDefinition{
		URL:       "https://example.com/csn.json",
		OrdID:     "test:resource:csn:v1",
		MediaType: "application/json",
		Content:   content,
	}
}

// mustNewProcessor creates an OverlayProcessor from a raw JSON string.
func mustNewProcessor(t *testing.T, content string) *OverlayProcessor {
	t.Helper()
	p, err := NewOverlayProcessor(makeDefinition(content))
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	return p
}

// definitions returns the "definitions" map from a result document.
func definitions(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	defs, ok := doc["definitions"].(map[string]any)
	if !ok {
		t.Fatalf("definitions: expected map[string]any, got %T", doc["definitions"])
	}
	return defs
}

// entity returns a named entity from the definitions map.
func entity(t *testing.T, doc map[string]any, name string) map[string]any {
	t.Helper()
	defs := definitions(t, doc)
	e, ok := defs[name].(map[string]any)
	if !ok {
		t.Fatalf("entity %q: expected map[string]any, got %T", name, defs[name])
	}
	return e
}

// element returns a named element from an entity's "elements" map.
func element(t *testing.T, doc map[string]any, entityName, elemName string) map[string]any {
	t.Helper()
	ent := entity(t, doc, entityName)
	elems, ok := ent["elements"].(map[string]any)
	if !ok {
		t.Fatalf("entity %q elements: expected map[string]any, got %T", entityName, ent["elements"])
	}
	elem, ok := elems[elemName].(map[string]any)
	if !ok {
		t.Fatalf("element %q.%q: expected map[string]any, got %T", entityName, elemName, elems[elemName])
	}
	return elem
}

// ---- NewOverlayProcessor ----------------------------------------------------

func TestNewOverlayProcessor_ValidJSON_Succeeds(t *testing.T) {
	if _, err := NewOverlayProcessor(makeDefinition(`{"definitions":{}}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewOverlayProcessor_InvalidJSON_ReturnsError(t *testing.T) {
	if _, err := NewOverlayProcessor(makeDefinition(`{not valid json`)); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestNewOverlayProcessor_StoresDefinitionMetadata(t *testing.T) {
	def := makeDefinition(`{"definitions":{}}`)
	def.OrdID = "ns:resource:csn:v2"
	p, err := NewOverlayProcessor(def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.definition.OrdID != "ns:resource:csn:v2" {
		t.Errorf("OrdID not stored: got %q", p.definition.OrdID)
	}
}

// ---- Apply: output fields ---------------------------------------------------

func TestApply_OutputFields_PurposeAndVisibilitySet(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	od := testutils.NoPatches()
	od.Purpose = "external"
	od.Overlay.Visibility = "private"

	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rd.Purpose != "external" {
		t.Errorf("Purpose: got %q, want %q", rd.Purpose, "external")
	}
	if rd.Visibility != "private" {
		t.Errorf("Visibility: got %q, want %q", rd.Visibility, "private")
	}
}

func TestApply_OutputFields_OriginalDefinitionFieldsPreserved(t *testing.T) {
	def := model.ResourceDefinition{
		Content:     flightContent,
		MediaType:   "application/json",
		OrdID:       "ns:resource:csn:v1",
		URL:         "https://example.com/csn.json",
		Perspective: "system-instance",
	}
	p, err := NewOverlayProcessor(def)
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	od := testutils.NoPatches()
	od.Overlay.Visibility = "public"

	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rd.OrdID != "ns:resource:csn:v1" {
		t.Errorf("OrdID: got %q", rd.OrdID)
	}
	if rd.URL != "https://example.com/csn.json" {
		t.Errorf("URL: got %q", rd.URL)
	}
	if rd.Perspective != "system-instance" {
		t.Errorf("Perspective: got %q", rd.Perspective)
	}
	// Visibility comes from the overlay, not the original definition.
	if rd.Visibility != "public" {
		t.Errorf("Visibility: got %q, want %q", rd.Visibility, "public")
	}
}

// ---- Apply: no patches / immutability ---------------------------------------

func TestApply_NoPatch_ContentUnchanged(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.NoPatches())
	if !reflect.DeepEqual(result, flightDoc) {
		t.Error("content changed with no patches applied")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	p := mustNewProcessor(t, flightContent)

	// First Apply modifies the Airline entity.
	testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline"},
		map[string]any{"@x-custom": "mutated"}))

	// Second Apply with no patches must return the original document.
	result := testutils.ApplyAndParse(t, p, testutils.NoPatches())
	if !reflect.DeepEqual(result, flightDoc) {
		t.Error("first Apply mutated the internal processor state")
	}
}

// ---- Apply: merge action ----------------------------------------------------

func TestApply_Merge_EntityType_AddsAnnotation(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline"},
		map[string]any{"@x-custom": "overlay-value"}))

	ent := entity(t, result, "Airline")
	if ent["@x-custom"] != "overlay-value" {
		t.Errorf("@x-custom: got %v, want %q", ent["@x-custom"], "overlay-value")
	}
}

func TestApply_Merge_EntityType_PreservesExistingFields(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline"},
		map[string]any{"@x-custom": "overlay-value"}))

	ent := entity(t, result, "Airline")
	if ent["kind"] != "entity" {
		t.Errorf("kind field lost after merge: got %v", ent["kind"])
	}
}

func TestApply_Merge_EntityType_OverwritesExistingAnnotation(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline"},
		map[string]any{"@EndUserText.label": "New Airline Label"}))

	ent := entity(t, result, "Airline")
	if ent["@EndUserText.label"] != "New Airline Label" {
		t.Errorf("@EndUserText.label: got %v, want %q", ent["@EndUserText.label"], "New Airline Label")
	}
}

func TestApply_Merge_EntityType_DoesNotAffectOtherEntities(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline"},
		map[string]any{"@x-custom": "overlay-value"}))

	defs := definitions(t, result)
	// Airport must not have the injected annotation.
	if airport, ok := defs["Airport"].(map[string]any); ok {
		if _, exists := airport["@x-custom"]; exists {
			t.Error("Airport should not have @x-custom after Airline merge")
		}
	}
}

func TestApply_Merge_EntityTypeWithProperty_AddsAnnotationToElement(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline", PropertyType: "Name"},
		map[string]any{"@x-custom": "element-value"}))

	elem := element(t, result, "Airline", "Name")
	if elem["@x-custom"] != "element-value" {
		t.Errorf("@x-custom: got %v, want %q", elem["@x-custom"], "element-value")
	}
}

func TestApply_Merge_EntityTypeWithProperty_PreservesOtherElements(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "Airline", PropertyType: "Name"},
		map[string]any{"@x-custom": "element-value"}))

	// AirlineID must be untouched.
	airlineID := element(t, result, "Airline", "AirlineID")
	if _, exists := airlineID["@x-custom"]; exists {
		t.Error("AirlineID should not have @x-custom after Name element merge")
	}
}

func TestApply_Merge_JSONPath_AddsField(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.definitions.Airline"},
		map[string]any{"@x-overlay": "json-path"}))

	ent := entity(t, result, "Airline")
	if ent["@x-overlay"] != "json-path" {
		t.Errorf("@x-overlay: got %v, want %q", ent["@x-overlay"], "json-path")
	}
}

func TestApply_Merge_Root_AddsTopLevelKey(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"x-overlay": "applied"}))

	if result["x-overlay"] != "applied" {
		t.Errorf("x-overlay: got %v, want %q", result["x-overlay"], "applied")
	}
	// Existing top-level fields must survive.
	if _, exists := result["definitions"]; !exists {
		t.Error("definitions field lost after root merge")
	}
}

// ---- Apply: update action ---------------------------------------------------

func TestApply_Update_EntityType_ReplacesEntityEntirely(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EntityType: "Airline"},
		map[string]any{"kind": "entity", "@EndUserText.label": "Replaced"}))

	ent := entity(t, result, "Airline")
	if ent["@EndUserText.label"] != "Replaced" {
		t.Errorf("@EndUserText.label: got %v, want %q", ent["@EndUserText.label"], "Replaced")
	}
	// "elements" was not in the replacement — must be gone after update.
	if _, exists := ent["elements"]; exists {
		t.Error("elements should be absent after entity update (replace, not merge)")
	}
}

func TestApply_Update_EntityTypeWithProperty_ReplacesElementEntirely(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EntityType: "Airline", PropertyType: "Name"},
		map[string]any{"type": "cds.String", "length": int64(100)}))

	elem := element(t, result, "Airline", "Name")
	if elem["type"] != "cds.String" {
		t.Errorf("type: got %v, want %q", elem["type"], "cds.String")
	}
	// "@EndUserText.label" was not in the replacement — must be gone.
	if _, exists := elem["@EndUserText.label"]; exists {
		t.Error("@EndUserText.label should be absent after element update")
	}
}

func TestApply_Update_Root_ReplacesEntireDocument(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"replaced": true}))

	if result["replaced"] != true {
		t.Errorf("replaced: got %v, want true", result["replaced"])
	}
	if _, exists := result["definitions"]; exists {
		t.Error("definitions should be gone after root update")
	}
}

// ---- Apply: remove action ---------------------------------------------------

func TestApply_Remove_JSONPath_RemovesField(t *testing.T) {
	// Remove via a nil-keyed Data entry targeting a top-level key.
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$"},
		map[string]any{"definitions": nil}))

	if _, exists := result["definitions"]; exists {
		t.Error("definitions should have been removed but still exists")
	}
}

func TestApply_Remove_NilData_RemovesTargetedNode(t *testing.T) {
	// nil Data is now fast-pathed by Decompose (returns the patch unchanged).
	// apply then resolves the selector and removes the targeted entity entirely.
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{EntityType: "Airline"}, nil))

	defs := definitions(t, result)
	if _, exists := defs["Airline"]; exists {
		t.Error("Airline should have been removed (nil Data remove now targets the node directly)")
	}
	// Other entities must survive.
	if _, exists := defs["Airport"]; !exists {
		t.Error("Airport should still be present after Airline removal")
	}
}

func TestApply_Remove_EntityType_RemovesEntityViaDecomposer(t *testing.T) {
	// Remove a definitions entry by targeting its parent with a nil-keyed entry.
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$.definitions"},
		map[string]any{"Airline": nil}))

	defs := definitions(t, result)
	if _, exists := defs["Airline"]; exists {
		t.Error("Airline should have been removed but still exists")
	}
	// Other entities must survive.
	if _, exists := defs["Airport"]; !exists {
		t.Error("Airport should still be present after Airline removal")
	}
}

func TestApply_Remove_WithNilDataKey_DeletesLeafViaDecomposer(t *testing.T) {
	// A nil data value causes PatchDecomposer to produce a leaf patch targeting
	// the specific key — the remove then deletes only that key.
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$.definitions.Airline"},
		map[string]any{"@EndUserText.label": nil}))

	ent := entity(t, result, "Airline")
	if _, exists := ent["@EndUserText.label"]; exists {
		t.Error("@EndUserText.label should have been removed via nil-leaf patch")
	}
	// kind must survive.
	if ent["kind"] != "entity" {
		t.Error("kind field was unexpectedly removed")
	}
}

func TestApply_RootSelector_Remove_ClearsDocument(t *testing.T) {
	// nil Data is now fast-pathed by Decompose. apply resolves the root selector
	// and the remove root branch returns an empty map.
	p := mustNewProcessor(t, `{"definitions":{}}`)
	rd, err := p.Apply(testutils.OnePatch("remove", model.Selector{Root: utils.Ptr(true)}, nil))
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	result := testutils.UnmarshalResult[map[string]any](t, rd.MediaType, rd.Content)
	if len(result) != 0 {
		t.Errorf("expected empty document after root remove, got: %v", result)
	}
}

// ---- Apply: patch ordering --------------------------------------------------

func TestApply_PatchesAppliedInOrder(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	od := model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{EntityType: "Airline"}, Data: map[string]any{"@EndUserText.label": "first"}},
			{Action: "merge", Selector: &model.Selector{EntityType: "Airline"}, Data: map[string]any{"@EndUserText.label": "second"}},
			{Action: "merge", Selector: &model.Selector{EntityType: "Airline"}, Data: map[string]any{"@EndUserText.label": "third"}},
		}},
	}
	result := testutils.ApplyAndParse(t, p, od)
	ent := entity(t, result, "Airline")
	if ent["@EndUserText.label"] != "third" {
		t.Errorf("@EndUserText.label: got %v, want %q (last patch must win)", ent["@EndUserText.label"], "third")
	}
}

func TestApply_MixedActions_AppliedInOrder(t *testing.T) {
	// merge adds annotation; remove deletes it; merge sets a new value.
	p := mustNewProcessor(t, flightContent)
	od := model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{EntityType: "Airline"}, Data: map[string]any{"@EndUserText.label": "step-1"}},
			{Action: "remove", Selector: &model.Selector{JSONPath: "$.definitions.Airline"}, Data: map[string]any{"@EndUserText.label": nil}},
			{Action: "merge", Selector: &model.Selector{EntityType: "Airline"}, Data: map[string]any{"@EndUserText.label": "step-3"}},
		}},
	}
	result := testutils.ApplyAndParse(t, p, od)
	ent := entity(t, result, "Airline")
	if ent["@EndUserText.label"] != "step-3" {
		t.Errorf("@EndUserText.label: got %v, want %q", ent["@EndUserText.label"], "step-3")
	}
}

func TestApply_MultiplePatches_StopsOnFirstError(t *testing.T) {
	// Unknown action with a nil-keyed entry forces Decompose to produce a patch
	// that reaches apply's default error branch.
	p := mustNewProcessor(t, flightContent)
	od := model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "badaction", Selector: &model.Selector{JSONPath: "$"}, Data: map[string]any{"csnInteropEffective": nil}},
			{Action: "merge", Selector: &model.Selector{EntityType: "Airline"}, Data: map[string]any{"@x-second": "should-not-appear"}},
		}},
	}
	_, err := p.Apply(od)
	if err == nil {
		t.Fatal("expected error for bad action in first patch, got nil")
	}
	// Processor state must be untouched — a subsequent Apply returns the original doc.
	result := testutils.ApplyAndParse(t, p, testutils.NoPatches())
	ent := entity(t, result, "Airline")
	if _, exists := ent["@x-second"]; exists {
		t.Error("second patch was applied despite first patch failing")
	}
}

// ---- Apply: error paths -----------------------------------------------------

func TestApply_UnknownAction_ReturnsError(t *testing.T) {
	// An unknown action only reaches the `default` error branch inside `apply`
	// when Decompose produces at least one patch for it. Use a nil-keyed entry
	// to force a patch through Decompose.
	p := mustNewProcessor(t, flightContent)
	_, err := p.Apply(testutils.OnePatch("replace",
		model.Selector{JSONPath: "$"},
		map[string]any{"csnInteropEffective": nil}))
	if err == nil {
		t.Fatal("expected error for unsupported action with nil-keyed data, got nil")
	}
}

func TestApply_UnknownAction_WithNonNilScalarData_IsNoOp(t *testing.T) {
	// Non-nil scalar values in Data are skipped by Decompose → zero patches →
	// no error, no change. This documents the current behavior.
	// Compare against an ojg-parsed baseline so number types are consistent.
	p := mustNewProcessor(t, flightContent)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("replace",
		model.Selector{EntityType: "Airline"},
		map[string]any{"x": 1}))
	if !reflect.DeepEqual(result, flightDoc) {
		t.Error("expected no-op for unknown action with non-nil scalar data, but content changed")
	}
}

func TestApply_InvalidJSONPath_ReturnsError(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{JSONPath: "$$[invalid"},
		map[string]any{"x": 1}))
	if err == nil {
		t.Fatal("expected error for invalid JSONPath, got nil")
	}
}

func TestApply_UnsupportedSelector_ReturnsError(t *testing.T) {
	p := mustNewProcessor(t, flightContent)
	_, err := p.Apply(testutils.OnePatch("merge", model.Selector{}, map[string]any{"x": 1}))
	if err == nil {
		t.Fatal("expected error for empty selector, got nil")
	}
}

func TestApply_MergeSelector_NotFound_ReturnsError(t *testing.T) {
	// Semantic selectors that match nothing still return an error.
	p := mustNewProcessor(t, flightContent)
	_, err := p.Apply(testutils.OnePatch("merge", model.Selector{EntityType: "NonExistent"}, map[string]any{"x": 1}))
	if err == nil {
		t.Error("expected error for semantic selector that matches nothing, got nil")
	}
}

func TestApply_Merge_JSONPath_NonExistentPath_ReturnsError(t *testing.T) {
	// A selector that matches nothing is an error: overlays never create a
	// missing target, not even via jsonPath.
	p := mustNewProcessor(t, flightContent)
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.nonexistent"},
		map[string]any{"x": "created"}))
	if err == nil {
		t.Fatal("expected an error for a merge that matches no existing node, got nil")
	}
}

func TestApply_Update_JSONPath_NonExistentPath_ReturnsError(t *testing.T) {
	// An update via a non-existent JSONPath is an error, not a create.
	p := mustNewProcessor(t, flightContent)
	_, err := p.Apply(testutils.OnePatch("update",
		model.Selector{JSONPath: "$.newnode"},
		map[string]any{"key": "value"}))
	if err == nil {
		t.Fatal("expected an error for an update that matches no existing node, got nil")
	}
}

// ---- Apply: root selector ---------------------------------------------------

func TestApply_RootSelector_Merge_AddsTopLevelKey(t *testing.T) {
	p := mustNewProcessor(t, `{"definitions":{}}`)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"x-overlay": "applied"}))

	if result["x-overlay"] != "applied" {
		t.Errorf("x-overlay: got %v, want %q", result["x-overlay"], "applied")
	}
	if _, exists := result["definitions"]; !exists {
		t.Error("definitions field lost after root merge")
	}
}

func TestApply_RootSelector_Update_ReplacesDocument(t *testing.T) {
	p := mustNewProcessor(t, `{"definitions":{}}`)
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"replaced": true}))

	if result["replaced"] != true {
		t.Errorf("replaced: got %v, want true", result["replaced"])
	}
	if _, exists := result["definitions"]; exists {
		t.Error("definitions should be gone after root update")
	}
}
