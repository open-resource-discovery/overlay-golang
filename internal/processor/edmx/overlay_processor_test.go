//go:build unit

package edmx

import (
	"strings"
	"testing"

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
	p, err := NewOverlayProcessor(model.ResourceDefinition{Content: xml})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	return p
}

// ─── NewOverlayProcessor ─────────────────────────────────────────────────────

func TestNewOverlayProcessor_ValidXML_ReturnsProcessor(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{Content: minimalXML("")}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewOverlayProcessor_InvalidContent_ReturnsError(t *testing.T) {
	if _, err := NewOverlayProcessor(model.ResourceDefinition{Content: "not xml"}); err == nil {
		t.Fatal("expected error for invalid XML, got nil")
	}
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

func TestApply_Merge_ExistingSameTerm_ReplacesInsteadOfDuplicating(t *testing.T) {
	// Merging a Term that already exists in the target block must replace it,
	// not append a second <Annotation Term="..."> (a single-valued term twice
	// is invalid OData).
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="old"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
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
	if n := strings.Count(result.Content, `Term="Core.Description"`); n != 1 {
		t.Errorf("expected exactly one Core.Description annotation, got %d:\n%s", n, result.Content)
	}
	if !strings.Contains(result.Content, `String="new"`) {
		t.Errorf("expected merged value String=\"new\" in output:\n%s", result.Content)
	}
	if strings.Contains(result.Content, `String="old"`) {
		t.Errorf("expected old value to be replaced, but String=\"old\" remains:\n%s", result.Content)
	}
}

func TestApply_Merge_SameTermDifferentQualifier_PreservesBoth(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="base"/>
        <Annotation Term="Core.Description" Qualifier="mobile" String="old"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "merge",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description#mobile": "new"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := strings.Count(result.Content, `Term="Core.Description"`); n != 2 {
		t.Errorf("expected unqualified and qualified annotations, got %d:\n%s", n, result.Content)
	}
	if !strings.Contains(result.Content, `Qualifier="mobile" String="new"`) {
		t.Errorf("expected qualified annotation to be replaced:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `String="base"`) {
		t.Errorf("expected unqualified annotation to be preserved:\n%s", result.Content)
	}
}

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

func TestApply_Remove_QualifiedAnnotation_PreservesOtherQualifiers(t *testing.T) {
	existing := `
      <Annotations Target="Svc.Book">
        <Annotation Term="Core.Description" String="base"/>
        <Annotation Term="Core.Description" Qualifier="mobile" String="mobile"/>
      </Annotations>`
	p := newProcessor(t, minimalXML(existing))
	result, err := p.Apply(model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{{
				Action:   "remove",
				Selector: &model.Selector{EntityType: "Svc.Book"},
				Data:     map[string]any{"@Core.Description#mobile": nil},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Content, `Qualifier="mobile"`) {
		t.Errorf("expected qualified annotation removed:\n%s", result.Content)
	}
	if !strings.Contains(result.Content, `Term="Core.Description"`) ||
		!strings.Contains(result.Content, `String="base"`) {
		t.Errorf("expected unqualified annotation preserved:\n%s", result.Content)
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
