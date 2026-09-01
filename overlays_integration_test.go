//go:build integration

package overlays

import (
	"fmt"
	"strings"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

func TestApply_EDMX_ODataV2_ReturnsUnsupportedErrorWithV4Workaround(t *testing.T) {
	definition := model.ResourceDefinition{
		DefinitionType: "edmx",
		MediaType:      "application/xml",
		Content: `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="1.0" xmlns:edmx="http://schemas.microsoft.com/ado/2007/06/edmx" xmlns:m="http://schemas.microsoft.com/ado/2007/08/dataservices/metadata">
  <edmx:DataServices m:DataServiceVersion="2.0">
    <Schema Namespace="Svc" xmlns="http://schemas.microsoft.com/ado/2008/09/edm">
      <EntityType Name="Book"/>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`,
	}
	overlay := testutils.OnePatch(
		"merge",
		model.Selector{EntityType: "Svc.Book"},
		map[string]any{"@Core.Description": "A book"},
	)
	overlay.Overlay.Target = &model.Target{}

	_, err := Apply(definition, []model.OverlayDefinition{overlay})
	if err == nil {
		t.Fatal("expected OData v2 EDMX to be rejected")
	}
	if message := err.Error(); !strings.Contains(message, "OData v2 EDMX") ||
		!strings.Contains(message, "OData v4 EDMX") {
		t.Fatalf("expected v2 rejection and v4 workaround hint, got: %v", err)
	}
}

func TestApply_EDMX_RecognizesODataV2FromCSDLNamespace(t *testing.T) {
	definition := model.ResourceDefinition{
		DefinitionType: "edmx",
		MediaType:      "application/xml",
		Content: `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="1.0" xmlns:edmx="http://schemas.microsoft.com/ado/2007/06/edmx">
  <edmx:DataServices>
    <Schema Namespace="Svc" xmlns="http://schemas.microsoft.com/ado/2008/09/edm">
      <EntityType Name="Book"/>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`,
	}
	overlay := testutils.OnePatch(
		"merge",
		model.Selector{EntityType: "Svc.Book"},
		map[string]any{"@Core.Description": "A book"},
	)
	overlay.Overlay.Target = &model.Target{}

	_, err := Apply(definition, []model.OverlayDefinition{overlay})
	if err == nil || !strings.Contains(err.Error(), "OData v2 EDMX") {
		t.Fatalf("expected OData v2 rejection, got: %v", err)
	}
}

// ---- helpers ----------------------------------------------------------------

// applyOne applies a single matching overlay via the top-level Apply function
// and returns the parsed result document. Fatals if Apply returns an error or
// does not return exactly one result.
//
// A non-nil Target is required by IsApplicable; callers that do not set one on
// their OverlayDefinition get an empty Target injected here so the overlay is
// always considered applicable by target constraints.
func applyOne(t *testing.T, def model.ResourceDefinition, od model.OverlayDefinition) map[string]any {
	t.Helper()
	if od.Overlay.Target == nil {
		od.Overlay.Target = &model.Target{}
	}
	results, err := Apply(def, []model.OverlayDefinition{od})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Apply: expected 1 result, got %d", len(results))
	}
	return testutils.UnmarshalResult[map[string]any](t, def.MediaType, results[0].Content)
}

// defFor builds a ResourceDefinition from a fixture file and media type.
func defFor(fixturePath, mediaType string) model.ResourceDefinition {
	return model.ResourceDefinition{
		Content:   testutils.LoadFixture(fixturePath),
		MediaType: mediaType,
	}
}

// loadExpected loads and parses a golden-file fixture.
func loadExpected(path string) map[string]any {
	return testutils.UnmarshalFixture[map[string]any](path)
}

// ============================================================================
// IsApplicable — filtering logic
// ============================================================================

