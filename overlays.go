package overlays

import (
	"github.com/go-errors/errors"
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

func Apply(definition model.ResourceDefinition, overlays []model.OverlayDefinition) (results []model.ResourceDefinition, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.WrapPrefix(r, "unexpected error occurred", 0)
		}
	}()

	if proc, err := processor.CreateFor(definition); err != nil {
		return nil, errors.WrapPrefix(err, "failed to process resource definition", 0)
	} else {
		for _, overlay := range overlays {
			if !IsApplicable(definition, overlay) {
				continue
			}

			if result, err := proc.Apply(overlay); err != nil {
				return nil, errors.WrapPrefix(err, "failed to apply overlay", 0)
			} else {
				results = append(results, result)
			}
		}
	}

	return results, nil
}
