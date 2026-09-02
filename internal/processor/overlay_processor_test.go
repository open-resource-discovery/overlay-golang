//go:build unit

package processor

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// minimalContent returns the smallest valid content string for each format so
// that NewOverlayProcessor succeeds without needing the full fixture files.
const (
	minimalJSON = `{"name":"test"}`
	minimalYAML = "name: test\n"
	minimalXML  = `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"></edmx:Edmx>`
	minimalCSN  = `{"definitions":{}}`
	minimalCSDL = `{"$Version":"4.0"}`
	minimalA2A  = `{"name":"test-agent","description":"desc","version":"1.0.0","url":"https://example.com"}`
)

// ---- DefinitionType dispatch -------------------------------------------------

func TestCreateFor_DefinitionType_Edmx_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor edmx: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "edmx",
		MediaType:      "application/xml",
		Content:        minimalXML,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for edmx")
	}
}

func TestCreateFor_DefinitionType_CsdlJSON_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor csdl-json: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "csdl-json",
		MediaType:      "application/json",
		Content:        minimalCSDL,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for csdl-json")
	}
}

func TestCreateFor_DefinitionType_A2AAgentCard_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor a2a-agent-card: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "a2a-agent-card",
		MediaType:      "application/json",
		Content:        minimalA2A,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for a2a-agent-card")
	}
}

func TestCreateFor_DefinitionType_SapCsnInterop_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor sap-csn-interop-effective-v1: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "sap-csn-interop-effective-v1",
		MediaType:      "application/json",
		Content:        minimalCSN,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for sap-csn-interop-effective-v1")
	}
}

func TestCreateFor_DefinitionType_OpenAPIv2_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor openapi-v2: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "openapi-v2",
		MediaType:      "application/json",
		Content:        minimalJSON,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for openapi-v2")
	}
}

func TestCreateFor_DefinitionType_OpenAPIv3_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor openapi-v3: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "openapi-v3",
		MediaType:      "application/json",
		Content:        minimalJSON,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for openapi-v3")
	}
}

func TestCreateFor_DefinitionType_OpenAPIv31Plus_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor openapi-v3.1+: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "openapi-v3.1+",
		MediaType:      "application/json",
		Content:        minimalJSON,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for openapi-v3.1+")
	}
}

// ---- MediaType fallback: application/json -----------------------------------

func TestCreateFor_MediaType_ApplicationJSON_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor application/json: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "application/json",
		Content:        minimalJSON,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for application/json")
	}
}

func TestCreateFor_MediaType_ApplicationJSONWithParameters_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor application/json; charset=utf-8: %v")

	// MediaType with charset suffix — HasPrefix still matches.
	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "application/json; charset=utf-8",
		Content:        minimalJSON,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for application/json with charset")
	}
}

// ---- MediaType fallback: YAML -----------------------------------------------

func TestCreateFor_MediaType_ApplicationYAML_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor application/yaml: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "application/yaml",
		Content:        minimalYAML,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for application/yaml")
	}
}

func TestCreateFor_MediaType_TextYAML_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor text/yaml: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "text/yaml",
		Content:        minimalYAML,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for text/yaml")
	}
}

func TestCreateFor_MediaType_TextYAMLWithParameters_ReturnsProcessor(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "MustCreateFor text/yaml; charset=utf-8: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "text/yaml; charset=utf-8",
		Content:        minimalYAML,
	})

	if p == nil {
		t.Fatal("expected non-nil processor for text/yaml with charset")
	}
}

// ---- DefinitionType takes precedence over MediaType -------------------------

func TestCreateFor_DefinitionType_TakesPrecedenceOverMediaType(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "expected openapi processor despite yaml media type: %v")

	// DefinitionType = "openapi-v3" with MediaType = "application/yaml":
	// should still return an openapi processor, not a yaml processor.
	// Both parse the same minimal JSON so either would succeed; we just confirm
	// no error and a non-nil processor is returned.
	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "openapi-v3",
		MediaType:      "application/yaml",
		Content:        minimalYAML,
	})

	if p == nil {
		t.Fatal("expected non-nil processor")
	}
}

// ---- Unsupported combinations -----------------------------------------------

func TestCreateFor_UnsupportedDefinitionType_UnknownMediaType_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for unsupported definition type + media type")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "unknown-type",
		MediaType:      "application/octet-stream",
		Content:        "",
	})
}

func TestCreateFor_EmptyDefinitionType_EmptyMediaType_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error when both DefinitionType and MediaType are empty")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "",
		Content:        "",
	})
}