// TestIsApplicable_NilTarget_IsNotApplicable verifies that an overlay with a
// nil Target is never applicable, regardless of definition fields.
func TestIsApplicable_NilTarget_IsNotApplicable(t *testing.T) {
	def := model.ResourceDefinition{URL: "https://example.com/api.json", OrdID: "sap.x:api:1.0.0", Perspective: "system-instance"}
	od := model.OverlayDefinition{Overlay: model.Overlay{}}
	if IsApplicable(def, od) {
		t.Error("expected overlay with nil Target to not be applicable")
	}
}

// TestIsApplicable_URLMatch_IsApplicable verifies a matching URL passes.
func TestIsApplicable_URLMatch_IsApplicable(t *testing.T) {
	def := model.ResourceDefinition{URL: "https://example.com/api.json"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{URL: "https://example.com/api.json"}}}
	if !IsApplicable(def, od) {
		t.Error("expected overlay to be applicable when URL matches")
	}
}

// TestIsApplicable_URLMismatch_IsNotApplicable verifies a mismatched URL is rejected.
func TestIsApplicable_URLMismatch_IsNotApplicable(t *testing.T) {
	def := model.ResourceDefinition{URL: "https://example.com/api.json"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{URL: "https://other.com/api.json"}}}
	if IsApplicable(def, od) {
		t.Error("expected overlay to not be applicable when URL does not match")
	}
}

// TestIsApplicable_OrdIDMatch_IsApplicable verifies a matching OrdID passes.
func TestIsApplicable_OrdIDMatch_IsApplicable(t *testing.T) {
	def := model.ResourceDefinition{OrdID: "sap.x:api:1.0.0"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{OrdID: "sap.x:api:1.0.0"}}}
	if !IsApplicable(def, od) {
		t.Error("expected overlay to be applicable when OrdID matches")
	}
}

// TestIsApplicable_OrdIDMismatch_IsNotApplicable verifies a mismatched OrdID is rejected.
func TestIsApplicable_OrdIDMismatch_IsNotApplicable(t *testing.T) {
	def := model.ResourceDefinition{OrdID: "sap.x:api:1.0.0"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{OrdID: "sap.x:api:2.0.0"}}}
	if IsApplicable(def, od) {
		t.Error("expected overlay to not be applicable when OrdID does not match")
	}
}

// TestIsApplicable_PerspectiveMatch_IsApplicable verifies a matching Perspective passes.
func TestIsApplicable_PerspectiveMatch_IsApplicable(t *testing.T) {
	def := model.ResourceDefinition{Perspective: "system-instance"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{}, Perspective: "system-instance"}}
	if !IsApplicable(def, od) {
		t.Error("expected overlay to be applicable when Perspective matches")
	}
}

// TestIsApplicable_PerspectiveMismatch_IsNotApplicable verifies a mismatched Perspective is rejected.
func TestIsApplicable_PerspectiveMismatch_IsNotApplicable(t *testing.T) {
	def := model.ResourceDefinition{Perspective: "system-instance"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{}, Perspective: "system-type"}}
	if IsApplicable(def, od) {
		t.Error("expected overlay to not be applicable when Perspective does not match")
	}
}

// TestIsApplicable_DefinitionTypeMatch_IsApplicable verifies a matching DefinitionType passes.
func TestIsApplicable_DefinitionTypeMatch_IsApplicable(t *testing.T) {
	def := model.ResourceDefinition{DefinitionType: "openapi-v3"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{DefinitionType: "openapi-v3"}}}
	if !IsApplicable(def, od) {
		t.Error("expected overlay to be applicable when DefinitionType matches")
	}
}

// TestIsApplicable_DefinitionTypeMismatch_IsNotApplicable verifies a mismatched DefinitionType is rejected.
func TestIsApplicable_DefinitionTypeMismatch_IsNotApplicable(t *testing.T) {
	def := model.ResourceDefinition{DefinitionType: "openapi-v3"}
	od := model.OverlayDefinition{Overlay: model.Overlay{Target: &model.Target{DefinitionType: "asyncapi-v2"}}}
	if IsApplicable(def, od) {
		t.Error("expected overlay to not be applicable when DefinitionType does not match")
	}
}

