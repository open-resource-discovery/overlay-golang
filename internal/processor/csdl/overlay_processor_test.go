//go:build unit

package csdl

import (
	"reflect"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

var odataContent = testutils.LoadFixture("testdata/odatademo.json")

// ---- NewOverlayProcessor ----------------------------------------------------

func TestNewOverlayProcessor_ValidJSON_Succeeds(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "unexpected error: %v")

	NewOverlayProcessor(model.ResourceDefinition{Content: `{"$Version":"4.0"}`})
}

func TestNewOverlayProcessor_InvalidJSON_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid JSON")

	NewOverlayProcessor(model.ResourceDefinition{Content: `{not valid json`})
}

// ---- Apply: output fields ---------------------------------------------------

func TestApply_SetsPurposeOnResult(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	rd, err := p.Apply(model.OverlayDefinition{
		Purpose: "my-purpose",
		Overlay: model.Overlay{Patches: []model.Patch{}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rd.Purpose != "my-purpose" {
		t.Errorf("Purpose: got %q, want %q", rd.Purpose, "my-purpose")
	}
}

func TestApply_SetsVisibilityOnResult(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	rd, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Visibility: "internal", Patches: []model.Patch{}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rd.Visibility != "internal" {
		t.Errorf("Visibility: got %q, want %q", rd.Visibility, "internal")
	}
}

func TestApply_NoPatch_ContentUnchanged(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{Overlay: model.Overlay{Patches: []model.Patch{}}})
	if !reflect.DeepEqual(result["ODataDemo"], csdlDoc["ODataDemo"]) {
		t.Error("content changed without any patches")
	}
}

func TestApply_DoesNotMutateOriginalContent(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@Core.Description": "mutated"},
	))
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{Overlay: model.Overlay{Patches: []model.Patch{}}})
	if testutils.Get(t, result, "ODataDemo", "Product", "@Core.Description") != nil {
		t.Error("first Apply mutated the internal document")
	}
}

// ---- Apply: merge — semantic selectors --------------------------------------

func TestApply_Merge_EntityType_AddsNewAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@Core.Description": "A product"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Product", "@Core.Description"); got != "A product" {
		t.Errorf("got %v, want %q", got, "A product")
	}
}

func TestApply_Merge_EntityType_PreservesExistingFields(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@Core.Description": "A product"},
	))
	// $Kind must still be present
	if got := testutils.Get(t, result, "ODataDemo", "Product", "$Kind"); got != "EntityType" {
		t.Errorf("$Kind lost after merge: got %v", got)
	}
}

func TestApply_Merge_EntityType_OverwritesExistingAnnotation(t *testing.T) {
	base := `{"ODataDemo":{"Product":{"$Kind":"EntityType","@Core.Description":"old"}}}`
	p := NewOverlayProcessor(model.ResourceDefinition{Content: base, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@Core.Description": "new"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Product", "@Core.Description"); got != "new" {
		t.Errorf("got %v, want %q", got, "new")
	}
}

func TestApply_Merge_EntityTypeProperty_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Description"},
		map[string]any{"@Core.IsLanguageDependent": true},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Product", "Description", "@Core.IsLanguageDependent"); got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestApply_Merge_ComplexType_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{ComplexType: "ODataDemo.Address"},
		map[string]any{"@Core.Description": "A mailing address"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Address", "@Core.Description"); got != "A mailing address" {
		t.Errorf("got %v, want %q", got, "A mailing address")
	}
}

func TestApply_Merge_ComplexTypeProperty_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{ComplexType: "ODataDemo.Address", PropertyType: "Street"},
		map[string]any{"@Core.Description": "Street name"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Address", "Street", "@Core.Description"); got != "Street name" {
		t.Errorf("got %v, want %q", got, "Street name")
	}
}

func TestApply_Merge_EnumType_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EnumType: "ODataDemo.FileAccess"},
		map[string]any{"@Core.Description": "File access flags"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "FileAccess", "@Core.Description"); got != "File access flags" {
		t.Errorf("got %v, want %q", got, "File access flags")
	}
}

