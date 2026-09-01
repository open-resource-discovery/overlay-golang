package overlays

import (
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/processor"
	"github.com/open-resource-discovery/overlay-golang/model"
)

func IsApplicable(definition model.ResourceDefinition, overlay model.OverlayDefinition) bool {
	if overlay.Overlay.Target == nil {
		return false
	}

	if len(overlay.Overlay.Target.URL) > 0 && definition.URL != overlay.Overlay.Target.URL {
		return false // not applicable, URL mismatch
	}

	if len(overlay.Overlay.Target.OrdID) > 0 && definition.OrdID != overlay.Overlay.Target.OrdID {
		return false // not applicable, OrdID mismatch
	}

	if len(overlay.Overlay.Perspective) > 0 && definition.Perspective != overlay.Overlay.Perspective {
		return false // not applicable, Perspective mismatch
	}

	if len(overlay.Overlay.Target.DefinitionType) > 0 && definition.DefinitionType != overlay.Overlay.Target.DefinitionType {
		return false // not applicable, DefinitionType mismatch
	}

	return true
}

func Apply(definition model.ResourceDefinition, overlays []model.OverlayDefinition) (results []model.ResourceDefinition, err *errors.OverlayError) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(errors.WrapPrefix(r, "unexpected error occurred"), errors.Severity_Fatal)
		}
	}()

	var aggregated *errors.OverlayError
	var proc = processor.MustCreateFor(definition)

	for _, overlay := range overlays {
		if !IsApplicable(definition, overlay) {
			continue
		}

		result, err := proc.Apply(overlay)

		results = append(results, result)
		aggregated = errors.Append(aggregated, err)
	}

	return results, aggregated
}
