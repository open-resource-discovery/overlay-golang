//go:build unit

package pointers

import (
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

// edmxDoc wraps an EDMX XML fragment in the minimal envelope that expressions.*
// expects: edmx:Edmx > edmx:DataServices > Schema(s).
func edmxDoc(schemasXML string) xml2json.Document {
	doc, err := xml2json.Convert(`<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>` + schemasXML + `</edmx:DataServices>
</edmx:Edmx>`)
	if err != nil {
		panic(err)
	}
	return doc
}

// ─── ParseQualifiedName ───────────────────────────────────────────────────────

func TestParseQualifiedName_SimpleNameNoParameters(t *testing.T) {
	ns, name, params := resolvers.ParseQualifiedName("My.Namespace.MyType")
	if ns != "My.Namespace" {
		t.Errorf("namespace: got %q, want %q", ns, "My.Namespace")
	}
	if name != "MyType" {
		t.Errorf("name: got %q, want %q", name, "MyType")
	}
	if params != nil {
		t.Errorf("parameters: got %v, want nil", params)
	}
}

func TestParseQualifiedName_WithParameters(t *testing.T) {
	ns, name, params := resolvers.ParseQualifiedName("My.Namespace.MyFunc(Edm.String,Edm.Int32)")
	if ns != "My.Namespace" {
		t.Errorf("namespace: got %q, want %q", ns, "My.Namespace")
	}
	if name != "MyFunc" {
		t.Errorf("name: got %q, want %q", name, "MyFunc")
	}
	if len(params) != 2 || params[0] != "Edm.String" || params[1] != "Edm.Int32" {
		t.Errorf("parameters: got %v, want [Edm.String Edm.Int32]", params)
	}
}

func TestParseQualifiedName_ParametersWithSpaces(t *testing.T) {
	_, _, params := resolvers.ParseQualifiedName("NS.Func( Edm.String , Edm.Int32 )")
	if len(params) != 2 || params[0] != "Edm.String" || params[1] != "Edm.Int32" {
		t.Errorf("parameters: got %v, want trimmed [Edm.String Edm.Int32]", params)
	}
}

func TestParseQualifiedName_EmptyParameters(t *testing.T) {
	_, _, params := resolvers.ParseQualifiedName("NS.Func()")
	if params == nil || len(params) != 0 {
		t.Errorf("parameters: got %v, want []", params)
	}
}

func TestParseQualifiedName_SingleSegmentName(t *testing.T) {
	ns, name, params := resolvers.ParseQualifiedName("Root")
	if ns != "" {
		t.Errorf("namespace: got %q, want %q", ns, "")
	}
	if name != "Root" {
		t.Errorf("name: got %q, want %q", name, "Root")
	}
	if params != nil {
		t.Errorf("parameters: got %v, want nil", params)
	}
}

func TestParseQualifiedName_SingleParameter(t *testing.T) {
	_, _, params := resolvers.ParseQualifiedName("NS.Func(Edm.Guid)")
	if len(params) != 1 || params[0] != "Edm.Guid" {
		t.Errorf("parameters: got %v, want [Edm.Guid]", params)
	}
}

// ─── helpers for ResolveNamespace / ResolveAnnotationsTarget ─────────────────

// ─── ResolveNamespace ─────────────────────────────────────────────────────────

func TestResolveNamespace_Schema(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm"/>`)

	expr := expressions.Schema("My.Service")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_EntityType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityType Name="Book"/>
  </Schema>`)

	expr := expressions.EntityType("My.Service", "Book", "")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_EntityTypeProperty(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityType Name="Book">
      <Property Name="title" Type="Edm.String"/>
    </EntityType>
  </Schema>`)

	expr := expressions.EntityType("My.Service", "Book", "title")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_ComplexType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address"/>
  </Schema>`)

	expr := expressions.ComplexType("My.Service", "Address", "")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_ComplexTypeProperty(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address">
      <Property Name="street" Type="Edm.String"/>
    </ComplexType>
  </Schema>`)

	expr := expressions.ComplexType("My.Service", "Address", "street")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_EnumType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre"/>
  </Schema>`)

	expr := expressions.EnumType("My.Service", "Genre", "")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_EnumTypeMember(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre">
      <Member Name="Fiction"/>
    </EnumType>
  </Schema>`)

	expr := expressions.EnumType("My.Service", "Genre", "Fiction")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_EntityContainer(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <FunctionImport Name="GetAll" Function="My.Service.GetAll"/>
    </EntityContainer>
  </Schema>`)

	expr := expressions.FunctionImport("My.Service", "GetAll")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_EntitySet(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <EntitySet Name="Books" EntityType="My.Service.Book"/>
    </EntityContainer>
  </Schema>`)

	expr := expressions.EntitySet("My.Service", "Container", "Books")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveNamespace(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveNamespace_ActionParameter(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="CreateBook">
      <Parameter Name="input" Type="My.Service.BookInput"/>
      <ReturnType Type="My.Service.Book"/>
    </Action>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.ActionParameter("My.Service", "CreateBook", nil, "input"),
		expressions.ActionParameter("My.Service", "CreateBook", []string{"My.Service.BookInput"}, "input"),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveNamespace(doc, pexpr)
		if got != "My.Service" {
			t.Errorf("got %q, want %q", got, "My.Service")
		}
	}
}