func TestApply_Merge_EnumTypeMember_AddsAnnotationWithPrefix(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EnumType: "ODataDemo.FileAccess", PropertyType: "Read"},
		map[string]any{"@Core.Description": "Read access"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "FileAccess", "Read@Core.Description"); got != "Read access" {
		t.Errorf("got %v, want %q", got, "Read access")
	}
}

func TestApply_Merge_EntitySet_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntitySet: "ODataDemo.DemoService.Products"},
		map[string]any{"@Core.Description": "All products"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "DemoService", "Products", "@Core.Description"); got != "All products" {
		t.Errorf("got %v, want %q", got, "All products")
	}
}

func TestApply_Merge_Namespace_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Namespace: "ODataDemo"},
		map[string]any{"@Core.Description": "Demo namespace"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "@Core.Description"); got != "Demo namespace" {
		t.Errorf("got %v, want %q", got, "Demo namespace")
	}
}

func TestApply_Merge_Operation_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "ODataDemo.ProductsByRating"},
		map[string]any{"@Core.Description": "Filter products by rating"},
	))
	ops := testutils.Get(t, result, "ODataDemo", "ProductsByRating").([]any)
	if got := ops[0].(map[string]any)["@Core.Description"]; got != "Filter products by rating" {
		t.Errorf("got %v, want %q", got, "Filter products by rating")
	}
}

func TestApply_Merge_OperationParameter_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "ODataDemo.ProductsByRating", Parameter: "Rating"},
		map[string]any{"@Core.Description": "Minimum rating"},
	))
	ops := testutils.Get(t, result, "ODataDemo", "ProductsByRating").([]any)
	params := ops[0].(map[string]any)["$Parameter"].([]any)
	if got := params[0].(map[string]any)["@Core.Description"]; got != "Minimum rating" {
		t.Errorf("got %v, want %q", got, "Minimum rating")
	}
}

func TestApply_Merge_OperationReturnType_AddsAnnotation(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Operation: "ODataDemo.ProductsByRating", ReturnType: utils.Ptr(true)},
		map[string]any{"@Core.Description": "The returned products"},
	))
	ops := testutils.Get(t, result, "ODataDemo", "ProductsByRating").([]any)
	if got := ops[0].(map[string]any)["$ReturnType"].(map[string]any)["@Core.Description"]; got != "The returned products" {
		t.Errorf("got %v, want %q", got, "The returned products")
	}
}

// ---- Apply: merge — inline property decomposition ---------------------------

func TestApply_Merge_EntityType_InlineProperty_AddsAnnotationOnProperty(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{
			"Rating": map[string]any{"@Core.Description": "Star rating"},
		},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Product", "Rating", "@Core.Description"); got != "Star rating" {
		t.Errorf("got %v, want %q", got, "Star rating")
	}
}

func TestApply_Merge_EntityType_MixedAnnotationAndProperty(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{
			"@Core.Description": "A product",
			"Rating":            map[string]any{"@Core.Description": "Star rating"},
		},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Product", "@Core.Description"); got != "A product" {
		t.Errorf("entity annotation: got %v, want %q", got, "A product")
	}
	if got := testutils.Get(t, result, "ODataDemo", "Product", "Rating", "@Core.Description"); got != "Star rating" {
		t.Errorf("property annotation: got %v, want %q", got, "Star rating")
	}
}

func TestApply_Merge_DollarPrefixedKeys_AreIgnored(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{
			"$Kind":             "SomethingElse",
			"@Core.Description": "A product",
		},
	))
	// $Kind must not be changed
	if got := testutils.Get(t, result, "ODataDemo", "Product", "$Kind"); got != "EntityType" {
		t.Errorf("$Kind was mutated: got %v", got)
	}
	if got := testutils.Get(t, result, "ODataDemo", "Product", "@Core.Description"); got != "A product" {
		t.Errorf("annotation missing: got %v", got)
	}
}