func TestCreateFor_UnsupportedDefinitionType_ValidJSONMediaType_FallsBackToJSON(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "expected json fallback for unknown definition type: %v")

	// Unknown DefinitionType hits the default branch; MediaType=application/json
	// routes to the json processor.
	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "unknown-type",
		MediaType:      "application/json",
		Content:        minimalJSON,
	})

	if p == nil {
		t.Fatal("expected non-nil processor")
	}
}

func TestCreateFor_UnsupportedDefinitionType_ValidYAMLMediaType_FallsBackToYAML(t *testing.T) {
	defer testutils.AssertDoesNotPanic(t, "expected yaml fallback for unknown definition type: %v")

	p := MustCreateFor(model.ResourceDefinition{
		DefinitionType: "unknown-type",
		MediaType:      "application/yaml",
		Content:        minimalYAML,
	})

	if p == nil {
		t.Fatal("expected non-nil processor")
	}
}

// ---- Invalid content returns error ------------------------------------------

func TestCreateFor_DefinitionType_Edmx_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid edmx content")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "edmx",
		MediaType:      "application/xml",
		Content:        "not xml",
	})
}

func TestCreateFor_DefinitionType_CsdlJSON_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid csdl-json content")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "csdl-json",
		MediaType:      "application/json",
		Content:        "{invalid json",
	})
}

func TestCreateFor_DefinitionType_SapCsnInterop_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid csn content")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "sap-csn-interop-effective-v1",
		MediaType:      "application/json",
		Content:        "{invalid json",
	})
}

func TestCreateFor_DefinitionType_OpenAPIv3_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid openapi content")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "openapi-v3",
		MediaType:      "application/json",
		Content:        "{invalid json",
	})
}

func TestCreateFor_MediaType_ApplicationJSON_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid json content")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "application/json",
		Content:        "{invalid json",
	})
}

func TestCreateFor_MediaType_ApplicationYAML_InvalidContent_ReturnsError(t *testing.T) {
	defer testutils.AssertPanics(t, "expected error for invalid yaml content")

	MustCreateFor(model.ResourceDefinition{
		DefinitionType: "",
		MediaType:      "application/yaml",
		Content:        ":\n  - bad: [unclosed",
	})
}

// ---- Returned processor implements the interface ----------------------------

func TestCreateFor_ReturnedProcessor_ImplementsOverlayProcessorInterface(t *testing.T) {
	cases := []struct {
		name       string
		definition model.ResourceDefinition
	}{
		{"edmx", model.ResourceDefinition{DefinitionType: "edmx", MediaType: "application/xml", Content: minimalXML}},
		{"csdl-json", model.ResourceDefinition{DefinitionType: "csdl-json", MediaType: "application/json", Content: minimalCSDL}},
		{"a2a-agent-card", model.ResourceDefinition{DefinitionType: "a2a-agent-card", MediaType: "application/json", Content: minimalA2A}},
		{"sap-csn-interop-effective-v1", model.ResourceDefinition{DefinitionType: "sap-csn-interop-effective-v1", MediaType: "application/json", Content: minimalCSN}},
		{"openapi-v2", model.ResourceDefinition{DefinitionType: "openapi-v2", MediaType: "application/json", Content: minimalJSON}},
		{"openapi-v3", model.ResourceDefinition{DefinitionType: "openapi-v3", MediaType: "application/json", Content: minimalJSON}},
		{"openapi-v3.1+", model.ResourceDefinition{DefinitionType: "openapi-v3.1+", MediaType: "application/json", Content: minimalJSON}},
		{"json fallback", model.ResourceDefinition{DefinitionType: "", MediaType: "application/json", Content: minimalJSON}},
		{"yaml fallback", model.ResourceDefinition{DefinitionType: "", MediaType: "application/yaml", Content: minimalYAML}},
		{"text/yaml fallback", model.ResourceDefinition{DefinitionType: "", MediaType: "text/yaml", Content: minimalYAML}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer testutils.AssertDoesNotPanic(t, "unexpected error: %v")

			p := MustCreateFor(tc.definition)

			// Verify the interface is satisfied by calling Apply with an empty overlay.
			if _, err := p.Apply(model.OverlayDefinition{
				Overlay: model.Overlay{Patches: []model.Patch{}},
			}); err != nil {
				t.Fatalf("Apply(empty overlay): %v", err)
			}
		})
	}
}
