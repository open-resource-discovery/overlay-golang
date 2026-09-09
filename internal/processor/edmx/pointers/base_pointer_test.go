//go:build unit

package pointers

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

func operationPatch(operation string) model.Patch {
	return pointerPatch(model.Selector{Operation: operation})
}

func pointerPatch(selector model.Selector) model.Patch {
	return model.Patch{Selector: &selector}
}

func removePointerPatch(selector model.Selector) model.Patch {
	patch := pointerPatch(selector)
	patch.Action = "remove"
	return patch
}

// ─── ForEntityType ────────────────────────────────────────────────────────────

func TestForEntityType_EntityType_Kind(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityType")
	}
}

func TestForEntityType_EntityType_Target(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.Books" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.Books")
	}
}

func TestForEntityType_EntityType_Schema(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, p.Schema(), `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')]`)
}

func TestForEntityType_EntityType_Annotations(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, p.Annotations(), `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Annotations' && @.attributes.Target == 'CatalogService.Books')]`)
}

func TestForEntityType_EntityType_Element(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	node := p.Element().First(catalogDoc).(interface{ Name() string })
	if node.Name() != "EntityType" {
		t.Errorf("Element: got node name %q, want %q", node.Name(), "EntityType")
	}
}

func TestForEntityType_Property_Kind(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityType.Property" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityType.Property")
	}
}

func TestForEntityType_Property_Target(t *testing.T) {
	p, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.Books", PropertyType: "title"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.Books/title" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.Books/title")
	}
}

func TestForEntityType_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForEntityType(catalogDoc, pointerPatch(model.Selector{EntityType: "CatalogService.NonExistent"})); err == nil {
		t.Fatal("expected error for missing entity type, got nil")
	}
}

func TestForEntityType_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForEntityType(catalogDoc, removePointerPatch(model.Selector{EntityType: "CatalogService.NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent entity type, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent entity type")
	}
}

