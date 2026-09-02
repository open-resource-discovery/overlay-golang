//go:build unit

package pointers

import (
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

// catalogDoc is the parsed CatalogService EDMX fixture loaded once for all tests.
var catalogDoc = testutils.UnmarshalFixture[xml2json.Document]("../testdata/catalogservice.xml")

// ---- helpers ----------------------------------------------------------------

// resolve returns e.First(catalogDoc) cast to xml2json.Node for convenience.
func resolve(e jp.Expr) xml2json.Node {
	return e.First(catalogDoc).(xml2json.Node)
}

// ---- Schema -----------------------------------------------------------------

func TestExpressions_Schema_ByNamespace(t *testing.T) {
	expr := expressions.Schema("CatalogService")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')]`)

	node := resolve(expr)
	if node.Name() != "Schema" || node.Attribute("Namespace") != "CatalogService" {
		t.Errorf("expected Schema Namespace=CatalogService, got name=%s namespace=%s", node.Name(), node.Attribute("Namespace"))
	}
}

func TestExpressions_Schema_EmptyNamespace(t *testing.T) {
	expr := expressions.Schema("")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema')]`)

	node := resolve(expr)
	if node.Name() != "Schema" {
		t.Errorf("expected Schema node, got %s", node.Name())
	}
}

// ---- EntityType -------------------------------------------------------------

func TestExpressions_EntityType_WithoutProperty(t *testing.T) {
	expr := expressions.EntityType("CatalogService", "Books", "")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EntityType' && @.attributes.Name == 'Books')]`)

	node := resolve(expr)
	if node.Name() != "EntityType" || node.Attribute("Name") != "Books" {
		t.Errorf("expected EntityType Books, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

func TestExpressions_EntityType_WithProperty(t *testing.T) {
	expr := expressions.EntityType("CatalogService", "Books", "title")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EntityType' && @.attributes.Name == 'Books')].nodes[?(@.name == 'Property' && @.attributes.Name == 'title')]`)

	node := resolve(expr)
	if node.Name() != "Property" || node.Attribute("Name") != "title" {
		t.Errorf("expected Property title, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

// ---- ComplexType ------------------------------------------------------------

func TestExpressions_ComplexType_WithoutProperty(t *testing.T) {
	expr := expressions.ComplexType("CatalogService", "MyComplex", "")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'ComplexType' && @.attributes.Name == 'MyComplex')]`)
}

func TestExpressions_ComplexType_WithProperty(t *testing.T) {
	expr := expressions.ComplexType("CatalogService", "MyComplex", "street")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'ComplexType' && @.attributes.Name == 'MyComplex')].nodes[?(@.name == 'Property' && @.attributes.Name == 'street')]`)
}

// ---- EnumType ---------------------------------------------------------------

func TestExpressions_EnumType_WithoutMember(t *testing.T) {
	expr := expressions.EnumType("CatalogService", "Priority", "")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EnumType' && @.attributes.Name == 'Priority')]`)
}

func TestExpressions_EnumType_WithMember(t *testing.T) {
	expr := expressions.EnumType("CatalogService", "Priority", "High")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EnumType' && @.attributes.Name == 'Priority')].nodes[?(@.name == 'Member' && @.attributes.Name == 'High')]`)
}

// ---- Action -----------------------------------------------------------------

func TestExpressions_Action_WithoutParameters(t *testing.T) {
	expr := expressions.Action("CatalogService", "submitOrder", nil)

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder')]`)
}

func TestExpressions_Action_WithEmptyParameters(t *testing.T) {
	expr := expressions.Action("CatalogService", "submitOrder", []string{})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder' && count(@.nodes[?(@.name == 'Parameter')]) == 0)]`)
}

func TestExpressions_Action_WithParameters(t *testing.T) {
	expr := expressions.Action("CatalogService", "submitOrder", []string{"Edm.Int32"})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder' && count(@.nodes[?(@.name == 'Parameter')]) == 1 && @.nodes[0].attributes.Type == 'Edm.Int32')]`)
}

// ---- Function ---------------------------------------------------------------

func TestExpressions_Function_WithoutParameters(t *testing.T) {
	expr := expressions.Function("CatalogService", "getBooks", nil)

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBooks')]`)

	node := resolve(expr)
	if node.Name() != "Function" || node.Attribute("Name") != "getBooks" {
		t.Errorf("expected Function getBooks, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

func TestExpressions_Function_WithEmptyParameters(t *testing.T) {
	expr := expressions.Function("CatalogService", "getBooks", []string{})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBooks' && count(@.nodes[?(@.name == 'Parameter')]) == 0)]`)

	node := resolve(expr)
	if node.Name() != "Function" || node.Attribute("Name") != "getBooks" {
		t.Errorf("expected Function getBooks, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

func TestExpressions_Function_WithParameters(t *testing.T) {
	expr := expressions.Function("CatalogService", "getBookById", []string{"Edm.Int32"})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBookById' && count(@.nodes[?(@.name == 'Parameter')]) == 1 && @.nodes[0].attributes.Type == 'Edm.Int32')]`)

	node := resolve(expr)
	if node.Name() != "Function" || node.Attribute("Name") != "getBookById" {
		t.Errorf("expected Function getBooks, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

// ---- EntitySet --------------------------------------------------------------

func TestExpressions_EntitySet_WithContainer(t *testing.T) {
	expr := expressions.EntitySet("CatalogService", "EntityContainer", "Books")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EntityContainer' && @.attributes.Name == 'EntityContainer')].nodes[?(@.name == 'EntitySet' && @.attributes.Name == 'Books')]`)

	node := resolve(expr)
	if node.Name() != "EntitySet" || node.Attribute("Name") != "Books" {
		t.Errorf("expected EntitySet Books, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

func TestExpressions_EntitySet_EmptyContainer(t *testing.T) {
	expr := expressions.EntitySet("CatalogService", "", "Books")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EntityContainer')].nodes[?(@.name == 'EntitySet' && @.attributes.Name == 'Books')]`)

	node := resolve(expr)
	if node.Name() != "EntitySet" || node.Attribute("Name") != "Books" {
		t.Errorf("expected EntitySet Books, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

// ---- FunctionImport ---------------------------------------------------------

func TestExpressions_FunctionImport_getBooks(t *testing.T) {
	expr := expressions.FunctionImport("CatalogService", "getBooks")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EntityContainer')].nodes[?(@.name == 'FunctionImport' && @.attributes.Name == 'getBooks')]`)

	node := resolve(expr)
	if node.Name() != "FunctionImport" || node.Attribute("Name") != "getBooks" {
		t.Errorf("expected FunctionImport getBooks, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

func TestExpressions_FunctionImport_getBookById(t *testing.T) {
	expr := expressions.FunctionImport("CatalogService", "getBookById")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'EntityContainer')].nodes[?(@.name == 'FunctionImport' && @.attributes.Name == 'getBookById')]`)

	node := resolve(expr)
	if node.Name() != "FunctionImport" || node.Attribute("Name") != "getBookById" {
		t.Errorf("expected FunctionImport getBookById, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

// ---- ActionParameter --------------------------------------------------------

func TestExpressions_ActionParameter_WithoutParameters(t *testing.T) {
	expr := expressions.ActionParameter("CatalogService", "submitOrder", nil, "orderId")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder')].nodes[?(@.name == 'Parameter' && @.attributes.Name == 'orderId')]`)
}

func TestExpressions_ActionParameter_WithEmptyParameters(t *testing.T) {
	expr := expressions.ActionParameter("CatalogService", "submitOrder", []string{}, "orderId")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder' && count(@.nodes[?(@.name == 'Parameter')]) == 0)].nodes[?(@.name == 'Parameter' && @.attributes.Name == 'orderId')]`)
}

func TestExpressions_ActionParameter_WithParameters(t *testing.T) {
	expr := expressions.ActionParameter("CatalogService", "submitOrder", []string{"Edm.Int32"}, "orderId")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder' && count(@.nodes[?(@.name == 'Parameter')]) == 1 && @.nodes[0].attributes.Type == 'Edm.Int32')].nodes[?(@.name == 'Parameter' && @.attributes.Name == 'orderId')]`)
}

// ---- ActionReturnType -------------------------------------------------------

func TestExpressions_ActionReturnType_WithoutParameters(t *testing.T) {
	expr := expressions.ActionReturnType("CatalogService", "submitOrder", nil)

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder')].nodes[?(@.name == 'ReturnType')]`)
}

func TestExpressions_ActionReturnType_WithEmptyParameters(t *testing.T) {
	expr := expressions.ActionReturnType("CatalogService", "submitOrder", []string{})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder' && count(@.nodes[?(@.name == 'Parameter')]) == 0)].nodes[?(@.name == 'ReturnType')]`)
}

func TestExpressions_ActionReturnType_WithParameters(t *testing.T) {
	expr := expressions.ActionReturnType("CatalogService", "submitOrder", []string{"Edm.Int32"})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Action' && @.attributes.Name == 'submitOrder' && count(@.nodes[?(@.name == 'Parameter')]) == 1 && @.nodes[0].attributes.Type == 'Edm.Int32')].nodes[?(@.name == 'ReturnType')]`)
}

// ---- FunctionParameter ------------------------------------------------------

func TestExpressions_FunctionParameter_WithoutParameters(t *testing.T) {
	expr := expressions.FunctionParameter("CatalogService", "getBookPriorityById", nil, "id")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBookPriorityById')].nodes[?(@.name == 'Parameter' && @.attributes.Name == 'id')]`)

	node := resolve(expr)
	if node.Name() != "Parameter" || node.Attribute("Name") != "id" {
		t.Errorf("expected Parameter id, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

func TestExpressions_FunctionParameter_WithParameters(t *testing.T) {
	expr := expressions.FunctionParameter("CatalogService", "getBookPriorityById", []string{"Edm.Int32"}, "id")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBookPriorityById' && count(@.nodes[?(@.name == 'Parameter')]) == 1 && @.nodes[0].attributes.Type == 'Edm.Int32')].nodes[?(@.name == 'Parameter' && @.attributes.Name == 'id')]`)

	node := resolve(expr)
	if node.Name() != "Parameter" || node.Attribute("Name") != "id" {
		t.Errorf("expected Parameter id, got name=%s attr=%s", node.Name(), node.Attribute("Name"))
	}
}

// ---- FunctionReturnType -----------------------------------------------------

func TestExpressions_FunctionReturnType_WithoutParameters(t *testing.T) {
	expr := expressions.FunctionReturnType("CatalogService", "getBooks", nil)

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBooks')].nodes[?(@.name == 'ReturnType')]`)

	node := resolve(expr)
	if node.Name() != "ReturnType" {
		t.Errorf("expected ReturnType, got %s", node.Name())
	}
}

func TestExpressions_FunctionReturnType_WithEmptyParameters(t *testing.T) {
	expr := expressions.FunctionReturnType("CatalogService", "getBooks", []string{})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBooks' && count(@.nodes[?(@.name == 'Parameter')]) == 0)].nodes[?(@.name == 'ReturnType')]`)

	node := resolve(expr)
	if node.Name() != "ReturnType" {
		t.Errorf("expected ReturnType, got %s", node.Name())
	}
}

func TestExpressions_FunctionReturnType_WithParameters(t *testing.T) {
	expr := expressions.FunctionReturnType("CatalogService", "getBookById", []string{"Edm.Int32"})

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Function' && @.attributes.Name == 'getBookById' && count(@.nodes[?(@.name == 'Parameter')]) == 1 && @.nodes[0].attributes.Type == 'Edm.Int32')].nodes[?(@.name == 'ReturnType')]`)

	node := resolve(expr)
	if node.Name() != "ReturnType" {
		t.Errorf("expected ReturnType, got %s", node.Name())
	}
}

// ---- Annotations ------------------------------------------------------------

func TestExpressions_Annotations_EntityType(t *testing.T) {
	expr := expressions.Annotations("CatalogService", "CatalogService.Books")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Annotations' && @.attributes.Target == 'CatalogService.Books')]`)

	node := resolve(expr)
	if node.Name() != "Annotations" || node.Attribute("Target") != "CatalogService.Books" {
		t.Errorf("expected Annotations Target=CatalogService.Books, got name=%s attr=%s", node.Name(), node.Attribute("Target"))
	}
}

func TestExpressions_Annotations_EntityContainer(t *testing.T) {
	expr := expressions.Annotations("CatalogService", "CatalogService.EntityContainer/Books")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Annotations' && @.attributes.Target == 'CatalogService.EntityContainer/Books')]`)

	node := resolve(expr)
	if node.Name() != "Annotations" || node.Attribute("Target") != "CatalogService.EntityContainer/Books" {
		t.Errorf("expected Annotations Target=CatalogService.EntityContainer/Books, got name=%s attr=%s", node.Name(), node.Attribute("Target"))
	}
}

func TestExpressions_Annotations_Property(t *testing.T) {
	expr := expressions.Annotations("CatalogService", "CatalogService.Books/ID")

	testutils.AssertExpr(t, expr, `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Annotations' && @.attributes.Target == 'CatalogService.Books/ID')]`)

	node := resolve(expr)
	if node.Name() != "Annotations" || node.Attribute("Target") != "CatalogService.Books/ID" {
		t.Errorf("expected Annotations Target=CatalogService.Books/ID, got name=%s attr=%s", node.Name(), node.Attribute("Target"))
	}
}
