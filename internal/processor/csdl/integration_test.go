//go:build integration

package csdl

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// integrationInput is the shared CSDL fixture for all integration tests.
var integrationInput = testutils.LoadFixture("testdata/odatademo.json")

// applyIntegration applies the given overlay to the integration input fixture
// and returns the parsed result document.
func applyIntegration(t *testing.T, od model.OverlayDefinition) map[string]any {
	t.Helper()

	definition := model.ResourceDefinition{
		Content:   integrationInput,
		MediaType: "application/json",
	}

	return testutils.ApplyAndParse(t, testutils.AssertNoError(NewOverlayProcessor(definition)), od)
}

// loadExpected loads and parses an expected-output fixture from the integration testdata directory.
func loadExpected(path string) map[string]any {
	return testutils.UnmarshalFixture[map[string]any]("testdata/integration/" + path)
}

// ---- merge: root selector ---------------------------------------------------

// TestIntegration_Merge_Root_AddsSchemaAnnotations merges two Core annotations at the
// root of the document, verifying that all existing top-level keys are preserved while
// the new annotations are added.
func TestIntegration_Merge_Root_AddsSchemaAnnotations(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"@Core.SchemaVersion": "1.0.0",
				"@Core.Description":   "OData Demo Service",
			},
		)),
	)
}

// ---- update: root selector --------------------------------------------------

// TestIntegration_Update_Root_ReplacesEntireDocument replaces the whole document with a
// minimal stub containing only $Version and $EntityContainer, verifying that the
// ODataDemo namespace and all its contents are gone.
func TestIntegration_Update_Root_ReplacesEntireDocument(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_root_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"$Version":         "4.01",
				"$EntityContainer": "ODataDemo.DemoService",
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

// ---- merge: JSONPath selector -----------------------------------------------

// TestIntegration_Merge_JSONPath_AddsDescriptionToProduct merges a @Core.Description
// annotation onto the Product entity via JSONPath, verifying that all existing
// entity properties are preserved.
func TestIntegration_Merge_JSONPath_AddsDescriptionToProduct(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_jsonpath_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{JSONPath: "$.ODataDemo.Product"},
			map[string]any{"@Core.Description": "Represents a product for sale."},
		)),
	)
}

// ---- update: JSONPath selector ----------------------------------------------

// TestIntegration_Update_JSONPath_ReplacesProductPriceProperty replaces the entire
// Price property on Product with an enriched definition that adds a @Core.Description
// annotation and changes $Nullable, verifying that the prior property is fully replaced.
func TestIntegration_Update_JSONPath_ReplacesProductPriceProperty(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_jsonpath_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{JSONPath: "$.ODataDemo.Product.Price"},
			map[string]any{
				"$Nullable":         true,
				"$Type":             "Edm.Decimal",
				"@Core.Description": "Product price including tax.",
			},
		)),
	)
}

// ---- remove: JSONPath selector ----------------------------------------------

// TestIntegration_Remove_JSONPath_DeletesProductDescriptionProperty removes the
// Description property from the Product entity via JSONPath, verifying that all
// other properties remain intact.
func TestIntegration_Remove_JSONPath_DeletesProductDescriptionProperty(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_jsonpath_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.ODataDemo.Product.Description"},
			nil,
		)),
	)
}

// ---- merge: namespace selector ----------------------------------------------

// TestIntegration_Merge_Namespace_AddsAnnotations merges two annotations onto the
// ODataDemo namespace object directly, verifying that all entity type and other
// namespace members are preserved.
func TestIntegration_Merge_Namespace_AddsAnnotations(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_namespace_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Namespace: "ODataDemo"},
			map[string]any{
				"@Core.DefaultNamespace": true,
				"@Core.SchemaVersion":    "2.0",
			},
		)),
	)
}

// ---- merge: EntityType selector ---------------------------------------------

// TestIntegration_Merge_EntityType_AddsAnnotations merges two Core annotations onto
// the Product entity type, verifying that existing properties ($Kind, $Key, ID,
// Description, Price) are all preserved.
func TestIntegration_Merge_EntityType_AddsAnnotations(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_entitytype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "ODataDemo.Product"},
			map[string]any{
				"@Core.Description":     "A product entity.",
				"@Core.LongDescription": "Detailed product data.",
			},
		)),
	)
}

// ---- update: EntityType selector --------------------------------------------