// ---- Apply: merge — JSONPath selector ---------------------------------------

func TestApply_Merge_JSONPath_AddsKey(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{JSONPath: "$.ODataDemo.Product"},
		map[string]any{"@Core.Description": "via jsonpath"},
	))
	if got := testutils.Get(t, result, "ODataDemo", "Product", "@Core.Description"); got != "via jsonpath" {
		t.Errorf("got %v, want %q", got, "via jsonpath")
	}
}

func TestApply_Merge_JSONPath_NonExistentPath_CreatesNode(t *testing.T) {
	// JSONPath selectors may now point to non-existing parts of the document.
	// A merge via a non-existent path creates the node rather than being a no-op.
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "merge",
				Selector: &model.Selector{JSONPath: "$.ODataDemo.NonExistent"},
				Data:     map[string]any{"@x": "created"},
			},
		}},
	})
	node := testutils.Get(t, result, "ODataDemo", "NonExistent")
	if node == nil {
		t.Fatal("expected node to be created for non-existent JSONPath, but key is absent")
	}
	m, ok := node.(map[string]any)
	if !ok {
		t.Fatalf("ODataDemo.NonExistent: expected map, got %T", node)
	}
	if m["@x"] != "created" {
		t.Errorf("ODataDemo.NonExistent['@x']: got %v, want %q", m["@x"], "created")
	}
}

func TestApply_Update_JSONPath_NonExistentPath_CreatesNode(t *testing.T) {
	// An update via a non-existent JSONPath creates the node rather than being a no-op.
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{
				Action:   "update",
				Selector: &model.Selector{JSONPath: "$.ODataDemo.NewNode"},
				Data:     map[string]any{"@x": "value"},
			},
		}},
	})
	node := testutils.Get(t, result, "ODataDemo", "NewNode")
	if node == nil {
		t.Fatal("expected node to be created for non-existent JSONPath, but key is absent")
	}
	m, ok := node.(map[string]any)
	if !ok {
		t.Fatalf("ODataDemo.NewNode: expected map, got %T", node)
	}
	if m["@x"] != "value" {
		t.Errorf("ODataDemo.NewNode['@x']: got %v, want %q", m["@x"], "value")
	}
}

// ---- Apply: merge — root selector -------------------------------------------

func TestApply_Merge_Root_MergesIntoDocument(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: `{"$Version":"4.0","ODataDemo":{}}`, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("merge",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"@Core.SchemaVersion": "1.0"},
	))
	if result["$Version"] != "4.0" {
		t.Errorf("$Version lost: got %v", result["$Version"])
	}
}

// ---- Apply: update action ---------------------------------------------------

func TestApply_Update_EntityType_ReplacesNode(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@Core.Description": "replaced"},
	))
	got := testutils.Get(t, result, "ODataDemo", "Product").(map[string]any)
	if got["@Core.Description"] != "replaced" {
		t.Errorf("@Core.Description: got %v, want %q", got["@Core.Description"], "replaced")
	}
}

func TestApply_Update_EntityType_RemovesExistingFields(t *testing.T) {
	// update with a semantic selector performs full replacement: existing fields
	// that are not present in the patch data must be absent from the result.
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"$Kind": "EntityType", "$Key": []any{"ID"}, "ID": map[string]any{}, "@Core.Description": "Replaced."},
	))
	got := testutils.Get(t, result, "ODataDemo", "Product").(map[string]any)
	if _, ok := got["$HasStream"]; ok {
		t.Error("expected $HasStream removed by update, but it still exists")
	}
	if _, ok := got["Price"]; ok {
		t.Error("expected Price removed by update, but it still exists")
	}
	if got["@Core.Description"] != "Replaced." {
		t.Errorf("@Core.Description: got %v, want %q", got["@Core.Description"], "Replaced.")
	}
}

