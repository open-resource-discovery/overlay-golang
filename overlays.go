package overlays

import (
	"fmt"
	"strings"

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

func Apply(definition model.ResourceDefinition, overlays []model.OverlayDefinition, handlers ...model.DiagnosticHandler) (results []model.ResourceDefinition, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.WrapPrefix(r, "unexpected error occurred", 0)
		}
	}()

	if proc, err := processor.CreateFor(definition); err != nil {
		return nil, errors.WrapPrefix(err, "failed to process resource definition", 0)
	} else {
		var failures []string
		for overlayIndex, overlay := range overlays {
			if !IsApplicable(definition, overlay) {
				continue
			}

			forward := func(diagnostic model.Diagnostic) {
				diagnostic.OverlayIndex = overlayIndex
				for _, handler := range handlers {
					if handler != nil {
						handler(diagnostic)
					}
				}
			}
			result, applyErr := proc.ApplyWithDiagnostics(overlay, forward)
			if result.Content != "" {
				results = append(results, result)
			}
			if applyErr != nil {
				failures = append(failures, fmt.Sprintf("- overlay #%d: %v", overlayIndex+1, applyErr))
			}
		}
		if len(failures) > 0 {
			return results, fmt.Errorf("failed to apply overlays:\n%s", strings.Join(failures, "\n"))
		}
	}

	return results, nil
}