// TestIntegration_Update_EntityType_ReplacesProductEntity replaces the Product entity
// with a minimal definition, verifying that the Description and Price properties
// are absent and the new @Core.Description annotation is present.
// TestIntegration_Update_EntityType_ReplacesProductEntity replaces the Product entity
// with a minimal definition, verifying that the @Core.Description annotation is present
// and the entity has been updated to the supplied definition.
func TestIntegration_Update_EntityType_ReplacesProductEntity(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_entitytype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{EntityType: "ODataDemo.Product"},
			map[string]any{
				"$Kind":             "EntityType",
				"$Key":              []any{"ID"},
				"ID":                map[string]any{},
				"@Core.Description": "Replaced product entity.",
			},
		)),
	)
}

// ---- remove: EntityType selector --------------------------------------------

// TestIntegration_Remove_EntityType_DeletesProductEntity removes the Product entity
// from the namespace entirely, verifying that Address and FileAccess still exist.
func TestIntegration_Remove_EntityType_DeletesProductEntity(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_entitytype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{EntityType: "ODataDemo.Product"},
			nil,
		)),
	)
}

// ---- merge: EntityType + PropertyType selector ------------------------------

// TestIntegration_Merge_EntityTypeProperty_AddsDescriptionToPrice merges a
// @Core.Description onto the Price property of Product, verifying that all other
// property attributes ($Nullable, $Type) and sibling properties are preserved.
func TestIntegration_Merge_EntityTypeProperty_AddsDescriptionToPrice(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_entitytype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Price"},
			map[string]any{"@Core.Description": "Sale price in default currency."},
		)),
	)
}

// ---- update: EntityType + PropertyType selector -----------------------------

// TestIntegration_Update_EntityTypeProperty_ReplacesPrice replaces the Price property
// entirely with a mandatory (non-nullable) definition that adds @Core.Description,
// verifying that the prior definition is fully replaced.
func TestIntegration_Update_EntityTypeProperty_ReplacesPrice(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_entitytype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Price"},
			map[string]any{
				"$Nullable":         false,
				"$Type":             "Edm.Decimal",
				"@Core.Description": "Mandatory price field.",
			},
		)),
	)
}

// ---- remove: EntityType + PropertyType selector -----------------------------

// TestIntegration_Remove_EntityTypeProperty_RemovesAnnotation removes the
// @Measures.ISOCurrency annotation from the Price property via a nil-keyed data entry,
// verifying that $Nullable and $Type survive.
func TestIntegration_Remove_EntityTypeProperty_RemovesISOCurrencyAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_entitytype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Price"},
			map[string]any{"@Measures.ISOCurrency": nil},
		)),
	)
}

// ---- merge: ComplexType selector --------------------------------------------

// TestIntegration_Merge_ComplexType_AddsAnnotations merges two Core annotations onto
// the Address complex type, verifying that Street and City properties are preserved.
func TestIntegration_Merge_ComplexType_AddsAnnotations(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_complextype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{ComplexType: "ODataDemo.Address"},
			map[string]any{
				"@Core.Description":     "Postal address.",
				"@Core.LongDescription": "Full mailing address.",
			},
		)),
	)
}

// ---- remove: ComplexType selector -------------------------------------------

// TestIntegration_Remove_ComplexType_DeletesAddress removes the Address complex type
// from the namespace entirely, verifying that other types are unaffected.
func TestIntegration_Remove_ComplexType_DeletesAddress(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_complextype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{ComplexType: "ODataDemo.Address"},
			nil,
		)),
	)
}

// ---- merge: ComplexType + PropertyType selector -----------------------------

// TestIntegration_Merge_ComplexTypeProperty_AddsDescriptionToCity merges a
// @Core.Description onto the City property of Address, verifying that $Nullable
// and the Street property are preserved.
func TestIntegration_Merge_ComplexTypeProperty_AddsDescriptionToCity(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_complextype_property_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{ComplexType: "ODataDemo.Address", PropertyType: "City"},
			map[string]any{"@Core.Description": "City name."},
		)),
	)
}

// ---- merge: EnumType selector -----------------------------------------------

// TestIntegration_Merge_EnumType_AddsDescription merges a @Core.Description onto the
// FileAccess enum type, verifying that all member values (Read, Write, Create, Delete)
// and $Kind are preserved.
func TestIntegration_Merge_EnumType_AddsDescription(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_enumtype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EnumType: "ODataDemo.FileAccess"},
			map[string]any{"@Core.Description": "Bitmask for file access permissions."},
		)),
	)
}