func TestApply_Update_EntityTypeProperty_ReplacesPropertyEntirely(t *testing.T) {
	// update on EntityType+PropertyType replaces the property value entirely;
	// attributes not supplied in the patch (e.g. @Measures.ISOCurrency) must be gone.
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Price"},
		map[string]any{"$Type": "Edm.Decimal", "$Nullable": false, "@Core.Description": "Mandatory price."},
	))
	price := testutils.Get(t, result, "ODataDemo", "Product", "Price").(map[string]any)
	if price["$Nullable"] != false {
		t.Errorf("$Nullable: got %v, want false", price["$Nullable"])
	}
	if price["@Core.Description"] != "Mandatory price." {
		t.Errorf("@Core.Description: got %v, want %q", price["@Core.Description"], "Mandatory price.")
	}
	if _, ok := price["@Measures.ISOCurrency"]; ok {
		t.Error("expected @Measures.ISOCurrency removed by update, but it still exists")
	}
}

func TestApply_Update_EnumTypeMember_ReplacesExistingAnnotation(t *testing.T) {
	// update on EnumType+PropertyType removes all existing MemberName@* annotation
	// keys for that member and sets the new annotation supplied in the patch.
	base := `{"ODataDemo":{"FileAccess":{"$Kind":"EnumType","Read":1,"Read@Core.Description":"old","Read@Core.LongDescription":"also old"}}}`
	p := NewOverlayProcessor(model.ResourceDefinition{Content: base, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EnumType: "ODataDemo.FileAccess", PropertyType: "Read"},
		map[string]any{"@Core.Description": "new description"},
	))
	fa := testutils.Get(t, result, "ODataDemo", "FileAccess").(map[string]any)
	if fa["Read@Core.Description"] != "new description" {
		t.Errorf("Read@Core.Description: got %v, want %q", fa["Read@Core.Description"], "new description")
	}
	if _, ok := fa["Read@Core.LongDescription"]; ok {
		t.Error("expected Read@Core.LongDescription removed by update, but it still exists")
	}
}

func TestApply_Update_EnumTypeMember_LeavesOtherMembersUntouched(t *testing.T) {
	// update on one enum member must not affect annotations of sibling members.
	base := `{"ODataDemo":{"FileAccess":{"$Kind":"EnumType","Read":1,"Read@Core.Description":"old","Write":2,"Write@Core.Description":"write desc"}}}`
	p := NewOverlayProcessor(model.ResourceDefinition{Content: base, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EnumType: "ODataDemo.FileAccess", PropertyType: "Read"},
		map[string]any{"@Core.Description": "updated read"},
	))
	fa := testutils.Get(t, result, "ODataDemo", "FileAccess").(map[string]any)
	if fa["Read@Core.Description"] != "updated read" {
		t.Errorf("Read@Core.Description: got %v, want %q", fa["Read@Core.Description"], "updated read")
	}
	if fa["Write@Core.Description"] != "write desc" {
		t.Errorf("Write@Core.Description must be preserved: got %v", fa["Write@Core.Description"])
	}
}

func TestApply_Remove_EnumTypeMember_RemovesMemberAndItsAnnotations(t *testing.T) {
	base := "{\"ODataDemo\":{\"FileAccess\":{\"$Kind\":\"EnumType\",\"Read\":1,\"Read@Core.Description\":\"read\",\"Write\":2,\"Write@Core.Description\":\"write\"}}}"
	p := NewOverlayProcessor(model.ResourceDefinition{Content: base, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{EnumType: "ODataDemo.FileAccess", PropertyType: "Read"},
		nil,
	))
	fileAccess := testutils.Get(t, result, "ODataDemo", "FileAccess").(map[string]any)
	if _, ok := fileAccess["Read"]; ok {
		t.Error("expected Read member removed")
	}
	if _, ok := fileAccess["Read@Core.Description"]; ok {
		t.Error("expected Read annotation removed")
	}
	if _, ok := fileAccess["Write"]; !ok || fileAccess["Write@Core.Description"] != "write" {
		t.Errorf("expected sibling member preserved, got %v", fileAccess)
	}
}

