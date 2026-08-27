//go:build unit

package csdl

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// csdlDoc is the parsed ODataDemo CSDL JSON fixture loaded once for all tests.
var csdlDoc = testutils.UnmarshalFixture[map[string]any]("testdata/odatademo.json")

// ---- Root -------------------------------------------------------------------

func TestExpressions_Root_ReturnsExpressionThatResolvesToDocument(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{Root: utils.Ptr(true)})
	if err != nil {
		t.Fatalf("Resolve(Root): %v", err)
	}
	testutils.AssertExpr(t, expr, "$")
	testutils.AssertResolvesToNode(t, csdlDoc, expr, csdlDoc)
}

// ---- Namespace --------------------------------------------------------------

func TestExpressions_Namespace_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).Namespace(csdlDoc, "ODataDemo"))

	testutils.AssertExpr(t, expression, "$.ODataDemo")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"])
}

func TestExpressions_Namespace_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Namespace(csdlDoc, "NonExistent"); err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
}

// ---- EntityType -------------------------------------------------------------

func TestExpressions_EntityType_FullyQualified_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).EntityType(csdlDoc, "ODataDemo.Product"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.Product")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["Product"])
}

func TestExpressions_EntityType_UnqualifiedName_Found(t *testing.T) {
	// Without a namespace prefix, a wildcard search is used; pinpoint resolves
	// the wildcard to the concrete path of the unique match.
	expression := testutils.AssertNoError(Expressions(0).EntityType(csdlDoc, "Product"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.Product")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["Product"])
}

func TestExpressions_EntityType_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).EntityType(csdlDoc, "ODataDemo.NonExistent"); err == nil {
		t.Fatal("expected error for missing entity type, got nil")
	}
}

func TestExpressions_EntityType_WrongKind_ReturnsError(t *testing.T) {
	// Address is a ComplexType, not an EntityType.
	if _, err := Expressions(0).EntityType(csdlDoc, "ODataDemo.Address"); err == nil {
		t.Fatal("expected error for wrong $Kind, got nil")
	}
}

// ---- EntityTypeProperty -----------------------------------------------------

func TestExpressions_EntityTypeProperty_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).EntityTypeProperty(csdlDoc, "ODataDemo.Product", "Description"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.Product.Description")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["Product"].(map[string]any)["Description"])
}

func TestExpressions_EntityTypeProperty_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).EntityTypeProperty(csdlDoc, "ODataDemo.Product", "NonExistent"); err == nil {
		t.Fatal("expected error for missing property, got nil")
	}
}

func TestExpressions_EntityTypeProperty_WrongParentKind_ReturnsError(t *testing.T) {
	// Address is a ComplexType, not an EntityType.
	if _, err := Expressions(0).EntityTypeProperty(csdlDoc, "ODataDemo.Address", "Street"); err == nil {
		t.Fatal("expected error when parent is not an EntityType, got nil")
	}
}

// ---- ComplexType ------------------------------------------------------------

func TestExpressions_ComplexType_FullyQualified_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).ComplexType(csdlDoc, "ODataDemo.Address"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.Address")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["Address"])
}

func TestExpressions_ComplexType_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).ComplexType(csdlDoc, "ODataDemo.NonExistent"); err == nil {
		t.Fatal("expected error for missing complex type, got nil")
	}
}

func TestExpressions_ComplexType_WrongKind_ReturnsError(t *testing.T) {
	// Product is an EntityType, not a ComplexType.
	if _, err := Expressions(0).ComplexType(csdlDoc, "ODataDemo.Product"); err == nil {
		t.Fatal("expected error for wrong $Kind, got nil")
	}
}

// ---- ComplexTypeProperty ----------------------------------------------------

func TestExpressions_ComplexTypeProperty_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).ComplexTypeProperty(csdlDoc, "ODataDemo.Address", "Street"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.Address.Street")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["Address"].(map[string]any)["Street"])
}

func TestExpressions_ComplexTypeProperty_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).ComplexTypeProperty(csdlDoc, "ODataDemo.Address", "NonExistent"); err == nil {
		t.Fatal("expected error for missing property, got nil")
	}
}

func TestExpressions_ComplexTypeProperty_WrongParentKind_ReturnsError(t *testing.T) {
	// Product is an EntityType, not a ComplexType.
	if _, err := Expressions(0).ComplexTypeProperty(csdlDoc, "ODataDemo.Product", "Street"); err == nil {
		t.Fatal("expected error when parent is not an EntityType, got nil")
	}
}