func TestResolveNamespace_ActionReturnType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="CreateBook">
      <ReturnType Type="My.Service.Book"/>
    </Action>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.ActionReturnType("My.Service", "CreateBook", nil),
		expressions.ActionReturnType("My.Service", "CreateBook", []string{}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveNamespace(doc, pexpr)
		if got != "My.Service" {
			t.Errorf("got %q, want %q", got, "My.Service")
		}
	}
}

func TestResolveNamespace_FunctionParameter(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Function Name="GetBook">
      <Parameter Name="id" Type="Edm.Int32"/>
      <ReturnType Type="My.Service.Book"/>
    </Function>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.FunctionParameter("My.Service", "GetBook", nil, "id"),
		expressions.FunctionParameter("My.Service", "GetBook", []string{"Edm.Int32"}, "id"),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveNamespace(doc, pexpr)
		if got != "My.Service" {
			t.Errorf("got %q, want %q", got, "My.Service")
		}
	}
}

func TestResolveNamespace_FunctionReturnType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Function Name="GetBook">
      <ReturnType Type="My.Service.Book"/>
    </Function>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.FunctionReturnType("My.Service", "GetBook", nil),
		expressions.FunctionReturnType("My.Service", "GetBook", []string{}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveNamespace(doc, pexpr)
		if got != "My.Service" {
			t.Errorf("got %q, want %q", got, "My.Service")
		}
	}
}

// ─── ResolveAnnotationsTarget ─────────────────────────────────────────────────

func TestResolveAnnotationsTarget_Schema(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm"/>`)

	expr := expressions.Schema("My.Service")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service" {
		t.Errorf("got %q, want %q", got, "My.Service")
	}
}

func TestResolveAnnotationsTarget_EntityType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityType Name="Book"/>
  </Schema>`)

	expr := expressions.EntityType("My.Service", "Book", "")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Book" {
		t.Errorf("got %q, want %q", got, "My.Service.Book")
	}
}

func TestResolveAnnotationsTarget_EntityTypeProperty(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityType Name="Book">
      <Property Name="title" Type="Edm.String"/>
    </EntityType>
  </Schema>`)

	expr := expressions.EntityType("My.Service", "Book", "title")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Book/title" {
		t.Errorf("got %q, want %q", got, "My.Service.Book/title")
	}
}

func TestResolveAnnotationsTarget_ComplexType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address"/>
  </Schema>`)

	expr := expressions.ComplexType("My.Service", "Address", "")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Address" {
		t.Errorf("got %q, want %q", got, "My.Service.Address")
	}
}

func TestResolveAnnotationsTarget_ComplexTypeProperty(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <ComplexType Name="Address">
      <Property Name="street" Type="Edm.String"/>
    </ComplexType>
  </Schema>`)

	expr := expressions.ComplexType("My.Service", "Address", "street")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Address/street" {
		t.Errorf("got %q, want %q", got, "My.Service.Address/street")
	}
}

func TestResolveAnnotationsTarget_EnumType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre"/>
  </Schema>`)

	expr := expressions.EnumType("My.Service", "Genre", "")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Genre" {
		t.Errorf("got %q, want %q", got, "My.Service.Genre")
	}
}

func TestResolveAnnotationsTarget_EnumTypeMember(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EnumType Name="Genre">
      <Member Name="Fiction"/>
    </EnumType>
  </Schema>`)

	expr := expressions.EnumType("My.Service", "Genre", "Fiction")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Genre/Fiction" {
		t.Errorf("got %q, want %q", got, "My.Service.Genre/Fiction")
	}
}

func TestResolveAnnotationsTarget_EntityContainer(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <FunctionImport Name="GetAll" Function="My.Service.GetAll"/>
    </EntityContainer>
  </Schema>`)

	// Reach EntityContainer via a FunctionImport, then trim to parent to get the container itself.
	// Easier: use EntitySet expression with empty name, but that requires a set.
	// Use FunctionImport to confirm the EntityContainer path via ResolveAnnotationsTarget branch.
	expr := expressions.FunctionImport("My.Service", "GetAll")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Container/GetAll" {
		t.Errorf("got %q, want %q", got, "My.Service.Container/GetAll")
	}
}

func TestResolveAnnotationsTarget_EntitySet(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <EntityContainer Name="Container">
      <EntitySet Name="Books" EntityType="My.Service.Book"/>
    </EntityContainer>
  </Schema>`)

	expr := expressions.EntitySet("My.Service", "Container", "Books")
	pexpr, found, err := xml2json.Pinpoint(doc, expr)
	if !found || err != nil {
		t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
	}

	got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
	if got != "My.Service.Container/Books" {
		t.Errorf("got %q, want %q", got, "My.Service.Container/Books")
	}
}