// TestIsApplicable_AllConstraintsMatch_IsApplicable verifies that an overlay
// with every constraint set passes when all match.
func TestIsApplicable_AllConstraintsMatch_IsApplicable(t *testing.T) {
	def := model.ResourceDefinition{
		URL:            "https://example.com/api.json",
		OrdID:          "sap.x:api:1.0.0",
		Perspective:    "system-instance",
		DefinitionType: "openapi-v3",
	}
	od := model.OverlayDefinition{Overlay: model.Overlay{
		Perspective: "system-instance",
		Target: &model.Target{
			URL:            "https://example.com/api.json",
			OrdID:          "sap.x:api:1.0.0",
			DefinitionType: "openapi-v3",
		},
	}}
	if !IsApplicable(def, od) {
		t.Error("expected overlay to be applicable when all constraints match")
	}
}

// TestIsApplicable_PartialMatch_OneConstraintFails_IsNotApplicable verifies that
// a single failing constraint rejects an otherwise matching overlay.
func TestIsApplicable_PartialMatch_OneConstraintFails_IsNotApplicable(t *testing.T) {
	def := model.ResourceDefinition{
		URL:   "https://example.com/api.json",
		OrdID: "sap.x:api:1.0.0",
	}
	od := model.OverlayDefinition{Overlay: model.Overlay{
		Target: &model.Target{
			URL:   "https://example.com/api.json",
			OrdID: "sap.x:api:2.0.0", // mismatch
		},
	}}
	if IsApplicable(def, od) {
		t.Error("expected overlay to not be applicable when one constraint does not match")
	}
}

// ============================================================================
// Apply — filtering behaviour
// ============================================================================

// TestApply_EmptyOverlayList_ReturnsEmptyResults verifies no results for an empty list.
func TestApply_EmptyOverlayList_ReturnsEmptyResults(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	results, err := Apply(def, []model.OverlayDefinition{})
	if err != nil {
		t.Fatalf("Apply: unexpected error: %v", err)
	}
	testutils.AssertEmpty(t, results)
}

// TestApply_NoMatchingOverlays_ReturnsEmptyResults verifies that overlays
// filtered out by IsApplicable produce no results.
func TestApply_NoMatchingOverlays_ReturnsEmptyResults(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	def.URL = "https://example.com/api.json"

	od := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x": 1})
	od.Overlay.Target = &model.Target{URL: "https://other.com/api.json"}

	results, err := Apply(def, []model.OverlayDefinition{od})
	if err != nil {
		t.Fatalf("Apply: unexpected error: %v", err)
	}
	testutils.AssertEmpty(t, results)
}