func TestApply_Update_EnumTypeAndMember_SimultaneouslyAnnotatesBoth(t *testing.T) {
	// A single update patch whose selector targets an enum type and whose data
	// contains both a top-level annotation (applied to the type) and a nested
	// map keyed by a member name (decomposed into a member-level update).
	base := `{"ODataDemo":{"FileAccess":{"$Kind":"EnumType","Read":1,"Read@Core.Description":"old read","Write":2,"Write@Core.Description":"old write"}}}`
	p := NewOverlayProcessor(model.ResourceDefinition{Content: base, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{EnumType: "ODataDemo.FileAccess"},
		map[string]any{
			"@Core.Description": "updated enum description",
			"Read":              map[string]any{"@Core.Description": "updated read description"},
		},
	))
	fa := testutils.Get(t, result, "ODataDemo", "FileAccess").(map[string]any)
	if fa["@Core.Description"] != "updated enum description" {
		t.Errorf("@Core.Description on enum type: got %v, want %q", fa["@Core.Description"], "updated enum description")
	}
	if fa["Read@Core.Description"] != "updated read description" {
		t.Errorf("Read@Core.Description: got %v, want %q", fa["Read@Core.Description"], "updated read description")
	}
	// sibling member annotation must survive
	if fa["Write@Core.Description"] != "old write" {
		t.Errorf("Write@Core.Description must be preserved: got %v", fa["Write@Core.Description"])
	}
}

func TestApply_Update_Root_ReplacesDocument(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: `{"old":"value"}`, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{Root: utils.Ptr(true)},
		map[string]any{"@new": "value"},
	))
	if result["@new"] != "value" {
		t.Errorf("expected @new key, got %v", result)
	}
	if _, exists := result["old"]; exists {
		t.Error("expected old key removed after root update")
	}
}

func TestApply_Update_JSONPath_ReplacesNode(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("update",
		model.Selector{JSONPath: "$.ODataDemo.Product"},
		map[string]any{"@Core.Description": "jsonpath-replaced"},
	))
	got := testutils.Get(t, result, "ODataDemo", "Product").(map[string]any)
	if got["@Core.Description"] != "jsonpath-replaced" {
		t.Errorf("got %v, want %q", got["@Core.Description"], "jsonpath-replaced")
	}
	// structural fields are gone — update replaces entirely
	if _, ok := got["$Kind"]; ok {
		t.Error("expected $Kind removed by update, but it still exists")
	}
}

// ---- Apply: remove action ---------------------------------------------------

func TestApply_Remove_EntityType_NilData_DeletesEntireNode(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "remove", Selector: &model.Selector{EntityType: "ODataDemo.Category"}, Data: nil},
		}},
	})
	ns := testutils.Get(t, result, "ODataDemo").(map[string]any)
	if _, exists := ns["Category"]; exists {
		t.Error("expected Category deleted, but it still exists")
	}
}

func TestApply_Remove_EntityTypeProperty_NilData_DeletesProperty(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "remove", Selector: &model.Selector{EntityType: "ODataDemo.Product", PropertyType: "Rating"}, Data: nil},
		}},
	})
	product := testutils.Get(t, result, "ODataDemo", "Product").(map[string]any)
	if _, exists := product["Rating"]; exists {
		t.Error("expected Rating deleted, but it still exists")
	}
}

func TestApply_Remove_AnnotationKey_NullValue_DeletesAnnotation(t *testing.T) {
	base := `{"ODataDemo":{"Product":{"$Kind":"EntityType","@Core.Description":"v","@Core.Tag":"t"}}}`
	p := NewOverlayProcessor(model.ResourceDefinition{Content: base, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, testutils.OnePatch("remove",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@Core.Description": nil},
	))
	product := testutils.Get(t, result, "ODataDemo", "Product").(map[string]any)
	if _, exists := product["@Core.Description"]; exists {
		t.Error("expected @Core.Description deleted, but it still exists")
	}
	// other annotation must survive
	if product["@Core.Tag"] != "t" {
		t.Errorf("expected @Core.Tag preserved, got %v", product["@Core.Tag"])
	}
}

func TestApply_Remove_JSONPath_NilData_DeletesNode(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "remove", Selector: &model.Selector{JSONPath: "$.ODataDemo.Category"}, Data: nil},
		}},
	})
	ns := testutils.Get(t, result, "ODataDemo").(map[string]any)
	if _, exists := ns["Category"]; exists {
		t.Error("expected Category deleted via JSONPath, but it still exists")
	}
}