// ---- Operation (Function) ---------------------------------------------------

func TestExpressions_Operation_Function_FullyQualified_Found(t *testing.T) {
	// ProductsByRating is stored as an array per the OData CSDL JSON spec.
	expression := testutils.AssertNoError(Expressions(0).Operation(csdlDoc, "ODataDemo.ProductsByRating"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.ProductsByRating[0]")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["ProductsByRating"].([]any)[0])
}

func TestExpressions_Operation_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Operation(csdlDoc, "ODataDemo.NonExistent"); err == nil {
		t.Fatal("expected error for missing operation, got nil")
	}
}

func TestExpressions_Operation_WrongKind_ReturnsError(t *testing.T) {
	// Product is an EntityType, not an Action or Function.
	if _, err := Expressions(0).Operation(csdlDoc, "ODataDemo.Product"); err == nil {
		t.Fatal("expected error for wrong $Kind, got nil")
	}
}

// ---- OperationParameter -----------------------------------------------------

func TestExpressions_OperationParameter_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).OperationParameter(csdlDoc, "ODataDemo.ProductsByRating", "Rating"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.ProductsByRating[0]['$Parameter'][?(@['$Name'] == 'Rating')]")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["ProductsByRating"].([]any)[0].(map[string]any)["$Parameter"].([]any)[0])
}

func TestExpressions_OperationParameter_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).OperationParameter(csdlDoc, "ODataDemo.ProductsByRating", "NonExistent"); err == nil {
		t.Fatal("expected error for missing parameter, got nil")
	}
}

func TestExpressions_OperationParameter_ParentNotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).OperationParameter(csdlDoc, "ODataDemo.NonExistent", "Rating"); err == nil {
		t.Fatal("expected error when parent operation does not exist, got nil")
	}
}

// ---- EntitySet --------------------------------------------------------------

