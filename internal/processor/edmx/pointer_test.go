//go:build unit

package edmx

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/overlay-golang/model"
)

var catalogDoc = testutils.UnmarshalFixture[xml2json.Document]("testdata/catalogservice.xml")

// edmxDoc wraps one or more Schema elements into a minimal EDMX document.
func edmxDoc(schemasXML string) xml2json.Document {
	raw := `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>` + schemasXML + `</edmx:DataServices>
</edmx:Edmx>`
	doc, err := xml2json.Convert(raw)
	if err != nil {
		panic("failed to parse inline EDMX: " + err.Error())
	}
	return doc
}

// ─── EnumType branch ──────────────────────────────────────────────────────────

func TestNewPointer_EnumType_Kind(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Status"/>
  </Schema>`)

	p, err := NewPointer(doc, &model.Selector{EnumType: "My.Service.Status"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EnumType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EnumType")
	}
}

func TestNewPointer_EnumType_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{EnumType: "CatalogService.NonExistent"}); err == nil {
		t.Fatal("expected error for missing enum type, got nil")
	}
}

// ─── Operation branch — Operation only ───────────────────────────────────────

func TestNewPointer_Operation_Kind(t *testing.T) {
	p, err := NewPointer(catalogDoc, &model.Selector{Operation: "CatalogService.getBooks"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function" && p.Kind() != "EntityContainer.FunctionImport" {
		t.Errorf("Kind: got %q, want Function or EntityContainer.FunctionImport", p.Kind())
	}
}

// ─── Operation branch — ReturnType ───────────────────────────────────────────

func TestNewPointer_OperationReturnType_Kind(t *testing.T) {
	p, err := NewPointer(catalogDoc, &model.Selector{Operation: "CatalogService.getBooks", ReturnType: utils.Ptr(true)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function.ReturnType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Function.ReturnType")
	}
}

func TestNewPointer_OperationReturnType_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{Operation: "CatalogService.NonExistent", ReturnType: utils.Ptr(true)}); err == nil {
		t.Fatal("expected error for missing operation return type, got nil")
	}
}

// ─── Operation branch — Parameter ────────────────────────────────────────────

func TestNewPointer_OperationParameter_Kind(t *testing.T) {
	// getBookPriorityById has a single overload with parameter "id"
	p, err := NewPointer(catalogDoc, &model.Selector{Operation: "CatalogService.getBookPriorityById", Parameter: "id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Function.Parameter" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Function.Parameter")
	}
}

func TestNewPointer_OperationParameter_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{Operation: "CatalogService.getBookPriorityById", Parameter: "nonExistent"}); err == nil {
		t.Fatal("expected error for missing parameter, got nil")
	}
}

// ─── EntitySet branch ─────────────────────────────────────────────────────────

func TestNewPointer_EntitySet_Kind(t *testing.T) {
	p, err := NewPointer(catalogDoc, &model.Selector{EntitySet: "CatalogService.EntityContainer.Books"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityContainer.EntitySet" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityContainer.EntitySet")
	}
}

func TestNewPointer_EntitySet_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{EntitySet: "CatalogService.EntityContainer.NonExistent"}); err == nil {
		t.Fatal("expected error for missing entity set, got nil")
	}
}

// ─── EntityType branch ────────────────────────────────────────────────────────

func TestNewPointer_EntityType_Kind(t *testing.T) {
	p, err := NewPointer(catalogDoc, &model.Selector{EntityType: "CatalogService.Books"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "EntityType" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "EntityType")
	}
}

func TestNewPointer_EntityType_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{EntityType: "CatalogService.NonExistent"}); err == nil {
		t.Fatal("expected error for missing entity type, got nil")
	}
}

// ─── ComplexType branch ───────────────────────────────────────────────────────

func TestNewPointer_ComplexType_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{ComplexType: "CatalogService.NonExistent"}); err == nil {
		t.Fatal("expected error for missing complex type, got nil")
	}
}

// ─── Namespace branch ─────────────────────────────────────────────────────────

func TestNewPointer_Namespace_Kind(t *testing.T) {
	p, err := NewPointer(catalogDoc, &model.Selector{Namespace: "CatalogService"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind() != "Schema" {
		t.Errorf("Kind: got %q, want %q", p.Kind(), "Schema")
	}
}

func TestNewPointer_Namespace_NotFound_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{Namespace: "NonExistent"}); err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
}

// ─── Unsupported selector ─────────────────────────────────────────────────────

func TestNewPointer_EmptySelector_ReturnsError(t *testing.T) {
	if _, err := NewPointer(catalogDoc, &model.Selector{}); err == nil {
		t.Fatal("expected error for empty selector, got nil")
	}
}
