//go:build unit

package edmx

import (
	"encoding/json"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// ---- nil data case ----------------------------------------------------------

func TestDecompose_NilData_ReturnsPatchUnchanged(t *testing.T) {
	patch := model.Patch{
		Action:   "remove",
		Selector: &model.Selector{EntityType: "CatalogService.Books"},
		Data:     nil,
	}
	testutils.AssertContainsInAnyOrder(t, PatchDecomposer(0).Decompose(patch), []model.Patch{patch})
}

// ---- empty / ignored-key cases ----------------------------------------------

func TestDecompose_EmptyData_ReturnsEmptySlice(t *testing.T) {
	testutils.AssertEmpty(t, PatchDecomposer(0).Decompose(model.Patch{
		Action:   "merge",
		Selector: &model.Selector{EntityType: "CatalogService.Books"},
		Data:     map[string]any{},
	}))
}

func TestDecompose_OnlyDollarPrefixedKeys_ReturnsEmptySlice(t *testing.T) {
	testutils.AssertEmpty(t, PatchDecomposer(0).Decompose(model.Patch{
		Action:   "merge",
		Selector: &model.Selector{EntityType: "CatalogService.Books"},
		Data: map[string]any{
			"$Kind":    "EntityType",
			"$Version": "4.0",
		},
	}))
}

// ---- annotation key cases ---------------------------------------------------

func TestDecompose_SingleAnnotation_NoOperation_ProducesOnePatch(t *testing.T) {
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data:        map[string]any{"@Core.Description": "A book"},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books"},
				Data:        map[string]any{"@Core.Description": "A book"},
			},
		},
	)
}

func TestDecompose_MultipleAnnotations_CollapsedIntoOnePatch(t *testing.T) {
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data: map[string]any{
				"@Core.Description":     "A book",
				"@Core.LongDescription": "A longer description",
			},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books"},
				Data: map[string]any{
					"@Core.Description":     "A book",
					"@Core.LongDescription": "A longer description",
				},
			},
		},
	)
}

// ---- property key cases -----------------------------------------------------

func TestDecompose_SingleProperty_NoOperation_SetsPropertyType(t *testing.T) {
	// Without Operation set, the property name goes into Selector.PropertyType.
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data:        map[string]any{"title": map[string]any{"@Core.Description": "The title"}},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
				Data:        map[string]any{"@Core.Description": "The title"},
			},
		},
	)
}

func TestDecompose_SingleProperty_OperationSet_SetsParameter(t *testing.T) {
	// With Operation set, the property name goes into Selector.Parameter instead.
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{Operation: "CatalogService.getBookById(Edm.Int32)"},
			Data:        map[string]any{"id": map[string]any{"@Core.OptionalParameter": map[string]any{}}},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{Operation: "CatalogService.getBookById(Edm.Int32)", Parameter: "id"},
				Data:        map[string]any{"@Core.OptionalParameter": map[string]any{}},
			},
		},
	)
}

func TestDecompose_MultipleProperties_OnePatchPerProperty(t *testing.T) {
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data: map[string]any{
				"title":    map[string]any{"@Core.Description": "Title"},
				"stock":    map[string]any{"@Core.Description": "Stock"},
				"priority": map[string]any{"@Core.Description": "Priority"},
			},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
				Data:        map[string]any{"@Core.Description": "Title"},
			},
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "stock"},
				Data:        map[string]any{"@Core.Description": "Stock"},
			},
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "priority"},
				Data:        map[string]any{"@Core.Description": "Priority"},
			},
		},
	)
}

// ---- mixed key cases --------------------------------------------------------

func TestDecompose_Mixed_AnnotationsAndProperties_ProducesCorrectPatches(t *testing.T) {
	// 1 annotation patch (all @-keys collapsed) + 1 property patch per property key.
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data: map[string]any{
				"@Core.Description":     "A book",
				"@Core.LongDescription": "Long description",
				"title":                 map[string]any{"@Core.Description": "The title"},
				"stock":                 map[string]any{"@Core.Description": "Stock level"},
			},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books"},
				Data: map[string]any{
					"@Core.Description":     "A book",
					"@Core.LongDescription": "Long description",
				},
			},
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
				Data:        map[string]any{"@Core.Description": "The title"},
			},
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "stock"},
				Data:        map[string]any{"@Core.Description": "Stock level"},
			},
		},
	)
}