// TestApply_MultipleOverlays_OnlyMatchingApplied verifies that among several
// overlays only those passing IsApplicable produce results.
func TestApply_MultipleOverlays_OnlyMatchingApplied(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	def.URL = "https://example.com/api.json"

	matching := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x": 1})
	matching.Overlay.Target = &model.Target{URL: "https://example.com/api.json"}

	nonMatching := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x": 2})
	nonMatching.Overlay.Target = &model.Target{URL: "https://other.com/api.json"}

	results, err := Apply(def, []model.OverlayDefinition{matching, nonMatching})
	if err != nil {
		t.Fatalf("Apply: unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

// TestApply_MultipleMatchingOverlays_AllProduceResults verifies that each
// matching overlay produces its own independent result entry.
func TestApply_MultipleMatchingOverlays_AllProduceResults(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")

	od1 := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x-first": true})
	od1.Overlay.Target = &model.Target{}
	od2 := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x-second": true})
	od2.Overlay.Target = &model.Target{}

	results, err := Apply(def, []model.OverlayDefinition{od1, od2})
	if err != nil {
		t.Fatalf("Apply: unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

// TestApply_OriginalDefinitionNotMutated verifies that Apply never modifies
// the input ResourceDefinition.
func TestApply_OriginalDefinitionNotMutated(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	originalContent := def.Content
	od := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x-applied": true})
	od.Overlay.Target = &model.Target{}
	if _, err := Apply(def, []model.OverlayDefinition{od}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if def.Content != originalContent {
		t.Error("Apply must not mutate the original ResourceDefinition content")
	}
}

// TestApply_UnsupportedMediaType_ReturnsError verifies that an unrecognised
// media type surfaces an error rather than panicking.
func TestApply_UnsupportedMediaType_ReturnsError(t *testing.T) {
	def := model.ResourceDefinition{Content: `{}`, MediaType: "application/x-unknown"}
	od := testutils.OnePatch("merge", model.Selector{Root: utils.Ptr(true)}, map[string]any{"x": 1})
	od.Overlay.Target = &model.Target{}
	if _, err := Apply(def, []model.OverlayDefinition{od}); err == nil {
		t.Error("expected an error for unsupported media type, got nil")
	}
}

// TestApply_UnsupportedSelector_ReturnsError verifies that an unsupported selector
// on a valid processor surfaces an error from proc.Apply rather than panicking.
func TestApply_UnsupportedSelector_ReturnsError(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	od := testutils.OnePatch("merge", model.Selector{EntityType: "Unsupported"}, map[string]any{"x": 1})
	od.Overlay.Target = &model.Target{}
	if _, err := Apply(def, []model.OverlayDefinition{od}); err == nil {
		t.Error("expected an error for unsupported selector on JSON processor, got nil")
	}
}

// ============================================================================
// Apply — JSON (MCP server card) processor path
// ============================================================================

// TestApply_JSON_Merge_Root_AddsStateInfo exercises the full dispatch chain
// through the JSON processor with a root merge, comparing against the golden
// file already used by the json package's own integration test.
func TestApply_JSON_Merge_Root_AddsStateInfo(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/json/testdata/integration/merge_root_expected.json"),
		applyOne(t, def, testutils.OnePatch(
			"merge",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"x-sap-stateInfo": map[string]any{
					"state":           "Active",
					"deprecationDate": "2026-12-31",
				},
			},
		)),
	)
}

// TestApply_JSON_Update_Root_ReplacesEntireDocument verifies full-document
// replacement via the JSON processor path.
func TestApply_JSON_Update_Root_ReplacesEntireDocument(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/json/testdata/integration/update_root_expected.json"),
		applyOne(t, def, testutils.OnePatch(
			"update",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{"name": "replaced", "version": "0.0.1"},
		)),
	)
}

// TestApply_JSON_Remove_JSONPath_RemovesPrompts verifies targeted field
// removal via the JSON processor path.
func TestApply_JSON_Remove_JSONPath_RemovesPrompts(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/json/testdata/integration/remove_jsonpath_expected.json"),
		applyOne(t, def, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.prompts"},
			nil,
		)),
	)
}

// TestApply_JSON_MultiPatch_RealisticOverlaySequence verifies that a sequence
// of four patches is applied in order through the full dispatch chain.
func TestApply_JSON_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	def := defFor("internal/processor/json/testdata/mcp_server_card.json", "application/json")
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/json/testdata/integration/multi_patch_expected.json"),
		applyOne(t, def, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{JSONPath: "$"},
					Data: map[string]any{
						"version":      "1.1.0",
						"x-sap-ord-id": "sap.bookshop:mcpServerCard:catalogservice:v1",
					},
				},
				{
					Action:   "update",
					Selector: &model.Selector{JSONPath: "$._meta"},
					Data: map[string]any{
						"sap/category": "AI Tools",
						"sap/owner":    "team-ai-services",
					},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{JSONPath: "$.prompts"},
					Data:     nil,
				},
				{
					Action:   "merge",
					Selector: &model.Selector{Root: utils.Ptr(true)},
					Data:     map[string]any{"x-sap-stateInfo": map[string]any{"state": "Active"}},
				},
			}},
		}),
	)
}

// ============================================================================
// Apply — OpenAPI processor path (JSON + YAML)
// ============================================================================

