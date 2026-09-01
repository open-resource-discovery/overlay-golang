package overlays

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/model"
)

type jsonPathPortabilityCase struct {
	Name     string      `json:"name"`
	Source   any         `json:"source"`
	Patch    model.Patch `json:"patch"`
	Expected any         `json:"expected"`
	Warnings int         `json:"warnings"`
}

func TestJSONPathPortableSubset(t *testing.T) {
	content, err := os.ReadFile("testdata/jsonpath_portable_cases.json")
	if err != nil {
		t.Fatal(err)
	}

	var cases []jsonPathPortabilityCase
	if err := json.Unmarshal(content, &cases); err != nil {
		t.Fatal(err)
	}

	for _, example := range cases {
		t.Run(example.Name, func(t *testing.T) {
			source, err := json.Marshal(example.Source)
			if err != nil {
				t.Fatal(err)
			}
			overlay := model.OverlayDefinition{Overlay: model.Overlay{
				OrdOverlay: "0.1",
				Target:     &model.Target{},
				Patches:    []model.Patch{example.Patch},
			}}
			var diagnostics []model.Diagnostic
			results, err := Apply(
				model.ResourceDefinition{Content: string(source), MediaType: "application/json"},
				[]model.OverlayDefinition{overlay},
				func(diagnostic model.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
			)
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("got %d results, want 1", len(results))
			}

			var result any
			if err := json.Unmarshal([]byte(results[0].Content), &result); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result, example.Expected) {
				t.Errorf("result mismatch\ngot:  %#v\nwant: %#v", result, example.Expected)
			}

			warningCount := 0
			for _, diagnostic := range diagnostics {
				if diagnostic.Severity == model.DiagnosticSeverityWarning {
					warningCount++
				}
			}
			if warningCount != example.Warnings {
				t.Errorf("got %d warnings, want %d", warningCount, example.Warnings)
			}
		})
	}
}