func TestResolveAnnotationsTarget_Action(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="CreateBook">
      <Parameter Name="input" Type="My.Service.BookInput"/>
    </Action>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.Action("My.Service", "CreateBook", nil),
		expressions.Action("My.Service", "CreateBook", []string{"My.Service.BookInput"}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		// Signature is built from child Parameter nodes: CreateBook(My.Service.BookInput)
		if got != "My.Service.CreateBook(My.Service.BookInput)" {
			t.Errorf("got %q, want %q", got, "My.Service.CreateBook(My.Service.BookInput)")
		}
	}
}

func TestResolveAnnotationsTarget_ActionNoParameters(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="Ping"/>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.Action("My.Service", "Ping", nil),
		expressions.Action("My.Service", "Ping", []string{}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		if got != "My.Service.Ping()" {
			t.Errorf("got %q, want %q", got, "My.Service.Ping()")
		}
	}
}

func TestResolveAnnotationsTarget_Function(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Function Name="GetBook">
      <Parameter Name="id" Type="Edm.Int32"/>
      <ReturnType Type="My.Service.Book"/>
    </Function>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.Function("My.Service", "GetBook", nil),
		expressions.Function("My.Service", "GetBook", []string{"Edm.Int32"}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		// Signature built from child Parameter nodes: GetBook(Edm.Int32)
		if got != "My.Service.GetBook(Edm.Int32)" {
			t.Errorf("got %q, want %q", got, "My.Service.GetBook(Edm.Int32)")
		}
	}
}

func TestResolveAnnotationsTarget_ActionParameter(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="CreateBook">
      <Parameter Name="input" Type="My.Service.BookInput"/>
    </Action>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.ActionParameter("My.Service", "CreateBook", nil, "input"),
		expressions.ActionParameter("My.Service", "CreateBook", []string{"My.Service.BookInput"}, "input"),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		if got != "My.Service.CreateBook(My.Service.BookInput)/input" {
			t.Errorf("got %q, want %q", got, "My.Service.CreateBook(My.Service.BookInput)/input")
		}
	}
}

func TestResolveAnnotationsTarget_ActionReturnType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Action Name="CreateBook">
      <Parameter Name="input" Type="My.Service.BookInput"/>
      <ReturnType Type="My.Service.Book"/>
    </Action>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.ActionReturnType("My.Service", "CreateBook", nil),
		expressions.ActionReturnType("My.Service", "CreateBook", []string{"My.Service.BookInput"}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		if got != "My.Service.CreateBook(My.Service.BookInput)/$ReturnType" {
			t.Errorf("got %q, want %q", got, "My.Service.CreateBook(My.Service.BookInput)/$ReturnType")
		}
	}
}

func TestResolveAnnotationsTarget_FunctionParameter(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Function Name="GetBook">
      <Parameter Name="id" Type="Edm.Int32"/>
      <ReturnType Type="My.Service.Book"/>
    </Function>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.FunctionParameter("My.Service", "GetBook", nil, "id"),
		expressions.FunctionParameter("My.Service", "GetBook", []string{"Edm.Int32"}, "id"),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		if got != "My.Service.GetBook(Edm.Int32)/id" {
			t.Errorf("got %q, want %q", got, "My.Service.GetBook(Edm.Int32)/id")
		}
	}
}

func TestResolveAnnotationsTarget_FunctionReturnType(t *testing.T) {
	doc := edmxDoc(`<Schema Namespace="My.Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
    <Function Name="GetBook">
      <Parameter Name="id" Type="Edm.Int32"/>
      <ReturnType Type="My.Service.Book"/>
    </Function>
  </Schema>`)

	for _, candidate := range []jp.Expr{
		expressions.FunctionReturnType("My.Service", "GetBook", nil),
		expressions.FunctionReturnType("My.Service", "GetBook", []string{"Edm.Int32"}),
	} {
		pexpr, found, err := xml2json.Pinpoint(doc, candidate)
		if !found || err != nil {
			t.Fatalf("Pinpoint failed: found=%v err=%v", found, err)
		}

		got := resolvers.ResolveAnnotationsTarget(doc, pexpr)
		if got != "My.Service.GetBook(Edm.Int32)/$ReturnType" {
			t.Errorf("got %q, want %q", got, "My.Service.GetBook(Edm.Int32)/$ReturnType")
		}
	}
}