// ---- remove: EnumType selector ----------------------------------------------

// TestIntegration_Remove_EnumType_DeletesFileAccess removes the FileAccess enum type
// from the namespace entirely, verifying that Product and Address still exist.
func TestIntegration_Remove_EnumType_DeletesFileAccess(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_enumtype_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{EnumType: "ODataDemo.FileAccess"},
			nil,
		)),
	)
}

// ---- merge: EnumType member selector ----------------------------------------

// TestIntegration_Merge_EnumTypeMember_AddsDescriptionToReadMember merges a
// @Core.Description onto the Read member of the FileAccess enum. Per the OData CSDL
// spec, member annotations are stored on the enum type itself, so the annotation
// appears at the FileAccess level with the member path as key.
func TestIntegration_Merge_EnumTypeMember_AddsDescriptionToReadMember(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_enumtype_member_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EnumType: "ODataDemo.FileAccess", PropertyType: "Read"},
			map[string]any{"@Core.Description": "Grants read permission."},
		)),
	)
}

// ---- merge: EntitySet selector ----------------------------------------------

// TestIntegration_Merge_EntitySet_AddsDescription merges a @Core.Description onto the
// Products entity set in DemoService, verifying that $Collection and $Type survive.
func TestIntegration_Merge_EntitySet_AddsDescription(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_entityset_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{EntitySet: "Products"},
			map[string]any{"@Core.Description": "Set of all available products."},
		)),
	)
}

// ---- remove: EntitySet selector ---------------------------------------------

// TestIntegration_Remove_EntitySet_RemovesDescriptionAnnotation removes the
// @Core.Description annotation from the Categories entity set via a nil-keyed data
// entry, verifying that $Collection and $Type survive.
func TestIntegration_Remove_EntitySet_RemovesDescriptionAnnotation(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("remove_entityset_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{EntitySet: "Categories"},
			map[string]any{"@Core.Description": nil},
		)),
	)
}

// ---- merge: operation selector ----------------------------------------------

// TestIntegration_Merge_Operation_AddsDescription merges a @Core.Description onto
// the ProductsByRating function, verifying that $Kind, $Parameter, and $ReturnType
// are all preserved.
func TestIntegration_Merge_Operation_AddsDescription(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("merge_operation_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Operation: "ODataDemo.ProductsByRating"},
			map[string]any{"@Core.Description": "Returns products filtered by minimum rating."},
		)),
	)
}

// ---- update: operation selector ---------------------------------------------

// TestIntegration_Update_Operation_ReplacesProductsByRating replaces the
// ProductsByRating function definition entirely, changing the Rating parameter to
// non-nullable and adding a @Core.Description, verifying full replacement semantics.
func TestIntegration_Update_Operation_ReplacesProductsByRating(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("update_operation_expected.json"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{Operation: "ODataDemo.ProductsByRating"},
			map[string]any{
				"$Kind": "Function",
				"$Parameter": []any{
					map[string]any{"$Name": "Rating", "$Nullable": false, "$Type": "Edm.Int32"},
				},
				"$ReturnType":       map[string]any{"$Collection": true, "$Type": "self.Product"},
				"@Core.Description": "Returns products with a rating at or above the given value.",
			},
		)),
	)
}

// ---- multi-patch: realistic overlay sequence --------------------------------

// TestIntegration_MultiPatch_RealisticOverlaySequence applies a sequence of four
// patches: add schema version to namespace, annotate Product entity, remove a
// property annotation, and annotate the Address complex type. Verifies that all
// changes are applied in order and independently.
func TestIntegration_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadExpected("multi_patch_expected.json"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{Namespace: "ODataDemo"},
					Data:     map[string]any{"@Core.SchemaVersion": "3.0"},
				},
				{
					Action:   "merge",
					Selector: &model.Selector{EntityType: "ODataDemo.Product"},
					Data:     map[string]any{"@Core.Description": "A product entity."},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Description"},
					Data:     map[string]any{"@Core.IsLanguageDependent": nil},
				},
				{
					Action:   "merge",
					Selector: &model.Selector{ComplexType: "ODataDemo.Address"},
					Data:     map[string]any{"@Core.Description": "Postal address."},
				},
			}},
		}),
	)
}
