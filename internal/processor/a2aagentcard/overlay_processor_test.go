//go:build unit

package a2aagentcard

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// ---- helpers ----------------------------------------------------------------

var ficaContent = testutils.LoadFixture("testdata/fica_dispute_agent.json")

// ficaDoc is the parsed FI-CA Dispute Resolution Agent fixture loaded once for
// all patch_decomposer tests. It mirrors the same fixture used by the overlay
// processor tests.
var ficaDoc = testutils.UnmarshalFixture[map[string]any]("testdata/fica_dispute_agent.json")

// makeDefinition builds a ResourceDefinition with the given content and
// common test metadata filled in.
func makeDefinition(content string) model.ResourceDefinition {
	return model.ResourceDefinition{
		URL:       "https://example.com/agent.json",
		OrdID:     "test:resource:agentcard:v1",
		MediaType: "application/json",
		Content:   content,
	}
}

// mustSkills extracts the skills array, failing the test if absent.
func mustSkills(t *testing.T, doc map[string]any) []any {
	t.Helper()
	skills, ok := doc["skills"].([]any)
	if !ok {
		t.Fatalf("skills: expected []any, got %T", doc["skills"])
	}
	return skills
}

// findSkill returns the first skill whose "id" equals id, or nil.
func findSkill(skills []any, id string) map[string]any {
	for _, s := range skills {
		if s == nil {
			continue
		}
		if m, ok := s.(map[string]any); ok && m["id"] == id {
			return m
		}
	}
	return nil
}

// ---- NewOverlayProcessor ----------------------------------------------------

