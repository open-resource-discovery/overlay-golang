//go:build unit

package edmx

import (
	"strings"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// minimalXML builds a minimal EDMX document with one EntityType and optional
// pre-existing Annotations blocks injected just before </Schema>.
func minimalXML(extraXML string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="Svc" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Book">
        <Property Name="title" Type="Edm.String"/>
      </EntityType>
      <EntityContainer Name="Container">
        <EntitySet Name="Books" EntityType="Svc.Book"/>
      </EntityContainer>` + extraXML + `
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`
}

func newProcessor(t *testing.T, xml string) *OverlayProcessor {
	t.Helper()

	return NewOverlayProcessor(model.ResourceDefinition{Content: xml})
}

// ─── NewOverlayProcessor ─────────────────────────────────────────────────────

func TestNewOverlayProcessor_ValidXML_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "unexpected error: %v")

	NewOverlayProcessor(model.ResourceDefinition{Content: minimalXML("")})
}

func TestNewOverlayProcessor_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid XML")

	NewOverlayProcessor(model.ResourceDefinition{Content: "not xml"})
}

func TestApply_EDMX_ODataV2_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for unsupported EDMX version")

	NewOverlayProcessor(model.ResourceDefinition{
		DefinitionType: "edmx",
		MediaType:      "application/xml",
		Content:        `<edmx:Edmx Version="1.0" />`,
	})
}

// ─── Apply — result shape ─────────────────────────────────────────────────────

func TestApply_SetsResultPurpose(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Purpose: "overlay-purpose",
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "A book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Purpose != "overlay-purpose" {
		t.Errorf("Purpose: got %q, want %q", result.Purpose, "overlay-purpose")
	}
}

func TestApply_SetsResultVisibility(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Visibility: "public",
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "A book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Visibility != "public" {
		t.Errorf("Visibility: got %q, want %q", result.Visibility, "public")
	}
}

func TestApply_OriginalContentUnmutated(t *testing.T) {
	xml := minimalXML("")
	p := newProcessor(t, xml)
	_, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "A book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Original definition content is unchanged
	if p.definition.Content != xml {
		t.Error("Apply mutated the original definition content")
	}
}

// ─── Apply — merge action ─────────────────────────────────────────────────────

func TestApply_Merge_NoExistingAnnotations_CreatesAnnotationsBlock(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "A book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, `Target="Svc.Book"`) {
		t.Errorf("expected Annotations Target=Svc.Book in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `Core.Description`) {
		t.Errorf("expected Core.Description annotation in output:\n%s", result.Content)
	}
}

func TestApply_Merge_ExplicitZeroArgumentOperationSelectsZeroParameterOverload(t *testing.T) {
	xml := minimalXML(`
      <Function Name="Find">
        <ReturnType Type="Svc.Book"/>
      </Function>
      <Function Name="Find">
        <Parameter Name="id" Type="Edm.String"/>
        <ReturnType Type="Svc.Book"/>
      </Function>`)
	p := newProcessor(t, xml)
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{Operation: "Svc.Find()"},
				Data:     map[string]any{"@Core.Description": "All books"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, `Target="Svc.Find()"`) {
		t.Errorf("expected zero-parameter operation target in output:\n%s", result.Content)
	}
}

func TestApply_Merge_OperationSignatureMatchesExactArity(t *testing.T) {
	xml := minimalXML(`
      <Function Name="Find">
        <Parameter Name="id" Type="Edm.String"/>
        <ReturnType Type="Svc.Book"/>
      </Function>
      <Function Name="Find">
        <Parameter Name="id" Type="Edm.String"/>
        <Parameter Name="edition" Type="Edm.Int32"/>
        <ReturnType Type="Svc.Book"/>
      </Function>`)
	p := newProcessor(t, xml)
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{Operation: "Svc.Find(Edm.String)"},
				Data:     map[string]any{"@Core.Description": "One book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, `Target="Svc.Find(Edm.String)"`) {
		t.Errorf("expected exact one-parameter operation target in output:\n%s", result.Content)
	}
}

func TestApply_Merge_ExplicitZeroArgumentSignatureDoesNotMatchFunctionImport(t *testing.T) {
	xml := strings.Replace(
		minimalXML(""),
		"</EntityContainer>",
		`<FunctionImport Name="Ping" Function="Svc.Ping"/></EntityContainer>`,
		1,
	)
	p := newProcessor(t, xml)
	_, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{Operation: "Svc.Ping()"},
				Data:     map[string]any{"@Core.Description": "Ping"},
			}},
		},
	})
	if err == nil {
		t.Fatal("expected signature-qualified selector not to match FunctionImport")
	}
}

func TestApply_Merge_ExistingAnnotations_AppendsToBlock(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.LongDescription" String="existing"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "A book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, "Core.LongDescription") {
		t.Errorf("expected existing Core.LongDescription preserved in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Core.Description") {
		t.Errorf("expected new Core.Description appended in output:\n%s", result.Content)
	}
}

func TestApply_Merge_AnnotationAlreadyExists_OverwritesExistingValue(t *testing.T) {
	// Document has: one unqualified + two qualified annotations on Core.Description.
	// Patch targets the unqualified one and the "q1" qualifier.
	// The "q2" qualifier must remain untouched.
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="unqualified-old"/>
        <Annotation Term="Core.Description" Qualifier="q1" String="q1-old"/>
        <Annotation Term="Core.Description" Qualifier="q2" String="q2-unchanged"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data: map[string]any{
					"@Core.Description":    "unqualified-new",
					"@Core.Description#q1": "q1-new",
				},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Exactly one annotation per distinct Term+Qualifier combination.
	if count := strings.Count(result.Content, `Term="Core.Description"`); count != 3 {
		t.Errorf("expected exactly 3 Core.Description annotations, got %d:\n%s", count, result.Content)
	}
	// Patched values are present.
	if !strings.Contains(result.Content, `String="unqualified-new"`) {
		t.Errorf("expected unqualified annotation updated to 'unqualified-new':\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `String="q1-new"`) {
		t.Errorf("expected q1 annotation updated to 'q1-new':\n%s", result.Content)
	}
	// Old values are gone.
	if strings.Contains(result.Content, `String="unqualified-old"`) {
		t.Errorf("expected old unqualified value replaced:\n%s", result.Content)
	}
	if strings.Contains(result.Content, `String="q1-old"`) {
		t.Errorf("expected old q1 value replaced:\n%s", result.Content)
	}
	// Untouched qualifier is unchanged.
	if !strings.Contains(result.Content, `String="q2-unchanged"`) {
		t.Errorf("expected q2 annotation left untouched:\n%s", result.Content)
	}
}

// ─── Apply — update action ────────────────────────────────────────────────────

func TestApply_Update_NoExistingAnnotations_CreatesAnnotationsBlock(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "update",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "A book"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, `Target="Svc.Book"`) {
		t.Errorf("expected Annotations Target=Svc.Book in output:\n%s", result.Content)
	}
}

func TestApply_Update_ExistingAnnotations_ReplacesAnnotations(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.LongDescription" String="old"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "update",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "new"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, "Core.Description") {
		t.Errorf("expected Core.Description in output:\n%s", result.Content)
	}
	if strings.Contains(result.Content, "Core.LongDescription") {
		t.Errorf("expected Core.LongDescription to be replaced, but it's still in output:\n%s", result.Content)
	}
}

// ─── Apply — remove action ────────────────────────────────────────────────────

func TestApply_Remove_ExistingAnnotations_PrunesMatchingTerms(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="keep me not"/>
        <Annotation Term="Core.LongDescription" String="keep me"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "remove",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": nil},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Content, "Core.Description") {
		t.Errorf("expected Core.Description removed from output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Core.LongDescription") {
		t.Errorf("expected Core.LongDescription preserved in output:\n%s", result.Content)
	}
}

func TestApply_Remove_NoExistingAnnotations_IsNoOp(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "remove",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": nil},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Content, "Core.Description") {
		t.Errorf("unexpected Core.Description in output:\n%s", result.Content)
	}
}

func TestApply_Remove_NilData_RemovesAllAnnotations(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="a"/>
        <Annotation Term="Core.LongDescription" String="b"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "remove",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     nil,
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Content, "Core.Description") {
		t.Errorf("expected Core.Description removed from output:\n%s", result.Content)
	}
	if strings.Contains(result.Content, "Core.LongDescription") {
		t.Errorf("expected Core.LongDescription removed from output:\n%s", result.Content)
	}
}

func TestApply_Remove_NilData_NoExistingAnnotations_IsNoOp(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "remove",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     nil,
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output must be a valid EDMX document — the entity type must still be present
	if !strings.Contains(result.Content, `Name="Book"`) {
		t.Errorf("EntityType Book missing from output:\n%s", result.Content)
	}
}

// ─── Apply — reconcile: pre-existing inline Annotation nodes ─────────────────

func TestApply_Reconcile_InlineAnnotations_MergedBeforeProcessing(t *testing.T) {
	// EntityType has an inline <Annotation> child — reconcile should strip and re-merge it.
	xml := `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="Svc" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Book">
        <Annotation Term="Core.LongDescription" String="inline"/>
        <Property Name="title" Type="Edm.String"/>
      </EntityType>
      <EntityContainer Name="Container">
        <EntitySet Name="Books" EntityType="Svc.Book"/>
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`
	p := newProcessor(t, xml)
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "new"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The inline annotation must appear in an Annotations block (not inline on EntityType)
	if !strings.Contains(result.Content, `Target="Svc.Book"`) {
		t.Errorf("expected Annotations block with Target=Svc.Book:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Core.LongDescription") {
		t.Errorf("expected reconciled Core.LongDescription in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, "Core.Description") {
		t.Errorf("expected merged Core.Description in output:\n%s", result.Content)
	}
}

// ─── Apply — qualifier support ────────────────────────────────────────────────

func TestApply_Merge_QualifiedAnnotation_WritesQualifierAttribute(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description#Restricted": "Restricted books only"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, `Term="Core.Description"`) {
		t.Errorf("expected Term=Core.Description in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `Qualifier="Restricted"`) {
		t.Errorf("expected Qualifier=Restricted in output:\n%s", result.Content)
	}
}

func TestApply_Merge_QualifiedAnnotation_DoesNotOverwriteUnqualifiedAnnotation(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="unqualified"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description#Restricted": "qualified"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, `String="unqualified"`) {
		t.Errorf("expected unqualified annotation preserved in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `Qualifier="Restricted"`) {
		t.Errorf("expected qualified annotation added in output:\n%s", result.Content)
	}
}

func TestApply_Update_QualifiedAnnotation_ReplacesQualifiedAnnotation(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" Qualifier="Restricted" String="old"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "update",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description#Restricted": "new"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Content, `String="old"`) {
		t.Errorf("expected old qualified annotation replaced in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `String="new"`) {
		t.Errorf("expected new qualified annotation value in output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `Qualifier="Restricted"`) {
		t.Errorf("expected Qualifier=Restricted in output:\n%s", result.Content)
	}
}

func TestApply_Remove_QualifiedAnnotation_PrunesOnlyMatchingQualifier(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="unqualified"/>
        <Annotation Term="Core.Description" Qualifier="Restricted" String="qualified"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "remove",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description#Restricted": nil},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Content, `Qualifier="Restricted"`) {
		t.Errorf("expected qualified annotation removed from output:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `String="unqualified"`) {
		t.Errorf("expected unqualified annotation preserved in output:\n%s", result.Content)
	}
}

// ─── Apply — error cases ──────────────────────────────────────────────────────

func TestApply_UnsupportedAction_ReturnsError(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	if _, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "replace",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description": "x"},
			}},
		},
	}); err == nil {
		t.Fatal("expected error for unsupported action, got nil")
	}
}

func TestApply_BadSelector_ReturnsError(t *testing.T) {
	p := newProcessor(t, minimalXML(""))
	if _, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.NonExistent"},
				Data:     map[string]any{"@Core.Description": "x"},
			}},
		},
	}); err == nil {
		t.Fatal("expected error for missing selector target, got nil")
	}
}
