//go:build unit || integration

package testutils

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/model"
)

func ApplyAndParse(t *testing.T, p interface {
	Apply(model.OverlayDefinition) (model.ResourceDefinition, error)
}, od model.OverlayDefinition) map[string]any {
	t.Helper()
	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	return UnmarshalResult[map[string]any](t, rd.MediaType, rd.Content)
}

// NoPatches is an OverlayDefinition with an empty patch list.
func NoPatches() model.OverlayDefinition {
	return model.OverlayDefinition{Overlay: model.Overlay{Patches: []model.Patch{}}}
}

// Get navigates a nested map[string]any by successive string keys, fataling on type mismatch.
func Get(t *testing.T, doc map[string]any, keys ...string) any {
	t.Helper()

	var cur any = doc
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("Get: expected map at key %q, got %T", k, cur)
		}
		cur = m[k]
	}
	return cur
}

// UnmarshalResult parses content with marshaller.Unmarshal and returns the result as map[string]any.
func UnmarshalResult[T any](t *testing.T, mediaType, content string) (result T) {
	t.Helper()

	parsed, err := marshaller.Unmarshal(mediaType, content)
	if err != nil {
		t.Fatalf("UnmarshalResult: %v", err)
	}

	if _, ok := parsed.(T); !ok {
		t.Fatalf("UnmarshalResult: expected %T, got %T", result, parsed)
	}

	return parsed.(T)
}

// FindByName scans a []any for the first map[string]any whose "name" field equals name.
func FindByName(t *testing.T, items []any, name string) map[string]any {
	t.Helper()
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if m["name"] == name {
			return m
		}
	}
	t.Fatalf("FindByName: no item named %q", name)
	return nil
}

// OnePatch builds a single-patch OverlayDefinition.
func OnePatch(action string, selector model.Selector, data any) model.OverlayDefinition {
	return model.OverlayDefinition{
		Overlay: model.Overlay{
			Patches: []model.Patch{
				{Action: action, Selector: &selector, Data: data},
			},
		},
	}
}
