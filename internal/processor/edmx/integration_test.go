//go:build integration

package edmx

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	xml2json "github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// integrationInput is the shared CatalogService EDMX fixture for all integration tests.
var integrationInput = testutils.LoadFixture("testdata/catalogservice.xml")

// normalizeXML parses and re-serializes XML to canonical compact form so that
// pretty-printed golden files and compact processor output compare equal.
func normalizeXML(t *testing.T, raw string) string {
	t.Helper()
	doc, err := xml2json.Convert(raw)
	if err != nil {
		t.Fatalf("normalizeXML: %v", err)
	}
	return doc.ToXML()
}

// applyIntegration applies the given overlay to the integration input fixture
// and returns the normalized result XML string.
func applyIntegration(t *testing.T, od model.OverlayDefinition) string {
	t.Helper()
	p, err := NewOverlayProcessor(model.ResourceDefinition{Content: integrationInput})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	result, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return normalizeXML(t, result.Content)
}

// loadIntegrationExpected loads and normalizes an expected-output XML golden file.
func loadIntegrationExpected(t *testing.T, name string) string {
	t.Helper()
	return normalizeXML(t, testutils.LoadFixture("testdata/integration/"+name))
}

// ---- merge: EntityType selector ---------------------------------------------

// TestIntegration_Merge_EntityType_AddsDescriptionAnnotation merges a
// Core.Description annotation onto the Books entity and compares to the golden file.
func TestIntegration_Merge_EntityType_AddsDescriptionAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "merge_entitytype_expected.xml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "CatalogService.Books"},
			map[string]any{"@Core.Description": "Catalog of available books"},
		)),
	)
}

// ---- merge: EntityType + PropertyType selector ------------------------------

// TestIntegration_Merge_EntityTypeWithProperty_AddsAnnotationOnProperty merges a
// Core.Description annotation onto the Books.title property.
func TestIntegration_Merge_EntityTypeWithProperty_AddsAnnotationOnProperty(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "merge_entitytype_property_expected.xml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
			map[string]any{"@Core.Description": "Title of the book"},
		)),
	)
}

// ---- merge: EntitySet selector ----------------------------------------------

// TestIntegration_Merge_EntitySet_AddsCapabilitiesAnnotation merges a
// Capabilities.SearchRestrictions annotation onto the Books entity set.
func TestIntegration_Merge_EntitySet_AddsCapabilitiesAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "merge_entityset_expected.xml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntitySet: "Books"},
			map[string]any{"@Capabilities.SearchRestrictions": map[string]any{"Searchable": true}},
		)),
	)
}

// ---- merge: Operation selector ----------------------------------------------

// TestIntegration_Merge_Operation_AddsAnnotationOnFunction merges a Core.LongDescription
// annotation onto the getBooks function (Core.Description already exists on this target
// in the fixture, so a distinct term is used to avoid an invalid duplicate).
func TestIntegration_Merge_Operation_AddsAnnotationOnFunction(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "merge_operation_expected.xml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Operation: "CatalogService.getBooks"},
			map[string]any{"@Core.LongDescription": "Returns a collection of all available books"},
		)),
	)
}

// ---- merge: Namespace selector ----------------------------------------------

// TestIntegration_Merge_Namespace_AddsSchemaAnnotation merges a Core.Description
// annotation onto the CatalogService schema namespace.
func TestIntegration_Merge_Namespace_AddsSchemaAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "merge_namespace_expected.xml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Namespace: "CatalogService"},
			map[string]any{"@Core.Description": "CatalogService namespace annotation"},
		)),
	)
}

// TestIntegration_Merge_QualifiedAnnotations_PreservesBothQualifiers verifies
// that annotations with the same term and different qualifiers remain distinct.
func TestIntegration_Merge_QualifiedAnnotations_PreservesBothQualifiers(t *testing.T) {
	input := testutils.LoadFixture("testdata/qualified_annotations.xml")
	p, err := NewOverlayProcessor(model.ResourceDefinition{Content: input})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}

	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "CatalogService.Books"},
				Data:     map[string]any{"@Core.Description#Q1": "First description"},
			},
			{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "CatalogService.Books"},
				Data:     map[string]any{"@Core.Description#Q2": "Second description"},
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "merge_qualified_annotations_expected.xml"),
		normalizeXML(t, result.Content),
	)
}

// ---- update: EntityType selector --------------------------------------------

// TestIntegration_Update_EntityType_ReplacesAnnotationsBlock verifies that
// updating an entity replaces prior annotations with only the new ones.
func TestIntegration_Update_EntityType_ReplacesAnnotationsBlock(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "update_entitytype_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "CatalogService.Books"},
					Data:     map[string]any{"@Core.LongDescription": "old description"},
				},
				{
					Action:   "update",
					Selector: &model.Selector{EntityType: "CatalogService.Books"},
					Data:     map[string]any{"@Core.Description": "replaced description"},
				},
			}},
		}),
	)
}

// ---- remove: EntityType selector --------------------------------------------

// TestIntegration_Remove_EntityType_PrunesAnnotationTerm verifies that removing
// a specific annotation term from Books leaves other terms untouched.
func TestIntegration_Remove_EntityType_PrunesAnnotationTerm(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "remove_entitytype_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "CatalogService.Books"},
					Data: map[string]any{
						"@Core.LongDescription": "keep me not",
						"@Core.Label":           "keep me",
					},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{EntityType: "CatalogService.Books"},
					Data:     map[string]any{"@Core.LongDescription": nil},
				},
			}},
		}),
	)
}