func TestExpressions_EntitySet_ShortForm_Found(t *testing.T) {
	// Short form: just the entity set name, no namespace prefix.
	expression := testutils.AssertNoError(Expressions(0).EntitySet(csdlDoc, "Products"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.DemoService.Products")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["DemoService"].(map[string]any)["Products"])
}

func TestExpressions_EntitySet_NamespaceAndEntitySet_Found(t *testing.T) {
	// Medium form: <namespace>.<entity-set>
	expression := testutils.AssertNoError(Expressions(0).EntitySet(csdlDoc, "ODataDemo.Products"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.DemoService.Products")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["DemoService"].(map[string]any)["Products"])
}

func TestExpressions_EntitySet_FullyQualified_Found(t *testing.T) {
	// Full form: <namespace>.<entity-container>.<entity-set>
	expression := testutils.AssertNoError(Expressions(0).EntitySet(csdlDoc, "ODataDemo.DemoService.Products"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.DemoService.Products")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["DemoService"].(map[string]any)["Products"])
}

func TestExpressions_EntitySet_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).EntitySet(csdlDoc, "NonExistent"); err == nil {
		t.Fatal("expected error for missing entity set, got nil")
	}
}

func TestExpressions_EntitySet_WrongKind_ReturnsError(t *testing.T) {
	// MainSupplier exists in the EntityContainer but is a singleton ($Collection != true).
	if _, err := Expressions(0).EntitySet(csdlDoc, "ODataDemo.MainSupplier"); err == nil {
		t.Fatal("expected error when element is not an entity set, got nil")
	}
}

// ---- Ambiguity --------------------------------------------------------------

func TestExpressions_EntityType_Ambiguous_ReturnsError(t *testing.T) {
	// A document with the same entity type name in two namespaces causes
	// an ambiguous match when the unqualified name is used.
	ambiguous := map[string]any{
		"ns.a": map[string]any{"Order": map[string]any{"$Kind": "EntityType"}},
		"ns.b": map[string]any{"Order": map[string]any{"$Kind": "EntityType"}},
	}

	if _, err := Expressions(0).EntityType(ambiguous, "Order"); err == nil {
		t.Fatal("expected error for ambiguous match, got nil")
	}
}

func TestExpressions_ComplexType_Ambiguous_ReturnsError(t *testing.T) {
	// Same complex type name in two namespaces — unqualified lookup must fail.
	ambiguous := map[string]any{
		"ns.a": map[string]any{"Address": map[string]any{"$Kind": "ComplexType"}},
		"ns.b": map[string]any{"Address": map[string]any{"$Kind": "ComplexType"}},
	}

	if _, err := Expressions(0).ComplexType(ambiguous, "Address"); err == nil {
		t.Fatal("expected error for ambiguous match, got nil")
	}
}

func TestExpressions_EntitySet_Ambiguous_ReturnsError(t *testing.T) {
	// Two entity containers each with the same entity set name — short-form
	// lookup must fail when the result is ambiguous.
	ambiguous := map[string]any{
		"ns.a": map[string]any{
			"ServiceA": map[string]any{
				"$Kind":  "EntityContainer",
				"Orders": map[string]any{"$Collection": true, "$Type": "ns.a.Order"},
			},
		},
		"ns.b": map[string]any{
			"ServiceB": map[string]any{
				"$Kind":  "EntityContainer",
				"Orders": map[string]any{"$Collection": true, "$Type": "ns.b.Order"},
			},
		},
	}

	if _, err := Expressions(0).EntitySet(ambiguous, "Orders"); err == nil {
		t.Fatal("expected error for ambiguous match, got nil")
	}
}

func TestExpressions_EntitySet_MediumForm_Ambiguous_ReturnsError(t *testing.T) {
	// Medium form <namespace>.<entity-set>: two containers in the same namespace
	// both expose an entity set with the same name — must fail as ambiguous.
	ambiguous := map[string]any{
		"ns.a": map[string]any{
			"ServiceA": map[string]any{
				"$Kind":   "EntityContainer",
				"Reports": map[string]any{"$Collection": true, "$Type": "ns.a.Report"},
			},
			"ServiceB": map[string]any{
				"$Kind":   "EntityContainer",
				"Reports": map[string]any{"$Collection": true, "$Type": "ns.a.Report"},
			},
		},
	}

	if _, err := Expressions(0).EntitySet(ambiguous, "ns.a.Reports"); err == nil {
		t.Fatal("expected error for ambiguous medium-form match, got nil")
	}
}

// ---- ComplexType (unqualified) ----------------------------------------------

func TestExpressions_ComplexType_UnqualifiedName_Found(t *testing.T) {
	// Without a namespace prefix, a wildcard search is used; pinpoint resolves
	// the wildcard to the concrete path of the unique match.
	expression := testutils.AssertNoError(Expressions(0).ComplexType(csdlDoc, "Address"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.Address")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["Address"])
}

// ---- EnumType ---------------------------------------------------------------

func TestExpressions_EnumType_FullyQualified_Found(t *testing.T) {
	expression := testutils.AssertNoError(Expressions(0).EnumType(csdlDoc, "ODataDemo.FileAccess"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.FileAccess")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["FileAccess"])
}

func TestExpressions_EnumType_UnqualifiedName_Found(t *testing.T) {
	// Without a namespace prefix, the wildcard path must resolve to the unique match.
	expression := testutils.AssertNoError(Expressions(0).EnumType(csdlDoc, "FileAccess"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.FileAccess")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["FileAccess"])
}

func TestExpressions_EnumType_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).EnumType(csdlDoc, "ODataDemo.NonExistent"); err == nil {
		t.Fatal("expected error for missing enum type, got nil")
	}
}

func TestExpressions_EnumType_WrongKind_ReturnsError(t *testing.T) {
	// Product is an EntityType, not an EnumType.
	if _, err := Expressions(0).EnumType(csdlDoc, "ODataDemo.Product"); err == nil {
		t.Fatal("expected error for wrong $Kind, got nil")
	}
}

// ---- EnumTypeMember ---------------------------------------------------------

func TestExpressions_EnumTypeMember_Found(t *testing.T) {
	// EnumTypeMember returns the parent expression (the EnumType path), not a
	// child path — per the OData CSDL JSON spec, member values live inside the
	// EnumType object itself.
	expression := testutils.AssertNoError(Expressions(0).EnumTypeMember(csdlDoc, "ODataDemo.FileAccess", "Read"))

	testutils.AssertExpr(t, expression, "$.ODataDemo.FileAccess")
	testutils.AssertResolvesToNode(t, csdlDoc, expression, csdlDoc["ODataDemo"].(map[string]any)["FileAccess"])
}

func TestExpressions_EnumTypeMember_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).EnumTypeMember(csdlDoc, "ODataDemo.FileAccess", "NonExistent"); err == nil {
		t.Fatal("expected error for missing enum member, got nil")
	}
}

func TestExpressions_EnumTypeMember_ParentNotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).EnumTypeMember(csdlDoc, "ODataDemo.NonExistent", "Read"); err == nil {
		t.Fatal("expected error when parent enum type does not exist, got nil")
	}
}

// ---- Resolve ----------------------------------------------------------------

func TestResolve_Root_ReturnsRootExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{Root: utils.Ptr(true)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$")
	testutils.AssertResolvesToNode(t, csdlDoc, expr, csdlDoc)
}

func TestResolve_JSONPath_ReturnsMatchingExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{JSONPath: "$.ODataDemo.Product"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.Product")
}

func TestResolve_JSONPath_InvalidSyntax_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{JSONPath: "$$[invalid"}); err == nil {
		t.Fatal("expected error for invalid JSONPath, got nil")
	}
}

func TestResolve_Operation_ReturnsOperationExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{Operation: "ODataDemo.ProductsByRating"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.ProductsByRating[0]")
}

func TestResolve_Operation_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{Operation: "ODataDemo.NonExistent"}); err == nil {
		t.Fatal("expected error for missing operation, got nil")
	}
}

