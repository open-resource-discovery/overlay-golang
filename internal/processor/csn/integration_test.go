//go:build integration

package csn

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// integrationInput is the shared flight-model fixture for all integration tests.
var integrationInput = testutils.LoadFixture("testdata/flight_model.json")

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

// TestIntegration_Merge_Root_AddsCreatorToMeta merges a "creator" field into the
// existing "meta" object, verifying that deep-merge preserves "document" and
// "features" while adding the new key.
func TestIntegration_Merge_Root_AddsCreatorToMeta(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{"meta": map[string]any{"creator": "overlay-pipeline"}},
		)),
	)
}

// ---- merge: JSONPath selector -----------------------------------------------

// TestIntegration_Merge_JSONPath_UpdatesExistingLabel renames the "@EndUserText.label"
// on the Airline entity from "Airline" to "Carrier" via a JSONPath merge.
func TestIntegration_Merge_JSONPath_UpdatesExistingLabel(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_jsonpath_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{JSONPath: "$.definitions.Airline"},
			map[string]any{"@EndUserText.label": "Carrier"},
		)),
	)
}

// ---- merge: EntityType selector ---------------------------------------------

// TestIntegration_Merge_EntityType_AddsRepresentativeKey adds an
// "@ObjectModel.representativeKey" annotation to the Airline entity, which does
// not have one in the input.
func TestIntegration_Merge_EntityType_AddsRepresentativeKey(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_entitytype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "Airline"},
			map[string]any{"@ObjectModel.representativeKey": map[string]any{"=": "AirlineID"}},
		)),
	)
}

// ---- merge: EntityType + PropertyType selector ------------------------------

// TestIntegration_Merge_EntityTypeWithProperty_RenamesElementLabel changes the
// "@EndUserText.label" on Airline.AirlineID from "Airline" to "Airline Code",
// verifying that an existing annotation on an element can be overwritten via merge.
func TestIntegration_Merge_EntityTypeWithProperty_RenamesElementLabel(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_entitytype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "Airline", PropertyType: "AirlineID"},
			map[string]any{"@EndUserText.label": "Airline Code"},
		)),
	)
}

// ---- update: root selector --------------------------------------------------

// TestIntegration_Update_Root_ReplacesEntireDocument replaces the whole document
// with a minimal stub, verifying that nothing from the original survives.
func TestIntegration_Update_Root_ReplacesEntireDocument(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{"replaced": true},
		)),
	)
}

// ---- update: JSONPath selector ----------------------------------------------

// TestIntegration_Update_JSONPath_ReplacesElement replaces the Flight.Price element
// entirely with a higher-precision definition, verifying that all prior annotations
// on that element (e.g. @Aggregation.default, @Semantics.amount.currencyCode) are gone.
func TestIntegration_Update_JSONPath_ReplacesElement(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_jsonpath_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{JSONPath: "$.definitions.Flight.elements.Price"},
			map[string]any{
				"@EndUserText.label": "Price",
				"type":               "cds.Decimal",
				"precision":          int64(20),
				"scale":              int64(5),
			},
		)),
	)
}

// ---- update: EntityType selector --------------------------------------------

// TestIntegration_Update_EntityType_ReplacesCountriesEntity replaces the Countries
// entity with an enriched definition that adds a "name" element, verifying that
// update is a full replacement (original "code" element survives only if re-supplied).
func TestIntegration_Update_EntityType_ReplacesCountriesEntity(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_entitytype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{EntityType: "Countries"},
			map[string]any{
				"kind":               "entity",
				"@EndUserText.label": "Countries",
				"elements": map[string]any{
					"code": map[string]any{
						"@EndUserText.label": "Country Code",
						"key":                true,
						"type":               "cds.String",
						"length":             int64(3),
					},
					"name": map[string]any{
						"@EndUserText.label": "Country Name",
						"@Semantics.text":    true,
						"type":               "cds.String",
						"length":             int64(255),
					},
				},
			},
		)),
	)
}

// ---- update: EntityType + PropertyType selector -----------------------------

// TestIntegration_Update_EntityTypeWithProperty_EnrichsCityElement replaces the
// Airport.City element with an enriched definition that adds @Semantics.cityName,
// verifying that the prior element definition is fully replaced.
func TestIntegration_Update_EntityTypeWithProperty_EnrichsCityElement(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_entitytype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{EntityType: "Airport", PropertyType: "City"},
			map[string]any{
				"@EndUserText.label":  "City",
				"@Semantics.cityName": true,
				"type":                "cds.String",
				"length":              int64(40),
			},
		)),
	)
}

// ---- remove: root selector --------------------------------------------------

func TestIntegration_Remove_Root_ReturnsError(t *testing.T) {
	selectors := map[string]model.Selector{
		"root selector": {Root: utils.Ptr(true)},
		"root JSONPath": {JSONPath: "$"},
	}
	for name, selector := range selectors {
		t.Run(name, func(t *testing.T) {
			p := testutils.AssertNoError(NewOverlayProcessor(model.ResourceDefinition{
				Content: integrationInput, MediaType: "application/json",
			}))
			_, err := p.Apply(testutils.OnePatch("remove", selector, nil))
			if err == nil {
				t.Fatal("expected root remove to return an error")
			}
		})
	}
}

// ---- remove: JSONPath selector ----------------------------------------------

// TestIntegration_Remove_JSONPath_RemovesAnnotation removes the
// "@ObjectModel.representativeKey" annotation from FlightConnection via JSONPath,
// verifying targeted removal of a single annotation while the rest of the entity survives.
func TestIntegration_Remove_JSONPath_RemovesAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_jsonpath_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.definitions.FlightConnection['@ObjectModel.representativeKey']"},
			nil,
		)),
	)
}

// ---- remove: EntityType selector --------------------------------------------

// TestIntegration_Remove_EntityType_RemovesCountriesTexts removes the
// Countries_texts entity entirely, verifying that other entities are unaffected.
func TestIntegration_Remove_EntityType_RemovesCountriesTexts(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_entitytype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{EntityType: "Countries_texts"},
			nil,
		)),
	)
}

// ---- remove: EntityType + PropertyType selector -----------------------------

// TestIntegration_Remove_EntityTypeWithProperty_RemovesCountryCodeElement removes
// the Airport.CountryCode_code element, which is a foreign-key helper column that
// a consuming system wishes to hide.
func TestIntegration_Remove_EntityTypeWithProperty_RemovesCountryCodeElement(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_entitytype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{EntityType: "Airport", PropertyType: "CountryCode_code"},
			nil,
		)),
	)
}

// ---- multi-patch: realistic overlay sequence --------------------------------

// TestIntegration_MultiPatch_RealisticOverlaySequence applies a sequence of patches
// that mirrors a real-world overlay: rename FlightConnection's label, add a description
// to Flight, remove the Countries_texts entity, and enrich Airport.City.
// Patches applied:
//  1. merge EntityType "FlightConnection" → renames "@EndUserText.label"
//  2. merge EntityType "Flight" → adds "@Core.Description"
//  3. remove EntityType "Countries_texts" → drops the entity entirely
//  4. update EntityType+PropertyType "Airport.City" → adds @Semantics.cityName
func TestIntegration_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("multi_patch_expected.json"),
		applyIntegration(t, model.OverlayDefinition{
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