// openAPIDef builds a ResourceDefinition for the Petstore fixture in the given format,
// with DefinitionType set so CreateFor routes to the OpenAPI processor.
func openAPIDef(format string) model.ResourceDefinition {
	return model.ResourceDefinition{
		Content:        testutils.LoadFixture(fmt.Sprintf("internal/processor/openapi/testdata/petstore.%s", format)),
		MediaType:      fmt.Sprintf("application/%s", format),
		DefinitionType: "openapi-v3",
	}
}

// csnDef builds a ResourceDefinition for the CSN flight model fixture,
// with DefinitionType set so CreateFor routes to the CSN processor.
func csnDef() model.ResourceDefinition {
	return model.ResourceDefinition{
		Content:        testutils.LoadFixture("internal/processor/csn/testdata/flight_model.json"),
		MediaType:      "application/json",
		DefinitionType: "sap-csn-interop-effective-v1",
	}
}

// TestApply_OpenAPI_Merge_Root_AddsStateInfo exercises the OpenAPI processor
// path in both JSON and YAML formats with a root merge.
func TestApply_OpenAPI_Merge_Root_AddsStateInfo(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			def := openAPIDef(format)
			testutils.AssertDeepEquals(
				t,
				loadExpected(fmt.Sprintf("internal/processor/openapi/testdata/integration/merge_root_expected.%s", format)),
				applyOne(t, def, testutils.OnePatch(
					"merge",
					model.Selector{Root: utils.Ptr(true)},
					map[string]any{
						"x-sap-stateInfo": map[string]any{
							"state":           "Active",
							"deprecationDate": "2026-12-31",
						},
					},
				)),
			)
		})
	}
}

// TestApply_OpenAPI_Merge_JSONPath_UpdatesInfoVersion bumps the API version
// and adds an x-sap-ord-id via the OpenAPI processor in both formats.
func TestApply_OpenAPI_Merge_JSONPath_UpdatesInfoVersion(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			def := openAPIDef(format)
			testutils.AssertDeepEquals(
				t,
				loadExpected(fmt.Sprintf("internal/processor/openapi/testdata/integration/merge_jsonpath_info_expected.%s", format)),
				applyOne(t, def, testutils.OnePatch(
					"merge",
					model.Selector{JSONPath: "$.info"},
					map[string]any{
						"version":      "2.0.0",
						"x-sap-ord-id": "sap.petstore:apiResource:petstore:v2",
					},
				)),
			)
		})
	}
}

// TestApply_OpenAPI_Merge_Operation_UpdatesListPetsDescription exercises the
// concept-level Operation selector through the OpenAPI processor.
func TestApply_OpenAPI_Merge_Operation_UpdatesListPetsDescription(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			def := openAPIDef(format)
			testutils.AssertDeepEquals(
				t,
				loadExpected(fmt.Sprintf("internal/processor/openapi/testdata/integration/merge_operation_expected.%s", format)),
				applyOne(t, def, testutils.OnePatch(
					"merge",
					model.Selector{Operation: "listPets"},
					map[string]any{
						"description":  "Returns a paginated list of all pets. Deprecated: use /v2/pets instead.",
						"x-deprecated": true,
					},
				)),
			)
		})
	}
}

// TestApply_OpenAPI_Remove_Operation_RemovesCreatePets verifies concept-level
// operation removal through the OpenAPI processor.
func TestApply_OpenAPI_Remove_Operation_RemovesCreatePets(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			def := openAPIDef(format)
			testutils.AssertDeepEquals(
				t,
				loadExpected(fmt.Sprintf("internal/processor/openapi/testdata/integration/remove_operation_expected.%s", format)),
				applyOne(t, def, testutils.OnePatch(
					"remove",
					model.Selector{Operation: "createPets"},
					nil,
				)),
			)
		})
	}
}