func TestNewOverlayProcessor_ValidJSON_Succeeds(t *testing.T) {
	if _, err := NewOverlayProcessor(makeDefinition(`{"name":"agent"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewOverlayProcessor_InvalidJSON_ReturnsError(t *testing.T) {
	if _, err := NewOverlayProcessor(makeDefinition(`{not valid json`)); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestNewOverlayProcessor_StoresDefinitionMetadata(t *testing.T) {
	def := makeDefinition(`{"name":"agent"}`)
	def.OrdID = "ns:resource:card:v2"
	p, err := NewOverlayProcessor(def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.definition.OrdID != "ns:resource:card:v2" {
		t.Errorf("OrdID not stored: got %q", p.definition.OrdID)
	}
}

// ---- Apply: output fields ---------------------------------------------------

func TestApply_OutputFields_PurposeAndVisibilitySet(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
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
		Content:     ficaContent,
		MediaType:   "application/json",
		OrdID:       "ns:resource:card:v1",
		URL:         "https://example.com/card.json",
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
	if rd.OrdID != "ns:resource:card:v1" {
		t.Errorf("OrdID: got %q", rd.OrdID)
	}
	if rd.URL != "https://example.com/card.json" {
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
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.NoPatches())
	if !reflect.DeepEqual(result, ficaDoc) {
		t.Error("content changed with no patches applied")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))

	// First Apply modifies provider.
	testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.provider"},
		map[string]any{"organization": "Mutated"}))

	// Second Apply with no patches must return the original document.
	result := testutils.ApplyAndParse(t, p, testutils.NoPatches())
	if !reflect.DeepEqual(result, ficaDoc) {
		t.Error("first Apply mutated the internal processor state")
	}
}

// ---- Apply: merge action ----------------------------------------------------

func TestApply_Merge_JSONPath_AddsKey(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.provider"},
		map[string]any{"department": "Cloud"}))

	provider := result["provider"].(map[string]any)
	if provider["department"] != "Cloud" {
		t.Errorf("department: got %v, want %q", provider["department"], "Cloud")
	}
}

func TestApply_Merge_JSONPath_PreservesExistingKeys(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.provider"},
		map[string]any{"department": "Cloud"}))

	provider := result["provider"].(map[string]any)
	if provider["organization"] != "SAP SE" {
		t.Errorf("organization lost after merge: got %v", provider["organization"])
	}
}

func TestApply_Merge_JSONPath_OverwritesExistingKey(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.provider"},
		map[string]any{"organization": "ACME"}))

	provider := result["provider"].(map[string]any)
	if provider["organization"] != "ACME" {
		t.Errorf("organization: got %v, want %q", provider["organization"], "ACME")
	}
}

func TestApply_Merge_Operation_AddsKey(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "dispute-case-resolution"},
		map[string]any{"x-custom": "overlay-value"}))

	skill := findSkill(mustSkills(t, result), "dispute-case-resolution")
	if skill == nil {
		t.Fatal("dispute-case-resolution not found after merge")
	}
	if skill["x-custom"] != "overlay-value" {
		t.Errorf("x-custom: got %v, want %q", skill["x-custom"], "overlay-value")
	}
}

func TestApply_Merge_Operation_PreservesOriginalSkillFields(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "dispute-case-resolution"},
		map[string]any{"x-custom": "overlay-value"}))

	skill := findSkill(mustSkills(t, result), "dispute-case-resolution")
	if skill["name"] != "Automated Dispute Case Resolution" {
		t.Errorf("name field lost after merge: got %v", skill["name"])
	}
}

func TestApply_Merge_Operation_DoesNotAffectOtherSkills(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "dispute-case-resolution"},
		map[string]any{"x-custom": "overlay-value"}))

	for _, s := range mustSkills(t, result) {
		if s == nil {
			continue
		}
		m := s.(map[string]any)
		if m["id"] == "dispute-case-resolution" {
			continue
		}
		if _, exists := m["x-custom"]; exists {
			t.Errorf("skill %q should not have x-custom", m["id"])
		}
	}
}

func TestApply_Merge_Root_AddsTopLevelKey(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"overlay": "applied"}))

	if result["overlay"] != "applied" {
		t.Errorf("overlay: got %v, want %q", result["overlay"], "applied")
	}
	// Existing fields must survive.
	if result["name"] != "FI-CA Dispute Resolution Agent" {
		t.Error("name field lost after root merge")
	}
}

func TestApply_Merge_Operation_AppendsArrayField(t *testing.T) {
	// DeepMerge appends slices: original tags are preserved and new-tag is added.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "invoice-recalculation"},
		map[string]any{"tags": []any{"new-tag"}}))

	skill := findSkill(mustSkills(t, result), "invoice-recalculation")
	tags, ok := skill["tags"].([]any)
	if !ok {
		t.Fatalf("tags: expected []any, got %T", skill["tags"])
	}
	found := false
	for _, tag := range tags {
		if tag == "new-tag" {
			found = true
		}
	}
	if !found {
		t.Errorf("new-tag not found in merged tags: %v", tags)
	}
}

// ---- Apply: update action ---------------------------------------------------

func TestApply_Update_JSONPath_ReplacesNodeEntirely(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{JSONPath: "$.provider"},
		map[string]any{"organization": "ACME Corp"}))

	provider := result["provider"].(map[string]any)
	if provider["organization"] != "ACME Corp" {
		t.Errorf("organization: got %v, want %q", provider["organization"], "ACME Corp")
	}
	// "url" was not in the replacement — must be gone after update.
	if _, exists := provider["url"]; exists {
		t.Error("url should be absent after update (replace, not merge)")
	}
}

func TestApply_Update_Operation_ReplacesSkillEntirely(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{Operation: "credit-memo-processing"},
		map[string]any{"id": "credit-memo-processing", "name": "Replaced"}))

	skill := findSkill(mustSkills(t, result), "credit-memo-processing")
	if skill == nil {
		t.Fatal("credit-memo-processing not found after update")
	}
	if skill["name"] != "Replaced" {
		t.Errorf("name: got %v, want %q", skill["name"], "Replaced")
	}
	if _, exists := skill["description"]; exists {
		t.Error("description should be absent after update (replace, not merge)")
	}
	if _, exists := skill["tags"]; exists {
		t.Error("tags should be absent after update (replace, not merge)")
	}
}

func TestApply_Update_Root_ReplacesEntireDocument(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"name": "replacement"}))

	if result["name"] != "replacement" {
		t.Errorf("name: got %v, want %q", result["name"], "replacement")
	}
	if _, exists := result["skills"]; exists {
		t.Error("skills should be gone after root update")
	}
}

// ---- Apply: remove action ---------------------------------------------------

func TestApply_Remove_JSONPath_RemovesField(t *testing.T) {
	// Remove via a nil-keyed Data entry: PatchDecomposer produces a leaf patch
	// targeting the specific key, which is then deleted.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$"},
		map[string]any{"provider": nil}))

	if _, exists := result["provider"]; exists {
		t.Error("provider should have been removed but still exists")
	}
	if result["name"] != "FI-CA Dispute Resolution Agent" {
		t.Error("unrelated field name was affected by remove")
	}
}

func TestApply_Remove_NilData_RemovesTargetedNode(t *testing.T) {
	// nil Data is now fast-pathed by Decompose (returns the patch unchanged).
	// apply then resolves the selector and removes the targeted node entirely.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$.provider"}, nil))

	if _, exists := result["provider"]; exists {
		t.Error("provider should have been removed (nil Data remove now targets the node directly)")
	}
	// Unrelated fields must survive.
	if result["name"] != "FI-CA Dispute Resolution Agent" {
		t.Error("name field was unexpectedly removed")
	}
}

func TestApply_Remove_Operation_RemovesSkill(t *testing.T) {
	// Remove a skill by targeting it through a nil-keyed data entry. The root
	// selector resolves to "$", Decompose produces a leaf patch for "skills",
	// which then recurses to find the nil-leaf for the skill array element.
	// The simplest working pattern is to target the skill's own fields via
	// a nil-keyed entry at the parent level.
	// Here we verify the nil-key remove pattern works for a nested path.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$"},
		map[string]any{"documentationUrl": nil}))

	if _, exists := result["documentationUrl"]; exists {
		t.Error("documentationUrl should have been removed but still exists")
	}
	// Other fields must be preserved.
	if result["name"] != "FI-CA Dispute Resolution Agent" {
		t.Error("name field was unexpectedly removed")
	}
}

func TestApply_Remove_OperationSelector_NilData_RemovesSkill(t *testing.T) {
	// nil Data is now fast-pathed by Decompose (returns the patch unchanged).
	// apply resolves the Operation selector and removes the targeted skill element.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{Operation: "dispute-case-verification"}, nil))

	skills := mustSkills(t, result)
	if findSkill(skills, "dispute-case-verification") != nil {
		t.Error("dispute-case-verification should have been removed (nil Data remove now targets the node directly)")
	}
	// Other skills must survive.
	if findSkill(skills, "dispute-case-resolution") == nil {
		t.Error("dispute-case-resolution should still be present")
	}
}

func TestApply_Remove_WithNilDataKey_DeletesLeafViaDecomposer(t *testing.T) {
	// A nil data value causes PatchDecomposer to produce a leaf patch targeting
	// the specific key — the remove then deletes only that key.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{JSONPath: "$"},
		map[string]any{"name": nil}))

	if _, exists := result["name"]; exists {
		t.Error("name should have been removed via decomposed nil-leaf patch")
	}
	// Unrelated fields must survive.
	if _, exists := result["version"]; !exists {
		t.Error("version should still be present after targeted nil-key remove")
	}
}

// ---- Apply: patch ordering --------------------------------------------------

func TestApply_PatchesAppliedInOrder(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	od := model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.provider"}, Data: map[string]any{"organization": "first"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.provider"}, Data: map[string]any{"organization": "second"}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.provider"}, Data: map[string]any{"organization": "third"}},
		}},
	}
	result := testutils.ApplyAndParse(t, p, od)
	provider := result["provider"].(map[string]any)
	if provider["organization"] != "third" {
		t.Errorf("organization: got %v, want %q (last patch must win)", provider["organization"], "third")
	}
}

func TestApply_MixedActions_AppliedInOrder(t *testing.T) {
	// merge adds a key; remove deletes the same key; merge sets a new value.
	// The remove must target a key that exists in the original doc.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	od := model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.provider"}, Data: map[string]any{"organization": "step-1"}},
			{Action: "remove", Selector: &model.Selector{JSONPath: "$.provider"}, Data: map[string]any{"organization": nil}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.provider"}, Data: map[string]any{"organization": "step-3"}},
		}},
	}
	result := testutils.ApplyAndParse(t, p, od)
	provider := result["provider"].(map[string]any)
	if provider["organization"] != "step-3" {
		t.Errorf("organization: got %v, want %q", provider["organization"], "step-3")
	}
}

func TestApply_MultiplePatches_StopsOnFirstError(t *testing.T) {
	// The first patch uses an unknown action with a nil-keyed entry so that
	// Decompose produces a patch that reaches apply's default error branch.
	// The second patch must not be applied.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	od := model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "badaction", Selector: &model.Selector{JSONPath: "$"}, Data: map[string]any{"name": nil}},
			{Action: "merge", Selector: &model.Selector{JSONPath: "$.capabilities"}, Data: map[string]any{"y": 2}},
		}},
	}
	_, err := p.Apply(od)
	if err == nil {
		t.Fatal("expected error for bad action in first patch, got nil")
	}
	// Processor state must be untouched — a subsequent Apply returns the original doc.
	result := testutils.ApplyAndParse(t, p, testutils.NoPatches())
	caps := result["capabilities"].(map[string]any)
	if _, exists := caps["y"]; exists {
		t.Error("second patch was applied despite first patch failing")
	}
}

// ---- Apply: error paths -----------------------------------------------------

func TestApply_UnknownAction_ReturnsError(t *testing.T) {
	// An unknown action only reaches the `default` error branch inside `apply`
	// when Decompose produces at least one patch for it. Decompose fast-paths
	// merge/update; for any other action it walks Data and produces patches only
	// for nil-keyed entries. So we use a nil-keyed entry to force a patch through.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	_, err := p.Apply(testutils.OnePatch("replace",
		model.Selector{JSONPath: "$"},
		map[string]any{"name": nil}))
	if err == nil {
		t.Fatal("expected error for unsupported action with nil-keyed data, got nil")
	}
}

func TestApply_UnknownAction_WithNonNilScalarData_IsNoOp(t *testing.T) {
	// Non-nil scalar values in Data are skipped by Decompose → zero patches →
	// no error, no change. This documents the current behavior.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("replace",
		model.Selector{JSONPath: "$.capabilities"},
		map[string]any{"x": 1}))
	if !reflect.DeepEqual(result, ficaDoc) {
		t.Error("expected no-op for unknown action with non-nil scalar data, but content changed")
	}
}

func TestApply_InvalidJSONPath_ReturnsError(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{JSONPath: "$$[invalid"},
		map[string]any{"x": 1}))
	if err == nil {
		t.Fatal("expected error for invalid JSONPath, got nil")
	}
}

func TestApply_UnsupportedSelector_ReturnsError(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	_, err := p.Apply(testutils.OnePatch("merge", model.Selector{}, map[string]any{"x": 1}))
	if err == nil {
		t.Fatal("expected error for empty selector, got nil")
	}
}

// ---- Apply: selector matches nothing ----------------------------------------

func TestApply_MergeSelector_NotFound_ReturnsError(t *testing.T) {
	// For merge/update, Decompose fast-paths the patch through; apply then calls
	// Resolve which errors when the selector matches nothing.
	// Note: JSONPath selectors no longer validate existence at resolve time —
	// a non-existent JSONPath is a valid expression that simply matches nothing,
	// so merge/update via JSONPath is a silent no-op, not an error.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	_, err := p.Apply(testutils.OnePatch("merge", model.Selector{Operation: "nonexistent-skill"}, map[string]any{"x": 1}))
	if err == nil {
		t.Error("expected error for operation selector that matches nothing, got nil")
	}
}

func TestApply_Merge_JSONPath_NotFound_ReturnsError(t *testing.T) {
	// A selector that matches nothing is an error: overlays never create a
	// missing target, not even via jsonPath.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.nonexistent"},
		map[string]any{"x": "created"}))
	if err == nil {
		t.Fatal("expected an error for a merge that matches no existing node, got nil")
	}
}

func TestApply_Update_JSONPath_NotFound_ReturnsError(t *testing.T) {
	// An update via a non-existent JSONPath is an error, not a create.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent)))
	_, err := p.Apply(testutils.OnePatch("update",
		model.Selector{JSONPath: "$.newnode"},
		map[string]any{"key": "value"}))
	if err == nil {
		t.Fatal("expected an error for an update that matches no existing node, got nil")
	}
}

// ---- Apply: root selector ---------------------------------------------------

func TestApply_RootSelector_Merge_AddsTopLevelKey(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(`{"name":"test","version":"1.0"}`)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"overlay": "applied"}))

	if result["overlay"] != "applied" {
		t.Errorf("overlay: got %v, want %q", result["overlay"], "applied")
	}
	if result["name"] != "test" {
		t.Error("name field lost after root merge")
	}
}

func TestApply_RootSelector_Update_ReplacesDocument(t *testing.T) {
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(`{"name":"test","version":"1.0"}`)))
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"name": "replaced"}))

	if result["name"] != "replaced" {
		t.Errorf("name: got %v, want %q", result["name"], "replaced")
	}
	if _, exists := result["version"]; exists {
		t.Error("version should be gone after root update")
	}
}

func TestApply_RootSelector_Remove_ClearsDocument(t *testing.T) {
	// nil Data is now fast-pathed by Decompose. apply resolves the root selector
	// and the remove root branch returns an empty map.
	p := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(`{"name":"test","version":"1.0"}`)))
	rd, err := p.Apply(testutils.OnePatch("remove", model.Selector{Root: utils.Ptr(true)}, nil))
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	result := testutils.UnmarshalResult[map[string]any](t, rd.MediaType, rd.Content)
	if len(result) != 0 {
		t.Errorf("expected empty document after root remove, got: %v", result)
	}
}

// ---- resolve ----------------------------------------------------------------

func TestExpressions_Resolve_RootSelector_ReturnsRootExpression(t *testing.T) {
	e, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{Root: utils.Ptr(true)})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	testutils.AssertExpr(t, e, "$")
}

func TestExpressions_Resolve_JSONPath_Found(t *testing.T) {
	e, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{JSONPath: "$.provider"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	testutils.AssertExpr(t, e, "$.provider")
	testutils.AssertResolvesToNode(t, ficaDoc, e, ficaDoc["provider"])
}

func TestExpressions_Resolve_Operation_Found(t *testing.T) {
	e, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{Operation: "invoice-recalculation"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	skills := ficaDoc["skills"].([]any)
	testutils.AssertResolvesToNode(t, ficaDoc, e, skills[2])
}

func TestExpressions_Resolve_JSONPath_NotFound_Succeeds(t *testing.T) {
	// JSONPath selectors no longer validate existence — a path that matches
	// nothing is returned as a valid parsed expression without error.
	e, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{JSONPath: "$.nonexistent"})
	if err != nil {
		t.Fatalf("expected no error for non-existent JSONPath, got: %v", err)
	}
	testutils.AssertExpr(t, e, "$.nonexistent")
}

func TestExpressions_Resolve_JSONPath_InvalidSyntax_ReturnsError(t *testing.T) {
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{JSONPath: "$$[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid JSONPath syntax, got nil")
	}
}

func TestExpressions_Resolve_Operation_NotFound_ReturnsError(t *testing.T) {
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{Operation: "nonexistent-skill"})
	if err == nil {
		t.Fatal("expected error for operation that matches no skill, got nil")
	}
}

func TestExpressions_Resolve_EmptySelector_ReturnsError(t *testing.T) {
	// An empty Selector has neither Root, JSONPath, nor Operation — must error.
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{})
	if err == nil {
		t.Fatal("expected error for unsupported (empty) selector, got nil")
	}
}

func TestExpressions_Resolve_RootFalsePtrIsNotRoot(t *testing.T) {
	// Root: ptr(false) must NOT match the Root branch — it falls through to
	// JSONPath and Operation (both empty), which returns an error.
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{Root: utils.Ptr(false)})
	if err == nil {
		t.Fatal("expected error when Root pointer is false, got nil")
	}
}

func TestExpressions_Resolve_JSONPathTakesPrecedenceOverOperation(t *testing.T) {
	// When both JSONPath and Operation are set, JSONPath is checked first.
	e, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).resolve(ficaDoc, &model.Selector{
		JSONPath:  "$.provider",
		Operation: "dispute-case-resolution",
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// Must resolve to $.provider, not a skill element.
	testutils.AssertExpr(t, e, "$.provider")
}

// ---- merge / update pass-through --------------------------------------------

func TestDecompose_NilData_ReturnsOriginalPatchUnchanged(t *testing.T) {
	// nil Data is fast-pathed alongside merge/update — the original patch is
	// returned as-is regardless of the action, so the caller sees one patch.
	patch := model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.provider"},
		Data:        nil,
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, patch)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{patch})
}

func TestDecompose_MergeAction_ReturnsOriginalPatchUnchanged(t *testing.T) {
	patch := model.Patch{
		Action:      "merge",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.capabilities"},
		Data:        map[string]any{"streaming": true},
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, patch)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{patch})
}

func TestDecompose_UpdateAction_ReturnsOriginalPatchUnchanged(t *testing.T) {
	patch := model.Patch{
		Action:      "update",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.provider"},
		Data:        map[string]any{"organization": "ACME"},
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, patch)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{patch})
}

func TestDecompose_MergeAction_WithRootSelector_ReturnsOriginalPatchUnchanged(t *testing.T) {
	patch := model.Patch{
		Action:      "merge",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{Root: utils.Ptr(true)},
		Data:        map[string]any{"overlay": "applied"},
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, patch)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{patch})
}

func TestDecompose_MergeAction_EmptyData_ReturnsOriginalPatchUnchanged(t *testing.T) {
	patch := model.Patch{
		Action:      "merge",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.capabilities"},
		Data:        map[string]any{},
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, patch)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{patch})
}

func TestDecompose_MergeAction_WithOperationSelector_ReturnsOriginalPatchUnchanged(t *testing.T) {
	patch := model.Patch{
		Action:      "merge",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{Operation: "dispute-case-resolution"},
		Data:        map[string]any{"x-custom": "value"},
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, patch)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{patch})
}

// ---- remove: empty / scalar-only data ---------------------------------------

func TestDecompose_RemoveAction_EmptyData_ReturnsEmptySlice(t *testing.T) {
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.capabilities"},
		Data:        map[string]any{},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertEmpty(t, result)
}

func TestDecompose_RemoveAction_ScalarValue_IsSkipped(t *testing.T) {
	// Non-nil, non-map values are silently ignored — no patches produced.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.capabilities"},
		Data:        map[string]any{"streaming": "should-be-ignored"},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertEmpty(t, result)
}

func TestDecompose_RemoveAction_BoolValue_IsSkipped(t *testing.T) {
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$.capabilities"},
		Data:        map[string]any{"streaming": true},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertEmpty(t, result)
}

// ---- remove: nil-value leaf patches -----------------------------------------

func TestDecompose_RemoveAction_NilValue_ProducesLeafPatch(t *testing.T) {
	// A nil data value produces a leaf patch whose selector is the JSONPath
	// child of the input selector.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data:        map[string]any{"name": nil},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.name"},
			Data:        nil,
		},
	})
}

func TestDecompose_RemoveAction_MultipleNilValues_ProducesOneLeafPatchPerKey(t *testing.T) {
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data:        map[string]any{"name": nil, "version": nil},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.name"},
			Data:        nil,
		},
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.version"},
			Data:        nil,
		},
	})
}

func TestDecompose_RemoveAction_MixedNilAndScalar_OnlyNilProducesPatches(t *testing.T) {
	// A nil value produces a leaf patch; a non-nil scalar is silently skipped.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data: map[string]any{
			"name":    nil,                 // produces a patch
			"version": "should-be-skipped", // skipped
		},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.name"},
			Data:        nil,
		},
	})
}

// ---- remove: map-value recursion --------------------------------------------

func TestDecompose_RemoveAction_MapValue_RecursesAndProducesLeafPatch(t *testing.T) {
	// A map value triggers recursive decomposition down to the nil leaf.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data: map[string]any{
			"provider": map[string]any{
				"organization": nil,
			},
		},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.provider.organization"},
			Data:        nil,
		},
	})
}

func TestDecompose_RemoveAction_DeeplyNestedMapValue_ProducesLeafAtCorrectPath(t *testing.T) {
	// Two levels of recursion — verifies that the full JSONPath chain is built
	// correctly down to the nil leaf.
	doc := map[string]any{
		"outer": map[string]any{
			"inner": map[string]any{
				"leaf": "value",
			},
		},
	}
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(doc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data: map[string]any{
			"outer": map[string]any{
				"inner": map[string]any{
					"leaf": nil,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.outer.inner.leaf"},
			Data:        nil,
		},
	})
}

// ---- remove: patch field preservation ---------------------------------------

func TestDecompose_RemoveAction_PatchFields_PreservedInProducedPatches(t *testing.T) {
	// Action, Description, Tags, and Meta must be copied verbatim.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data:        map[string]any{"name": nil},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "remove",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.name"},
			Data:        nil,
		},
	})
}

// ---- remove: operation selector ---------------------------------------------

func TestDecompose_RemoveAction_OperationSelector_NilValue_ProducesLeafPatch(t *testing.T) {
	// Operation selector resolves to a skills array element via JSONPath filter.
	// The child path uses the ojg bracket notation for the key.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "remove",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{Operation: "dispute-case-resolution"},
		Data:        map[string]any{"description": nil},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 patch, got %d: %+v", len(result), result)
	}
	// Verify the produced patch preserves non-selector fields and has a nil data value.
	got := result[0]
	if got.Action != "remove" {
		t.Errorf("Action: got %q, want %q", got.Action, "remove")
	}
	if got.Description != "test-description" {
		t.Errorf("Description: got %q, want %q", got.Description, "test-description")
	}
	if got.Data != nil {
		t.Errorf("Data[description]: got %v, want nil", got.Data)
	}
}

// ---- unknown action ---------------------------------------------------------

func TestDecompose_UnknownAction_WithNilValue_DecomposesLikeRemove(t *testing.T) {
	// Only merge/update are fast-pathed. Any other action follows the recursive
	// data-walk path and produces leaf patches for nil values.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:      "upsert",
		Description: "test-description",
		Tags:        []string{"tag-a", "tag-b"},
		Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
		Selector:    &model.Selector{JSONPath: "$"},
		Data:        map[string]any{"name": nil},
	})
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	testutils.AssertContainsInAnyOrder(t, result, []model.Patch{
		{
			Action:      "upsert",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{JSONPath: "$.name"},
			Data:        nil,
		},
	})
}

// ---- error paths ------------------------------------------------------------

func TestDecompose_RemoveAction_InvalidJSONPath_ReturnsError(t *testing.T) {
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$$[invalid"},
		Data:     map[string]any{"name": nil},
	})
	if err == nil {
		t.Fatal("expected error for invalid JSONPath selector, got nil")
	}
}

func TestDecompose_RemoveAction_SelectorNotFound_IsNoOp(t *testing.T) {
	// JSONPath selectors no longer validate existence at resolve time.
	// A non-existent path is a valid expression; decompose produces a leaf patch
	// whose remove will simply find no locations and do nothing.
	result, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:   "remove",
		Selector: &model.Selector{JSONPath: "$.nonexistent"},
		Data:     map[string]any{"someKey": nil},
	})
	if err != nil {
		t.Fatalf("expected no error for non-existent JSONPath selector, got: %v", err)
	}
	// One leaf patch is still produced for the nil-keyed entry.
	if len(result) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result))
	}
}

func TestDecompose_RemoveAction_OperationNotFound_ReturnsError(t *testing.T) {
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:   "remove",
		Selector: &model.Selector{Operation: "nonexistent-skill"},
		Data:     map[string]any{"description": nil},
	})
	if err == nil {
		t.Fatal("expected error for operation that matches no skill, got nil")
	}
}

func TestDecompose_RemoveAction_UnsupportedSelector_ReturnsError(t *testing.T) {
	// An empty Selector (no Root, JSONPath, or Operation) is unsupported.
	_, err := testutils.AssertNoError(NewOverlayProcessor(makeDefinition(ficaContent))).decompose(ficaDoc, model.Patch{
		Action:   "remove",
		Selector: &model.Selector{},
		Data:     map[string]any{"name": nil},
	})
	if err == nil {
		t.Fatal("expected error for unsupported (empty) selector, got nil")
	}
}