func TestApply_Remove_JSONPath_NonExistentPath_IsNoOp(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	// Should not error; expression.Has returns false → skipped
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "remove", Selector: &model.Selector{JSONPath: "$.ODataDemo.NonExistent"}, Data: nil},
		}},
	})
	if testutils.Get(t, result, "ODataDemo") == nil {
		t.Error("ODataDemo unexpectedly removed")
	}
}

func TestApply_Remove_Root_NilData_ReturnsError(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: `{"key":"value"}`, MediaType: "application/json"})
	_, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "remove", Selector: &model.Selector{Root: utils.Ptr(true)}, Data: nil},
		}},
	})
	if err == nil {
		t.Fatal("expected root remove to return an error")
	}
}

// ---- Apply: patch ordering --------------------------------------------------

func TestApply_PatchesAppliedInOrder_LastWins(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{EntityType: "ODataDemo.Product"}, Data: map[string]any{"@x": "a"}},
			{Action: "merge", Selector: &model.Selector{EntityType: "ODataDemo.Product"}, Data: map[string]any{"@x": "b"}},
			{Action: "merge", Selector: &model.Selector{EntityType: "ODataDemo.Product"}, Data: map[string]any{"@x": "c"}},
		}},
	})
	if got := testutils.Get(t, result, "ODataDemo", "Product", "@x"); got != "c" {
		t.Errorf("got %v, want %q", got, "c")
	}
}

func TestApply_MergeFollowedByRemove_KeyRemoved(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	result := testutils.ApplyAndParse(t, p, model.OverlayDefinition{
		Overlay: model.Overlay{Patches: []model.Patch{
			{Action: "merge", Selector: &model.Selector{EntityType: "ODataDemo.Product"}, Data: map[string]any{"@Core.Description": "added"}},
			{Action: "remove", Selector: &model.Selector{EntityType: "ODataDemo.Product"}, Data: map[string]any{"@Core.Description": nil}},
		}},
	})
	product := testutils.Get(t, result, "ODataDemo", "Product").(map[string]any)
	if _, exists := product["@Core.Description"]; exists {
		t.Error("expected @Core.Description removed after merge+remove sequence")
	}
}

// ---- Apply: error cases -----------------------------------------------------

func TestApply_UnknownAction_ReturnsError(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	_, err := p.Apply(testutils.OnePatch("upsert",
		model.Selector{EntityType: "ODataDemo.Product"},
		map[string]any{"@x": "y"},
	))
	if err == nil {
		t.Fatal("expected error for unknown action, got nil")
	}
}

func TestApply_SemanticSelector_NotFound_ReturnsError(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{EntityType: "ODataDemo.NonExistent"},
		map[string]any{"@x": "y"},
	))
	if err == nil {
		t.Fatal("expected error for non-existent entity type, got nil")
	}
}

func TestApply_UnsupportedSelector_ReturnsError(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{},
		map[string]any{"@x": "y"},
	))
	if err == nil {
		t.Fatal("expected error for empty selector, got nil")
	}
}

func TestApply_InvalidJSONPath_ReturnsError(t *testing.T) {
	p := NewOverlayProcessor(model.ResourceDefinition{Content: odataContent, MediaType: "application/json"})
	_, err := p.Apply(testutils.OnePatch("merge",
		model.Selector{JSONPath: "$$[invalid"},
		map[string]any{"@x": "y"},
	))
	if err == nil {
		t.Fatal("expected error for invalid JSONPath, got nil")
	}
}