// TestApply_OpenAPI_MultiPatch_RealisticOverlaySequence verifies a five-patch
// sequence through the full OpenAPI dispatch chain in both formats.
func TestApply_OpenAPI_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			def := openAPIDef(format)
			testutils.AssertDeepEquals(
				t,
				loadExpected(fmt.Sprintf("internal/processor/openapi/testdata/integration/multi_patch_expected.%s", format)),
				applyOne(t, def, model.OverlayDefinition{
					Overlay: model.Overlay{Patches: []model.Patch{
						{
							Action:   "merge",
							Selector: &model.Selector{JSONPath: "$.info"},
							Data: map[string]any{
								"version":      "1.1.0",
								"x-sap-ord-id": "sap.petstore:apiResource:petstore:v1",
							},
						},
						{
							Action:   "update",
							Selector: &model.Selector{JSONPath: "$.servers"},
							Data: []any{
								map[string]any{
									"url":         "https://api.petstore.example/v1",
									"description": "Production",
								},
							},
						},
						{
							Action:   "remove",
							Selector: &model.Selector{JSONPath: "$.externalDocs"},
							Data:     nil,
						},
						{
							Action:   "merge",
							Selector: &model.Selector{Operation: "listPets"},
							Data: map[string]any{
								"description":  "Returns a paginated list of all pets. Deprecated: use /v2/pets instead.",
								"x-deprecated": true,
							},
						},
						{
							Action:   "merge",
							Selector: &model.Selector{Operation: "showPetById", Parameter: "petId"},
							Data: map[string]any{
								"description": "The unique identifier of the pet. Must be a non-empty string.",
								"example":     "pet-42",
							},
						},
					}},
				}),
			)
		})
	}
}

// ============================================================================
// Apply — CSN processor path
// ============================================================================

// TestApply_CSN_Merge_EntityType_AddsRepresentativeKey exercises the concept-
// level EntityType selector through the CSN processor dispatch path.
func TestApply_CSN_Merge_EntityType_AddsRepresentativeKey(t *testing.T) {
	def := csnDef()
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/csn/testdata/integration/merge_entitytype_expected.json"),
		applyOne(t, def, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "Airline"},
			map[string]any{"@ObjectModel.representativeKey": map[string]any{"=": "AirlineID"}},
		)),
	)
}

// TestApply_CSN_Merge_EntityTypeWithProperty_RenamesElementLabel exercises the
// EntityType + PropertyType selector through the CSN processor dispatch path.
func TestApply_CSN_Merge_EntityTypeWithProperty_RenamesElementLabel(t *testing.T) {
	def := csnDef()
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/csn/testdata/integration/merge_entitytype_property_expected.json"),
		applyOne(t, def, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "Airline", PropertyType: "AirlineID"},
			map[string]any{"@EndUserText.label": "Airline Code"},
		)),
	)
}

// TestApply_CSN_Remove_EntityType_RemovesCountriesTexts verifies entity removal
// through the CSN processor dispatch path.
func TestApply_CSN_Remove_EntityType_RemovesCountriesTexts(t *testing.T) {
	def := csnDef()
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/csn/testdata/integration/remove_entitytype_expected.json"),
		applyOne(t, def, testutils.OnePatch(
			"remove",
			model.Selector{EntityType: "Countries_texts"},
			nil,
		)),
	)
}

// TestApply_CSN_MultiPatch_RealisticOverlaySequence verifies a four-patch
// sequence through the full CSN dispatch chain.
func TestApply_CSN_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	def := csnDef()
	testutils.AssertDeepEquals(
		t,
		loadExpected("internal/processor/csn/testdata/integration/multi_patch_expected.json"),
		applyOne(t, def, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "FlightConnection"},
					Data:     map[string]any{"@EndUserText.label": "Route"},
				},
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "Flight"},
					Data:     map[string]any{"@Core.Description": "A scheduled flight instance on a given date."},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{EntityType: "Countries_texts"},
					Data:     nil,
				},
				{
					Action:   "update",
					Selector: &model.Selector{EntityType: "Airport", PropertyType: "City"},
					Data: map[string]any{
						"@EndUserText.label":  "City",
						"@Semantics.cityName": true,
						"type":                "cds.String",
						"length":              int64(40),
					},
				},
			}},
		}),
	)
}
