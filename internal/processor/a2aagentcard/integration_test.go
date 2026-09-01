//go:build integration

package a2aagentcard

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// integrationInput is the shared minimal agent fixture for all integration tests.
var integrationInput = testutils.LoadFixture("testdata/fica_dispute_agent.json")

// applyIntegration applies the given overlay to the integration input fixture
// and returns the parsed result document.
func applyIntegration(t *testing.T, od model.OverlayDefinition) map[string]any {
	t.Helper()

	definition := model.ResourceDefinition{
		MediaType: "application/json",
		Content:   integrationInput,
	}

	return testutils.ApplyAndParse(t, testutils.AssertNoError(NewOverlayProcessor(definition)), od)
}

// loadExpected loads and parses an expected-output fixture from the integration testdata directory.
func loadExpected(path string) map[string]any {
	return testutils.UnmarshalFixture[map[string]any]("testdata/integration/" + path)
}

// ---- merge: root selector ---------------------------------------------------

// TestIntegration_Merge_Root_AddsTopLevelFields merges two SAP extension fields into
// the root of the document, verifying that all existing content is preserved while
// the new keys are added at the top level.
func TestIntegration_Merge_Root_AddsTopLevelFields(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"x-sap-ord-visibility": "public",
				"x-sap-shortText":      "Dispute resolution via FI-CA",
			},
		)),
	)
}

// ---- update: root selector --------------------------------------------------

// TestIntegration_Update_Root_ReplacesEntireDocument replaces the whole agent document
// with a minimal stub, verifying that nothing from the original survives.
func TestIntegration_Update_Root_ReplacesEntireDocument(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"name":    "Replacement Agent",
				"version": "2.0.0",
			},
		)),
	)
}

// ---- remove: root selector --------------------------------------------------

// TestIntegration_Remove_Root_ClearsDocument removes the root node, producing
// an empty document.
func TestIntegration_Remove_Root_ClearsDocument(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{Root: utils.Ptr(true)},
			nil,
		)),
	)
}

// ---- merge: JSONPath selector -----------------------------------------------

// TestIntegration_Merge_JSONPath_UpdatesNameAndDescription applies two sequential
// JSONPath merge patches to update $.name and $.description independently, verifying
// that all other top-level fields are preserved.
func TestIntegration_Merge_JSONPath_UpdatesNameAndDescription(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_jsonpath_metadata_expected.json"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{JSONPath: "$.name"},
					Data:     "FI-CA Dispute Resolution Agent (External)",
				},
				{
					Action:   "merge",
					Selector: &model.Selector{JSONPath: "$.description"},
					Data:     "Resolves financial disputes automatically using AI.",
				},
			}},
		}),
	)
}

// ---- update: JSONPath selector ----------------------------------------------

// TestIntegration_Update_JSONPath_ReplacesProvider replaces the entire $.provider
// object with a new one that includes a nested contact field that did not exist before.
func TestIntegration_Update_JSONPath_ReplacesProvider(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_jsonpath_provider_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{JSONPath: "$.provider"},
			map[string]any{
				"organization": "ACME Corp",
				"url":          "https://www.acme.example.com",
				"contact": map[string]any{
					"email": "support@acme.example.com",
				},
			},
		)),
	)
}

// ---- remove: JSONPath selector ----------------------------------------------

// TestIntegration_Remove_JSONPath_DeletesTopLevelField removes $.documentationUrl
// from the document, verifying that all sibling fields are preserved.
func TestIntegration_Remove_JSONPath_DeletesTopLevelField(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_jsonpath_field_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.documentationUrl"},
			nil,
		)),
	)
}

// TestIntegration_Remove_JSONPath_DeletesSubkey removes $.capabilities.pushNotifications
// via a nil-keyed data entry targeting the parent object, verifying that the parent
// and its remaining key (streaming) are preserved.
func TestIntegration_Remove_JSONPath_DeletesSubkey(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_jsonpath_subkey_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.capabilities"},
			map[string]any{"pushNotifications": nil},
		)),
	)
}

// ---- merge: JSONPath no-match -----------------------------------------------

// TestIntegration_Merge_JSONPath_NonExistentNode_IsNoOp verifies that a valid
// JSONPath with zero matches does not create the missing node.
func TestIntegration_Merge_JSONPath_NonExistentNode_IsNoOp(t *testing.T) {
	result := applyIntegration(t, testutils.OnePatch(
		"merge",
		model.Selector{JSONPath: "$.licensing"},
		map[string]any{
			"type":    "commercial",
			"contact": "licensing@sap.com",
		},
	))
	if _, exists := result["licensing"]; exists {
		t.Fatal("zero-match JSONPath merge must not create a node")
	}
}

// ---- merge: operation selector ----------------------------------------------

// TestIntegration_Merge_Operation_AddsTagsExamplesAndMetadata merges into the
// dispute-case-resolution skill: appends two new tags, appends a new example,
// and adds an x-sap-visibility field. The invoice-recalculation skill is untouched.
func TestIntegration_Merge_Operation_AddsTagsExamplesAndMetadata(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_operation_skill_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Operation: "dispute-case-resolution"},
			map[string]any{
				"tags":             []any{"dispute-resolution", "credit-memo"},
				"examples":         []any{"Process the incorrect invoice dispute for case ID 12345"},
				"x-sap-visibility": "external",
			},
		)),
	)
}

// ---- update: operation selector ---------------------------------------------

// TestIntegration_Update_Operation_ReplacesSkillEntirely replaces the
// invoice-recalculation skill entirely with a new definition that omits tags and
// examples, verifying that update is a full replacement. The dispute-case-resolution
// skill is untouched.
func TestIntegration_Update_Operation_ReplacesSkillEntirely(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_operation_skill_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{Operation: "invoice-recalculation"},
			map[string]any{
				"id":          "invoice-recalculation",
				"name":        "Replaced Recalculation Skill",
				"description": "This skill has been replaced by the overlay.",
			},
		)),
	)
}

// ---- remove: operation selector ---------------------------------------------

// TestIntegration_Remove_Operation_DeletesSkill removes the dispute-case-resolution
// skill from the skills array entirely, verifying that the invoice-recalculation
// skill remains.
func TestIntegration_Remove_Operation_DeletesSkill(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_operation_skill_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{Operation: "dispute-case-resolution"},
			nil,
		)),
	)
}