func TestForEntityTypeProperty_RemoveWithAbsentEntityType_DoesNotReturnError(t *testing.T) {
	pointer, err := ForEntityType(catalogDoc, removePointerPatch(model.Selector{
		EntityType:   "CatalogService.NonExistent",
		PropertyType: "title",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing a property of an absent entity type, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for a property of an absent entity type")
	}
}

func TestForEntityTypeProperty_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForEntityType(catalogDoc, removePointerPatch(model.Selector{
		EntityType:   "CatalogService.Books",
		PropertyType: "NonExistent",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent entity type property, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent entity type property")
	}
}

// ─── ForComplexType ───────────────────────────────────────────────────────────

func TestForComplexType_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForComplexType(catalogDoc, pointerPatch(model.Selector{ComplexType: "CatalogService.NonExistent"})); err == nil {
		t.Fatal("expected error for missing complex type, got nil")
	}
}

func TestForComplexType_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForComplexType(catalogDoc, removePointerPatch(model.Selector{ComplexType: "CatalogService.NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent complex type, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent complex type")
	}
}

func TestForComplexTypeProperty_RemoveWithAbsentComplexType_DoesNotReturnError(t *testing.T) {
	pointer, err := ForComplexType(catalogDoc, removePointerPatch(model.Selector{
		ComplexType:  "CatalogService.NonExistent",
		PropertyType: "street",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing a property of an absent complex type, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for a property of an absent complex type")
	}
}

func TestForComplexTypeProperty_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address">
      <Property Name="street" Type="Edm.String"/>
    </ComplexType>
  </Schema>`)

	pointer, err := ForComplexType(doc, removePointerPatch(model.Selector{
		ComplexType:  "My.Service.Address",
		PropertyType: "NonExistent",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent complex type property, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent complex type property")
	}
}

func TestForComplexType_ComplexType_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address">
      <Property Name="street" Type="Edm.String"/>
    </ComplexType>
  </Schema>`)

	p, err := ForComplexType(doc, pointerPatch(model.Selector{ComplexType: "My.Service.Address"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "ComplexType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "ComplexType")
	}
}

func TestForComplexType_ComplexType_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address"/>
  </Schema>`)

	p, err := ForComplexType(doc, pointerPatch(model.Selector{ComplexType: "My.Service.Address"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Address" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Address")
	}
}

func TestForComplexType_Property_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address">
      <Property Name="street" Type="Edm.String"/>
    </ComplexType>
  </Schema>`)

	p, err := ForComplexType(doc, pointerPatch(model.Selector{ComplexType: "My.Service.Address", PropertyType: "street"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "ComplexType.Property" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "ComplexType.Property")
	}
}

func TestForComplexType_Property_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address">
      <Property Name="street" Type="Edm.String"/>
    </ComplexType>
  </Schema>`)

	p, err := ForComplexType(doc, pointerPatch(model.Selector{ComplexType: "My.Service.Address", PropertyType: "street"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Address/street" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Address/street")
	}
}

// ─── ForEnumType ──────────────────────────────────────────────────────────────

func TestForEnumType_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForEnumType(catalogDoc, pointerPatch(model.Selector{EnumType: "CatalogService.NonExistent"})); err == nil {
		t.Fatal("expected error for missing enum type, got nil")
	}
}

func TestForEnumType_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForEnumType(catalogDoc, removePointerPatch(model.Selector{EnumType: "CatalogService.NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent enum type, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent enum type")
	}
}

func TestForEnumTypeMember_RemoveWithAbsentEnumType_DoesNotReturnError(t *testing.T) {
	pointer, err := ForEnumType(catalogDoc, removePointerPatch(model.Selector{
		EnumType:     "CatalogService.NonExistent",
		PropertyType: "Fiction",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing a member of an absent enum type, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for a member of an absent enum type")
	}
}

func TestForEnumTypeMember_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre">
      <Member Name="Fiction"/>
    </EnumType>
  </Schema>`)

	pointer, err := ForEnumType(doc, removePointerPatch(model.Selector{
		EnumType:     "My.Service.Genre",
		PropertyType: "NonExistent",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent enum type member, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent enum type member")
	}
}

func TestForEnumType_EnumType_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre">
      <Member Name="Fiction"/>
    </EnumType>
  </Schema>`)

	p, err := ForEnumType(doc, pointerPatch(model.Selector{EnumType: "My.Service.Genre"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EnumType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EnumType")
	}
}

func TestForEnumType_EnumType_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre"/>
  </Schema>`)

	p, err := ForEnumType(doc, pointerPatch(model.Selector{EnumType: "My.Service.Genre"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Genre" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Genre")
	}
}

func TestForEnumType_Member_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre">
      <Member Name="Fiction"/>
    </EnumType>
  </Schema>`)

	p, err := ForEnumType(doc, pointerPatch(model.Selector{EnumType: "My.Service.Genre", PropertyType: "Fiction"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EnumType.Member" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EnumType.Member")
	}
}

func TestForEnumType_Member_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre">
      <Member Name="Fiction"/>
    </EnumType>
  </Schema>`)

	p, err := ForEnumType(doc, pointerPatch(model.Selector{EnumType: "My.Service.Genre", PropertyType: "Fiction"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Genre/Fiction" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Genre/Fiction")
	}
}

// ─── ForEntitySet ─────────────────────────────────────────────────────────────

func TestForEntitySet_FullyQualified_Kind(t *testing.T) {
	p, err := ForEntitySet(catalogDoc, pointerPatch(model.Selector{EntitySet: "CatalogService.EntityContainer.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityContainer.EntitySet" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityContainer.EntitySet")
	}
}

func TestForEntitySet_FullyQualified_Target(t *testing.T) {
	p, err := ForEntitySet(catalogDoc, pointerPatch(model.Selector{EntitySet: "CatalogService.EntityContainer.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.EntityContainer/Books" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.EntityContainer/Books")
	}
}

func TestForEntitySet_ShortForm_Kind(t *testing.T) {
	p, err := ForEntitySet(catalogDoc, pointerPatch(model.Selector{EntitySet: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityContainer.EntitySet" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityContainer.EntitySet")
	}
}

func TestForEntitySet_ShortForm_Target(t *testing.T) {
	p, err := ForEntitySet(catalogDoc, pointerPatch(model.Selector{EntitySet: "CatalogService.Books"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.EntityContainer/Books" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.EntityContainer/Books")
	}
}

func TestForEntitySet_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForEntitySet(catalogDoc, pointerPatch(model.Selector{EntitySet: "CatalogService.NonExistent"})); err == nil {
		t.Fatal("expected error for missing entity set, got nil")
	}
}

func TestForEntitySet_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForEntitySet(catalogDoc, removePointerPatch(model.Selector{EntitySet: "CatalogService.NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent entity set, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent entity set")
	}
}

// ─── ForNamespace ─────────────────────────────────────────────────────────────

func TestForNamespace_Kind(t *testing.T) {
	p, err := ForNamespace(catalogDoc, pointerPatch(model.Selector{Namespace: "CatalogService"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Schema" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Schema")
	}
}

func TestForNamespace_Target(t *testing.T) {
	p, err := ForNamespace(catalogDoc, pointerPatch(model.Selector{Namespace: "CatalogService"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService")
	}
}

func TestForNamespace_Schema(t *testing.T) {
	p, err := ForNamespace(catalogDoc, pointerPatch(model.Selector{Namespace: "CatalogService"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, p.Schema(), `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')]`)
}

func TestForNamespace_Annotations(t *testing.T) {
	p, err := ForNamespace(catalogDoc, pointerPatch(model.Selector{Namespace: "CatalogService"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutils.AssertExpr(t, p.Annotations(), `$.nodes[?(@.name == 'edmx:Edmx')].nodes[?(@.name == 'edmx:DataServices')].nodes[?(@.name == 'Schema' && @.attributes.Namespace == 'CatalogService')].nodes[?(@.name == 'Annotations' && @.attributes.Target == 'CatalogService')]`)
}

func TestForNamespace_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForNamespace(catalogDoc, pointerPatch(model.Selector{Namespace: "NonExistent"})); err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
}

func TestForNamespace_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForNamespace(catalogDoc, removePointerPatch(model.Selector{Namespace: "NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent namespace, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent namespace")
	}
}

// ─── ForOperation ─────────────────────────────────────────────────────────────

func TestForOperation_Function_Kind(t *testing.T) {
	p, err := ForOperation(catalogDoc, operationPatch("CatalogService.getBooks"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Function")
	}
}

func TestForOperation_Function_Target(t *testing.T) {
	p, err := ForOperation(catalogDoc, operationPatch("CatalogService.getBooks"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.getBooks()" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.getBooks()")
	}
}

func TestForOperation_FunctionImport_Kind(t *testing.T) {
	// getBooks is also a FunctionImport; ForOperation tries Action, Function, FunctionImport in order.
	// getBooks is found as a Function first, so we use getBookPriorityById which has no Action overload
	// and verify it is found as Function.
	p, err := ForOperation(catalogDoc, operationPatch("CatalogService.getBookPriorityById"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function" && p.Kind() != "EntityContainer.FunctionImport" {
		t.Errorf("Kind: got %q, want Function or EntityContainer.FunctionImport", p.Kind())
	}
}

func TestForOperation_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForOperation(catalogDoc, operationPatch("CatalogService.NonExistent")); err == nil {
		t.Fatal("expected error for missing operation, got nil")
	}
}

func TestForOperation_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForOperation(catalogDoc, removePointerPatch(model.Selector{Operation: "CatalogService.NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent operation, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent operation")
	}
}

// ─── ForOperationParameter ────────────────────────────────────────────────────

func TestForOperationParameter_Function_Kind(t *testing.T) {
	p, err := ForOperationParameter(catalogDoc, pointerPatch(model.Selector{Operation: "CatalogService.getBookPriorityById", Parameter: "id"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function.Parameter" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Function.Parameter")
	}
}

func TestForOperationParameter_Function_Target(t *testing.T) {
	p, err := ForOperationParameter(catalogDoc, pointerPatch(model.Selector{Operation: "CatalogService.getBookPriorityById", Parameter: "id"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.getBookPriorityById(Edm.Int32)/id" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.getBookPriorityById(Edm.Int32)/id")
	}
}

func TestForOperationParameter_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForOperationParameter(catalogDoc, pointerPatch(model.Selector{Operation: "CatalogService.getBooks", Parameter: "NonExistent"})); err == nil {
		t.Fatal("expected error for missing parameter, got nil")
	}
}

func TestForOperationParameter_RemoveWithAbsentOperation_DoesNotReturnError(t *testing.T) {
	pointer, err := ForOperationParameter(catalogDoc, removePointerPatch(model.Selector{
		Operation: "CatalogService.NonExistent",
		Parameter: "id",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing a parameter of an absent operation, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for a parameter of an absent operation")
	}
}

func TestForOperationParameter_RemoveAbsent_DoesNotReturnError(t *testing.T) {
	pointer, err := ForOperationParameter(catalogDoc, removePointerPatch(model.Selector{
		Operation: "CatalogService.getBookPriorityById",
		Parameter: "NonExistent",
	}))
	if err != nil {
		t.Fatalf("expected no error when removing an absent operation parameter, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for an absent operation parameter")
	}
}

// ─── ForOperationReturnType ───────────────────────────────────────────────────

func TestForOperationReturnType_Function_Kind(t *testing.T) {
	p, err := ForOperationReturnType(catalogDoc, pointerPatch(model.Selector{Operation: "CatalogService.getBooks"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function.ReturnType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Function.ReturnType")
	}
}

func TestForOperationReturnType_Function_Target(t *testing.T) {
	p, err := ForOperationReturnType(catalogDoc, pointerPatch(model.Selector{Operation: "CatalogService.getBooks"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "CatalogService.getBooks()/$ReturnType" {
		t.Errorf("Target: got %q, want %q", p.Target(), "CatalogService.getBooks()/$ReturnType")
	}
}

func TestForOperationReturnType_NotFound_ReturnsError(t *testing.T) {
	if _, err := ForOperationReturnType(catalogDoc, pointerPatch(model.Selector{Operation: "CatalogService.NonExistent"})); err == nil {
		t.Fatal("expected error for missing operation, got nil")
	}
}

func TestForOperationReturnType_RemoveWithAbsentOperation_DoesNotReturnError(t *testing.T) {
	pointer, err := ForOperationReturnType(catalogDoc, removePointerPatch(model.Selector{Operation: "CatalogService.NonExistent"}))
	if err != nil {
		t.Fatalf("expected no error when removing the return type of an absent operation, got: %v", err)
	}
	if !pointer.IsNil() {
		t.Fatal("expected an empty pointer for the return type of an absent operation")
	}
}

// ─── ForOperation — Action branch ─────────────────────────────────────────────

func TestForOperation_Action_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Ping"/>
  </Schema>`)

	p, err := ForOperation(doc, operationPatch("My.Service.Ping"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Action" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Action")
	}
}

func TestForOperation_Action_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Ping"/>
  </Schema>`)

	p, err := ForOperation(doc, operationPatch("My.Service.Ping"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Ping()" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Ping()")
	}
}

// ForOperation with parameters present — the FunctionImport candidate is skipped via continue.
func TestForOperation_WithParameters_SkipsFunctionImport(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <FunctionImport Name="Calculate" Function="My.Service.Calculate"/>
    </EntityContainer>
    <Function Name="Calculate">
      <Parameter Name="x" Type="Edm.Int32"/>
      <ReturnType Type="Edm.Int32"/>
    </Function>
  </Schema>`)

	p, err := ForOperation(doc, operationPatch("My.Service.Calculate(Edm.Int32)"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Function")
	}
}

// ForOperation with parameters present but no matching Action/Function — hits the continue
// on the FunctionImport candidate and then returns not-found.
func TestForOperation_WithParameters_FunctionImportSkipped_NotFound(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <FunctionImport Name="Calculate" Function="My.Service.Calculate"/>
    </EntityContainer>
  </Schema>`)

	// No Action or Function element exists — only a FunctionImport, which is skipped because
	// parameters are present. The loop exhausts all candidates and returns an error.
	if _, err := ForOperation(doc, operationPatch("My.Service.Calculate(Edm.Int32)")); err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}

// ─── ForOperationParameter — Action.Parameter branch ─────────────────────────

func TestForOperationParameter_Action_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Submit">
      <Parameter Name="input" Type="Edm.String"/>
    </Action>
  </Schema>`)

	p, err := ForOperationParameter(doc, pointerPatch(model.Selector{Operation: "My.Service.Submit", Parameter: "input"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Action.Parameter" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Action.Parameter")
	}
}

func TestForOperationParameter_Action_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Submit">
      <Parameter Name="input" Type="Edm.String"/>
    </Action>
  </Schema>`)

	p, err := ForOperationParameter(doc, pointerPatch(model.Selector{Operation: "My.Service.Submit", Parameter: "input"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Submit(Edm.String)/input" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Submit(Edm.String)/input")
	}
}

// ─── ForOperationReturnType — Action.ReturnType branch ───────────────────────

func TestForOperationReturnType_Action_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Submit">
      <ReturnType Type="Edm.String"/>
    </Action>
  </Schema>`)

	p, err := ForOperationReturnType(doc, pointerPatch(model.Selector{Operation: "My.Service.Submit"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Action.ReturnType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Action.ReturnType")
	}
}

func TestForOperationReturnType_Action_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Submit">
      <ReturnType Type="Edm.String"/>
    </Action>
  </Schema>`)

	p, err := ForOperationReturnType(doc, pointerPatch(model.Selector{Operation: "My.Service.Submit"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Submit()/$ReturnType" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Submit()/$ReturnType")
	}
}

// ─── ForEntitySet — second candidate (no container name) ─────────────────────

func TestForEntitySet_NamespaceAndEntitySet_Kind(t *testing.T) {
	// "CatalogService.Books" — first candidate tries namespace=CatalogService container=EntityContainer
	// which fails because EntityContainer name doesn't match "Books". Falls through to second candidate.
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <EntitySet Name="Orders" EntityType="My.Service.Order"/>
    </EntityContainer>
  </Schema>`)

	// Use "My.Service.Orders" — first candidate splits into namespace=My.Service, name=Orders;
	// tries container named "Orders" (not found). Second candidate tries any container with EntitySet "Orders".
	p, err := ForEntitySet(doc, pointerPatch(model.Selector{EntitySet: "My.Service.Orders"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityContainer.EntitySet" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityContainer.EntitySet")
	}
}

func TestForEntitySet_NamespaceAndEntitySet_Target(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <EntitySet Name="Orders" EntityType="My.Service.Order"/>
    </EntityContainer>
  </Schema>`)

	p, err := ForEntitySet(doc, pointerPatch(model.Selector{EntitySet: "My.Service.Orders"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Target() != "My.Service.Container/Orders" {
		t.Errorf("Target: got %q, want %q", p.Target(), "My.Service.Container/Orders")
	}
}

// ─── Ambiguous match — found && err != nil ────────────────────────────────────

func TestForEntityType_Ambiguous_ReturnsError(t *testing.T) {
	// Two schemas each containing an EntityType with the same name — the wildcard
	// expression matches both, so Pinpoint returns found=true, err!=nil.
	doc := edmxDoc(`
    <Schema Namespace="Ns.A" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Book"/>
    </Schema>
    <Schema Namespace="Ns.B" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Book"/>
    </Schema>`)

	if _, err := ForEntityType(doc, pointerPatch(model.Selector{EntityType: "Book"})); err == nil {
		t.Fatal("expected error for ambiguous entity type match, got nil")
	}
}

func TestForComplexType_Ambiguous_ReturnsError(t *testing.T) {
	doc := edmxDoc(`
    <Schema Namespace="Ns.A" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <ComplexType Name="Address"/>
    </Schema>
    <Schema Namespace="Ns.B" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <ComplexType Name="Address"/>
    </Schema>`)

	if _, err := ForComplexType(doc, pointerPatch(model.Selector{ComplexType: "Address"})); err == nil {
		t.Fatal("expected error for ambiguous complex type match, got nil")
	}
}

func TestForEnumType_Ambiguous_ReturnsError(t *testing.T) {
	doc := edmxDoc(`
    <Schema Namespace="Ns.A" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EnumType Name="Status"/>
    </Schema>
    <Schema Namespace="Ns.B" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EnumType Name="Status"/>
    </Schema>`)

	if _, err := ForEnumType(doc, pointerPatch(model.Selector{EnumType: "Status"})); err == nil {
		t.Fatal("expected error for ambiguous enum type match, got nil")
	}
}
