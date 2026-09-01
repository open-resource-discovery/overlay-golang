//go:build unit

package edmx

import (
	"reflect"
	"strings"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

// nodeToXML serialises a single Node to an XML snippet by wrapping it in a
// minimal Document and stripping surrounding whitespace.
func nodeToXML(node xml2json.Node) string {
	doc := xml2json.NewDocument([]xml2json.Node{node}, []string{}, []string{})
	return strings.TrimSpace(doc.ToXML())
}

// assertXML compares the parsed form of node to wantXML.
func assertXML(t *testing.T, node xml2json.Node, wantXML string) {
	t.Helper()

	wanted, err := xml2json.Convert(wantXML)
	if err != nil {
		t.Fatalf("failed to parse XML: %s", wantXML)
	}

	if !reflect.DeepEqual([]xml2json.Node{node}, wanted.Nodes()) {
		t.Errorf("XML mismatch:\n  got:\n%s\n  want:\n%s", nodeToXML(node), wanted.ToXML())
	}
}

// mustConvert calls Convert and fails the test if it returns an error,
// returning the node for the happy-path assertions.
func mustConvert(t *testing.T, name string, value any) xml2json.Node {
	t.Helper()
	node, err := AnnotationConverter(0).Convert(name, value)
	if err != nil {
		t.Fatalf("Convert(%q): unexpected error: %v", name, err)
	}
	return node
}

// ---- Convert: scalar values -------------------------------------------------

func TestConvert_StringValue_ProducesInlineAnnotation(t *testing.T) {
	// NewAttributes("Term", "@Core.Description", "String", "A book") →
	// sorted: String="A book" Term="@Core.Description"
	assertXML(t,
		mustConvert(t, "@Core.Description", "A book"),
		`<Annotation Term="Core.Description" String="A book" />`,
	)
}

func TestConvert_QualifiedAnnotation_SeparatesTermAndQualifier(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Core.Description#mobile", "A book"),
		`<Annotation Term="Core.Description" Qualifier="mobile" String="A book" />`,
	)
}

func TestConvert_BoolValue_True_ProducesInlineAnnotation(t *testing.T) {
	// NewAttributes("Term", "@Core.Computed", "Bool", "true") →
	// sorted: Bool="true" Term="@Core.Computed"
	assertXML(t,
		mustConvert(t, "@Core.Computed", true),
		`<Annotation Term="Core.Computed" Bool="true" />`,
	)
}

func TestConvert_BoolValue_False_ProducesInlineAnnotation(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Core.Computed", false),
		`<Annotation Term="Core.Computed" Bool="false" />`,
	)
}

func TestConvert_IntValue_ProducesInlineAnnotation(t *testing.T) {
	// NewAttributes("Term", "@MaxItems", "Int", "42") →
	// sorted: Int="42" Term="@MaxItems"
	assertXML(t,
		mustConvert(t, "@MaxItems", 42),
		`<Annotation Term="MaxItems" Int="42" />`,
	)
}

func TestConvert_FloatValue_ProducesInlineAnnotation(t *testing.T) {
	// NewAttributes("Term", "@Scale", "Float", "3.14") →
	// sorted: Float="3.14" Term="@Scale"
	assertXML(t,
		mustConvert(t, "@Scale", float64(3.14)),
		`<Annotation Term="Scale" Float="3.14" />`,
	)
}

// ---- Convert: term name stripping -------------------------------------------

func TestConvert_CollectionValue_TermAtPrefixStripped(t *testing.T) {
	node := mustConvert(t, "@Capabilities.BatchSupportType", []any{"Single"})
	if node.Attribute("Term") != "Capabilities.BatchSupportType" {
		t.Errorf("Term attribute: got %q, want %q", node.Attribute("Term"), "Capabilities.BatchSupportType")
	}
}

// ---- Convert: collection value ----------------------------------------------

func TestConvert_CollectionOfStrings_ProducesCollectionElement(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Capabilities.BatchSupportType", []any{"Single", "Transactional"}),
		`<Annotation Term="Capabilities.BatchSupportType">
  <Collection>
    <String>Single</String>
    <String>Transactional</String>
  </Collection>
</Annotation>`,
	)
}