func TestResolve_OperationParameter_ReturnsParameterExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		Operation: "ODataDemo.ProductsByRating",
		Parameter: "Rating",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.ProductsByRating[0]['$Parameter'][?(@['$Name'] == 'Rating')]")
}

func TestResolve_OperationParameter_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		Operation: "ODataDemo.ProductsByRating",
		Parameter: "NonExistent",
	}); err == nil {
		t.Fatal("expected error for missing parameter, got nil")
	}
}

func TestResolve_EntityType_ReturnsEntityTypeExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{EntityType: "ODataDemo.Product"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.Product")
}

func TestResolve_EntityType_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{EntityType: "ODataDemo.NonExistent"}); err == nil {
		t.Fatal("expected error for missing entity type, got nil")
	}
}

func TestResolve_EntityTypeProperty_ReturnsPropertyExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		EntityType:   "ODataDemo.Product",
		PropertyType: "Description",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.Product.Description")
}

func TestResolve_EntityTypeProperty_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		EntityType:   "ODataDemo.Product",
		PropertyType: "NonExistent",
	}); err == nil {
		t.Fatal("expected error for missing entity type property, got nil")
	}
}

func TestResolve_ComplexType_ReturnsComplexTypeExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{ComplexType: "ODataDemo.Address"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.Address")
}

func TestResolve_ComplexType_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{ComplexType: "ODataDemo.NonExistent"}); err == nil {
		t.Fatal("expected error for missing complex type, got nil")
	}
}

func TestResolve_ComplexTypeProperty_ReturnsPropertyExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		ComplexType:  "ODataDemo.Address",
		PropertyType: "Street",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.Address.Street")
}

func TestResolve_ComplexTypeProperty_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		ComplexType:  "ODataDemo.Address",
		PropertyType: "NonExistent",
	}); err == nil {
		t.Fatal("expected error for missing complex type property, got nil")
	}
}

func TestResolve_EnumType_ReturnsEnumTypeExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{EnumType: "ODataDemo.FileAccess"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.FileAccess")
}

func TestResolve_EnumType_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{EnumType: "ODataDemo.NonExistent"}); err == nil {
		t.Fatal("expected error for missing enum type, got nil")
	}
}

func TestResolve_EnumTypeMember_ReturnsEnumTypeExpression(t *testing.T) {
	// EnumTypeMember returns the parent EnumType expression per the OData spec.
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		EnumType:     "ODataDemo.FileAccess",
		PropertyType: "Read",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.FileAccess")
}

func TestResolve_EnumTypeMember_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{
		EnumType:     "ODataDemo.FileAccess",
		PropertyType: "NonExistent",
	}); err == nil {
		t.Fatal("expected error for missing enum member, got nil")
	}
}

func TestResolve_EntitySet_ReturnsEntitySetExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{EntitySet: "ODataDemo.DemoService.Products"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo.DemoService.Products")
}

func TestResolve_EntitySet_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{EntitySet: "NonExistent"}); err == nil {
		t.Fatal("expected error for missing entity set, got nil")
	}
}

func TestResolve_Namespace_ReturnsNamespaceExpression(t *testing.T) {
	expr, err := Expressions(0).Resolve(csdlDoc, &model.Selector{Namespace: "ODataDemo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, expr, "$.ODataDemo")
}

func TestResolve_Namespace_NotFound_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{Namespace: "NonExistent"}); err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
}

func TestResolve_UnsupportedSelector_ReturnsError(t *testing.T) {
	if _, err := Expressions(0).Resolve(csdlDoc, &model.Selector{}); err == nil {
		t.Fatal("expected error for empty/unsupported selector, got nil")
	}
}
