package processor

import (
	"strings"

	"github.com/go-errors/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/a2aagentcard"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/csdl"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/csn"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/edmx"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/json"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/openapi"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/yaml"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type OverlayProcessor interface {
	// Apply applies the given Patch to this resource definition and returns
	// the patched definition or an error if the patch could not be applied.
	Apply(overlay model.OverlayDefinition) (model.ResourceDefinition, error)

	// ApplyWithDiagnostics applies the given Patch to this resource definition and returns
	// the patched definition, diagnostics, and any aggregate error.
	ApplyWithDiagnostics(overlay model.OverlayDefinition, handlers ...model.DiagnosticHandler) (model.ResourceDefinition, error)
}

func CreateFor(definition model.ResourceDefinition) (OverlayProcessor, error) {
	switch definition.DefinitionType {
	case "edmx":
		return edmx.NewOverlayProcessor(definition)
	case "csdl-json":
		return csdl.NewOverlayProcessor(definition)
	case "a2a-agent-card":
		return a2aagentcard.NewOverlayProcessor(definition)
	case "sap-csn-interop-effective-v1":
		return csn.NewOverlayProcessor(definition)
	case "openapi-v2", "openapi-v3", "openapi-v3.1+":
		return openapi.NewOverlayProcessor(definition)
	default:
		if strings.HasPrefix(definition.MediaType, "application/json") {
			return json.NewOverlayProcessor(definition)
		}

		if strings.HasPrefix(definition.MediaType, "text/yaml") || strings.HasPrefix(definition.MediaType, "application/yaml") {
			return yaml.NewOverlayProcessor(definition)
		}

		return nil, errors.Errorf("unsupported resource definition: %s (%s)", definition.DefinitionType, definition.MediaType)
	}
}