func TestConvert_CollectionOfInts_ProducesCollectionElement(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Validation.AllowedValues", []any{1, 2, 3}),
		`<Annotation Term="Validation.AllowedValues">
  <Collection>
    <Int>1</Int>
    <Int>2</Int>
    <Int>3</Int>
  </Collection>
</Annotation>`,
	)
}

func TestConvert_EmptyCollection_ProducesSelfClosingCollectionElement(t *testing.T) {
	// An empty Collection has no children, so the serialiser emits a self-closing tag.
	assertXML(t,
		mustConvert(t, "@Core.Links", []any{}),
		`<Annotation Term="Core.Links">
  <Collection />
</Annotation>`,
	)
}

// ---- Convert: record value --------------------------------------------------

func TestConvert_RecordWithScalarBoolProperty_ProducesInlinePropertyValue(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Capabilities.InsertRestrictions", map[string]any{
			"Insertable": false,
		}),
		`<Annotation Term="Capabilities.InsertRestrictions">
  <Record>
    <PropertyValue Property="Insertable" Bool="false" />
  </Record>
</Annotation>`,
	)
}

func TestConvert_RecordWithStringProperty_ProducesInlinePropertyValue(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Core.OptionalParameter", map[string]any{
			"DefaultValue": "none",
		}),
		`<Annotation Term="Core.OptionalParameter">
  <Record>
    <PropertyValue Property="DefaultValue" String="none"/>
  </Record>
</Annotation>`,
	)
}

func TestConvert_RecordWithNestedRecord_ProducesNestedRecordElement(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Capabilities.InsertRestrictions", map[string]any{
			"QueryOptions": map[string]any{
				"ExpandSupported": true,
			},
		}),
		`<Annotation Term="Capabilities.InsertRestrictions">
  <Record>
    <PropertyValue Property="QueryOptions">
      <Record>
        <PropertyValue Property="ExpandSupported" Bool="true" />
      </Record>
    </PropertyValue>
  </Record>
</Annotation>`,
	)
}

func TestConvert_RecordWithCollectionProperty_ProducesCollectionInsideRecord(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Capabilities.InsertRestrictions", map[string]any{
			"CustomHeaders": []any{"X-Request-ID", "X-Tenant"},
		}),
		`<Annotation Term="Capabilities.InsertRestrictions">
  <Record>
    <PropertyValue Property="CustomHeaders">
      <Collection>
        <String>X-Request-ID</String>
        <String>X-Tenant</String>
      </Collection>
    </PropertyValue>
  </Record>
</Annotation>`,
	)
}

// ---- Convert: collection of records -----------------------------------------

func TestConvert_CollectionOfRecords_SingleRecord_ProducesRecordInsideCollection(t *testing.T) {
	// Single-property record avoids map-iteration order non-determinism.
	assertXML(t,
		mustConvert(t, "@Core.Links", []any{
			map[string]any{"rel": "author"},
		}),
		`<Annotation Term="Core.Links">
  <Collection>
    <Record>
      <PropertyValue Property="rel" String="author"/>
    </Record>
  </Collection>
</Annotation>`,
	)
}

// ---- resolveTypeName: error on unsupported type -----------------------------

func TestConvert_UnsupportedScalarType_ReturnsError(t *testing.T) {
	_, err := AnnotationConverter(0).Convert("@Core.Description", []byte("unsupported"))
	if err == nil {
		t.Error("expected an error for an unsupported value type, got nil")
	}
}

// ---- asCollectionElement: nested collection ---------------------------------

func TestConvert_CollectionContainingCollection_ProducesNestedCollections(t *testing.T) {
	assertXML(t,
		mustConvert(t, "@Core.Links", []any{
			[]any{"Single", "Transactional"},
		}),
		`<Annotation Term="Core.Links">
  <Collection>
    <Collection>
      <String>Single</String>
      <String>Transactional</String>
    </Collection>
  </Collection>
</Annotation>`,
	)
}