// ---- update: Namespace selector ---------------------------------------------

// TestIntegration_Update_Namespace_ReplacesSchemaAnnotation updates the CatalogService
// schema namespace annotation, replacing any prior annotation with only the new one.
func TestIntegration_Update_Namespace_ReplacesSchemaAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "update_namespace_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{Namespace: "CatalogService"},
					Data:     map[string]any{"@Core.LongDescription": "old namespace description"},
				},
				{
					Action:   "update",
					Selector: &model.Selector{Namespace: "CatalogService"},
					Data:     map[string]any{"@Core.Description": "replaced namespace description"},
				},
			}},
		}),
	)
}

// ---- remove: Namespace selector ---------------------------------------------

// TestIntegration_Remove_Namespace_PrunesSchemaAnnotation removes a specific annotation
// term from the CatalogService namespace, leaving other terms untouched.
func TestIntegration_Remove_Namespace_PrunesSchemaAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "remove_namespace_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{Namespace: "CatalogService"},
					Data: map[string]any{
						"@Core.LongDescription": "keep me not",
						"@Core.Description":     "keep me",
					},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{Namespace: "CatalogService"},
					Data:     map[string]any{"@Core.LongDescription": nil},
				},
			}},
		}),
	)
}

// ---- update: EntityType + PropertyType selector -----------------------------

// TestIntegration_Update_EntityTypeWithProperty_ReplacesPropertyAnnotation updates
// the Books.title property annotation, replacing any prior annotation with only the new one.
func TestIntegration_Update_EntityTypeWithProperty_ReplacesPropertyAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "update_entitytype_property_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
					Data:     map[string]any{"@Core.LongDescription": "old title description"},
				},
				{
					Action:   "update",
					Selector: &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
					Data:     map[string]any{"@Core.Description": "The title of the book"},
				},
			}},
		}),
	)
}

// ---- remove: EntityType + PropertyType selector -----------------------------

// TestIntegration_Remove_EntityTypeWithProperty_PrunesPropertyAnnotation removes a
// specific annotation term from the Books.title property, leaving other terms untouched.
func TestIntegration_Remove_EntityTypeWithProperty_PrunesPropertyAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "remove_entitytype_property_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
					Data: map[string]any{
						"@Core.LongDescription": "keep me not",
						"@Core.Description":     "keep me",
					},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
					Data:     map[string]any{"@Core.LongDescription": nil},
				},
			}},
		}),
	)
}

// ---- update: EntitySet selector ---------------------------------------------

// TestIntegration_Update_EntitySet_ReplacesCapabilitiesAnnotation updates the Books
// entity set annotation, replacing prior annotations with only the new one.
func TestIntegration_Update_EntitySet_ReplacesCapabilitiesAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "update_entityset_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntitySet: "Books"},
					Data:     map[string]any{"@Capabilities.SearchRestrictions": map[string]any{"Searchable": false}},
				},
				{
					Action:   "update",
					Selector: &model.Selector{EntitySet: "Books"},
					Data:     map[string]any{"@Capabilities.SearchRestrictions": map[string]any{"Searchable": true}},
				},
			}},
		}),
	)
}

// ---- remove: EntitySet selector ---------------------------------------------

// TestIntegration_Remove_EntitySet_PrunesCapabilitiesAnnotation removes a specific
// annotation term from the Books entity set, leaving other terms untouched.
func TestIntegration_Remove_EntitySet_PrunesCapabilitiesAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "remove_entityset_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntitySet: "Books"},
					Data: map[string]any{
						"@Capabilities.SearchRestrictions": map[string]any{"Searchable": true},
						"@Core.Description":                "keep me",
					},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{EntitySet: "Books"},
					Data:     map[string]any{"@Capabilities.SearchRestrictions": nil},
				},
			}},
		}),
	)
}

// ---- update: Operation selector ---------------------------------------------

// TestIntegration_Update_Operation_ReplacesAnnotationOnFunction updates the getBooks
// function annotation, replacing prior annotations with only the new one.
func TestIntegration_Update_Operation_ReplacesAnnotationOnFunction(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "update_operation_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{Operation: "CatalogService.getBooks"},
					Data:     map[string]any{"@Core.LongDescription": "old operation description"},
				},
				{
					Action:   "update",
					Selector: &model.Selector{Operation: "CatalogService.getBooks"},
					Data:     map[string]any{"@Core.LongDescription": "Returns all books in the catalog"},
				},
			}},
		}),
	)
}

// ---- multi-patch: realistic overlay sequence --------------------------------

// TestIntegration_MultiPatch_RealisticOverlaySequence applies a sequence of patches
// that mirrors a real-world EDMX overlay:
//  1. merge EntityType "CatalogService.Books" → adds Core.Description
//  2. merge EntityType+PropertyType "CatalogService.Books/title" → adds Core.Description on property
//  3. merge EntitySet "Books" → adds Capabilities.SearchRestrictions annotation
func TestIntegration_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	testutils.AssertDeepEquals(t,
		loadIntegrationExpected(t, "multi_patch_expected.xml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "CatalogService.Books"},
					Data:     map[string]any{"@Core.Description": "Catalog of available books"},
				},
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
					Data:     map[string]any{"@Core.Description": "Title of the book"},
				},
				{
					Action:   "merge",
					Selector: &model.Selector{EntitySet: "Books"},
					Data:     map[string]any{"@Capabilities.SearchRestrictions": map[string]any{"Searchable": true}},
				},
			}},
		}),
	)
}