func TestDecompose_DollarKeys_IgnoredAlongsideAnnotationsAndProperties(t *testing.T) {
	// $-prefixed keys are silently ignored; only the @ and property keys produce patches.
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data: map[string]any{
				"$Kind":             "EntityType",
				"@Core.Description": "A book",
				"title":             map[string]any{"@Core.Description": "The title"},
			},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books"},
				Data:        map[string]any{"@Core.Description": "A book"},
			},
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
				Data:        map[string]any{"@Core.Description": "The title"},
			},
		},
	)
}

// ---- field preservation -----------------------------------------------------

func TestDecompose_PatchFields_PreservedInAllProducedPatches(t *testing.T) {
	// Action, Description, Tags, and Meta must be copied verbatim into every
	// produced patch; only Selector and Data are rewritten.
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "update",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books"},
			Data: map[string]any{
				"@Core.Description": "A book",
				"title":             map[string]any{"@Core.Description": "The title"},
			},
		}),
		[]model.Patch{
			{
				Action:      "update",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books"},
				Data:        map[string]any{"@Core.Description": "A book"},
			},
			{
				Action:      "update",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
				Data:        map[string]any{"@Core.Description": "The title"},
			},
		},
	)
}

// ---- annotation patch: selector fields unchanged ----------------------------

func TestDecompose_AnnotationPatch_SelectorFieldsUnchanged(t *testing.T) {
	// The annotation patch must carry the original selector untouched.
	testutils.AssertContainsInAnyOrder(
		t,
		PatchDecomposer(0).Decompose(model.Patch{
			Action:      "merge",
			Description: "test-description",
			Tags:        []string{"tag-a", "tag-b"},
			Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
			Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
			Data:        map[string]any{"@Core.Description": "A book"},
		}),
		[]model.Patch{
			{
				Action:      "merge",
				Description: "test-description",
				Tags:        []string{"tag-a", "tag-b"},
				Meta:        map[string]json.RawMessage{"source": json.RawMessage(`"unit-test"`)},
				Selector:    &model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"},
				Data:        map[string]any{"@Core.Description": "A book"},
			},
		},
	)
}

// ---- property patch: Parameter cleared when no Operation --------------------

func TestDecompose_PropertyPatch_NoOperation_ParameterIsEmpty(t *testing.T) {
	// When Operation is empty, the produced property patch must have Parameter=""
	// (PropertyType gets the property name, Parameter stays empty).
	result := PatchDecomposer(0).Decompose(model.Patch{
		Action:   "merge",
		Selector: &model.Selector{EntityType: "CatalogService.Books"},
		Data:     map[string]any{"title": map[string]any{"@Core.Description": "The title"}},
	})
	if len(result) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result))
	}
	if result[0].Selector.Parameter != "" {
		t.Errorf("Parameter: got %q, want empty string", result[0].Selector.Parameter)
	}
	if result[0].Selector.PropertyType != "title" {
		t.Errorf("PropertyType: got %q, want %q", result[0].Selector.PropertyType, "title")
	}
}

// ---- property patch: PropertyType cleared when Operation is set -------------

func TestDecompose_PropertyPatch_OperationSet_PropertyTypeIsEmpty(t *testing.T) {
	// When Operation is set, the produced property patch must have PropertyType=""
	// (Parameter gets the property name, PropertyType stays empty).
	result := PatchDecomposer(0).Decompose(model.Patch{
		Action:   "merge",
		Selector: &model.Selector{Operation: "CatalogService.getBookById(Edm.Int32)"},
		Data:     map[string]any{"id": map[string]any{"@Core.OptionalParameter": map[string]any{}}},
	})
	if len(result) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result))
	}
	if result[0].Selector.PropertyType != "" {
		t.Errorf("PropertyType: got %q, want empty string", result[0].Selector.PropertyType)
	}
	if result[0].Selector.Parameter != "id" {
		t.Errorf("Parameter: got %q, want %q", result[0].Selector.Parameter, "id")
	}
}
