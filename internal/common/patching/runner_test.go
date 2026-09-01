//go:build unit

package patching

import (
	"errors"
	"strings"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/model"
)

func TestRunKeepsFailedPatchAtomicAndContinues(t *testing.T) {
	content := map[string]int{"initial": 1}
	patches := []model.Patch{
		{Description: "fails after mutating its candidate"},
		{Description: "succeeds"},
	}
	var diagnostics []model.Diagnostic

	result, err := Run(
		content,
		patches,
		func(value map[string]int) map[string]int {
			copy := make(map[string]int, len(value))
			for key, item := range value {
				copy[key] = item
			}
			return copy
		},
		nil,
		func(patch model.Patch, candidate map[string]int) (map[string]int, error) {
			if patch.Description == "fails after mutating its candidate" {
				candidate["discarded"] = 1
				return candidate, errors.New("invalid patch")
			}
			candidate["applied"] = 1
			return candidate, nil
		},
		func(diagnostic model.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
	)

	if err == nil || !strings.Contains(err.Error(), "patch #1: invalid patch") {
		t.Fatalf("got error %v, want aggregate error for patch #1", err)
	}
	if _, found := result["discarded"]; found {
		t.Fatal("failed patch mutation was retained")
	}
	if result["applied"] != 1 {
		t.Fatal("valid patch after the error was not applied")
	}
	if len(diagnostics) != 1 || diagnostics[0].Severity != model.DiagnosticSeverityError {
		t.Fatalf("got diagnostics %#v, want one error", diagnostics)
	}
}

func TestRunWarnsAndSkipsUnmatchedJSONPath(t *testing.T) {
	content := map[string]any{"existing": true}
	patches := []model.Patch{{
		Action:   "merge",
		Selector: &model.Selector{JSONPath: "$.missing"},
		Data:     map[string]any{"created": true},
	}}
	applyCalled := false
	var diagnostics []model.Diagnostic

	result, err := Run(
		content,
		patches,
		func(value map[string]any) map[string]any { return value },
		CheckJSONPathMatch[map[string]any],
		func(_ model.Patch, candidate map[string]any) (map[string]any, error) {
			applyCalled = true
			return candidate, nil
		},
		func(diagnostic model.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
	)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if applyCalled {
		t.Fatal("unmatched JSONPath was applied")
	}
	if result["existing"] != true {
		t.Fatal("content changed")
	}
	if len(diagnostics) != 1 || diagnostics[0].Severity != model.DiagnosticSeverityWarning {
		t.Fatalf("got diagnostics %#v, want one warning", diagnostics)
	}
}
