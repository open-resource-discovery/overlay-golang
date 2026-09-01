package patching

import (
	"fmt"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type ApplyPatch[T any] func(model.Patch, T) (T, error)
type Clone[T any] func(T) T
type Preflight[T any] func(T, model.Patch) (bool, error)

func Run[T any](
	content T,
	patches []model.Patch,
	clone Clone[T],
	preflight Preflight[T],
	apply ApplyPatch[T],
	handlers ...model.DiagnosticHandler,
) (T, error) {
	var failures []string

	for patchIndex, patch := range patches {
		candidate := clone(content)
		if preflight != nil {
			matched, err := preflight(candidate, patch)
			if err != nil {
				failures = appendFailure(failures, handlers, patchIndex, err)
				continue
			}
			if !matched {
				emit(handlers, model.Diagnostic{
					Severity:   model.DiagnosticSeverityWarning,
					PatchIndex: patchIndex,
					Message:    "selector did not match any target element",
				})
				continue
			}
		}

		next, err := apply(patch, candidate)
		if err != nil {
			if patch.Action == "remove" && strings.HasPrefix(err.Error(), "no such element:") {
				emit(handlers, model.Diagnostic{
					Severity:   model.DiagnosticSeverityWarning,
					PatchIndex: patchIndex,
					Message:    err.Error(),
				})
				continue
			}
			failures = appendFailure(failures, handlers, patchIndex, err)
			continue
		}
		content = next
	}

	if len(failures) > 0 {
		return content, fmt.Errorf("overlay merge failed:\n%s", strings.Join(failures, "\n"))
	}

	return content, nil
}

func CheckJSONPathMatch[T any](content T, patch model.Patch) (bool, error) {
	if patch.Selector == nil || patch.Selector.JSONPath == "" {
		return true, nil
	}

	expression, err := jp.ParseString(patch.Selector.JSONPath)
	if err != nil {
		return false, err
	}
	return expression.Has(content), nil
}

func appendFailure(
	failures []string,
	handlers []model.DiagnosticHandler,
	patchIndex int,
	err error,
) []string {
	message := fmt.Sprintf("patch #%d: %v", patchIndex+1, err)
	emit(handlers, model.Diagnostic{
		Severity:   model.DiagnosticSeverityError,
		PatchIndex: patchIndex,
		Message:    err.Error(),
	})
	return append(failures, "- "+message)
}

func emit(handlers []model.DiagnosticHandler, diagnostic model.Diagnostic) {
	for _, handler := range handlers {
		if handler != nil {
			handler(diagnostic)
		}
	}
}
