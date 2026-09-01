//go:build integration

package yaml

import (
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// integrationInput is the shared MCP server card fixture for all integration tests.
var integrationInput = testutils.LoadFixture("testdata/mcp_server_card.yaml")

// applyIntegration applies the given overlay to the integration input fixture
// and returns the parsed result document.
func applyIntegration(t *testing.T, od model.OverlayDefinition) map[string]any {
	t.Helper()
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   integrationInput,
		MediaType: "application/yaml",
	})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}
	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	parsed, err := marshaller.Unmarshal("application/yaml", rd.Content)
	if err != nil {
		t.Fatalf("Unmarshal result: %v", err)
	}
	return parsed.(map[string]any)
}

// loadIntegrationExpected loads and parses an expected-output YAML fixture.
func loadIntegrationExpected(path string) map[string]any {
	return testutils.UnmarshalFixture[map[string]any]("testdata/integration/" + path)
}

// ---- merge: root selector ---------------------------------------------------

// TestIntegration_Merge_Root_AddsStateInfo merges x-sap-stateInfo into the root
// document, verifying that deep-merge preserves all existing top-level keys while
// adding the new extension field.
func TestIntegration_Merge_Root_AddsStateInfo(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("merge_root_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"x-sap-stateInfo": map[string]any{
					"state":           "Active",
					"deprecationDate": "2026-12-31",
				},
			},
		)),
	)
}

// ---- update: root selector --------------------------------------------------

// TestIntegration_Update_Root_ReplacesEntireDocument replaces the whole document
// with a minimal stub, verifying nothing from the original survives.
func TestIntegration_Update_Root_ReplacesEntireDocument(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("update_root_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{Root: utils.Ptr(true)},
			map[string]any{
				"name":    "replaced",
				"version": "0.0.1",
			},
		)),
	)
}

// ---- remove: root selector --------------------------------------------------

func TestIntegration_Remove_Root_ReturnsError(t *testing.T) {
	selectors := map[string]model.Selector{
		"root selector": {Root: utils.Ptr(true)},
		"root JSONPath": {JSONPath: "$"},
	}
	for name, selector := range selectors {
		t.Run(name, func(t *testing.T) {
			p := testutils.AssertNoError(NewOverlayProcessor(model.ResourceDefinition{
				Content: integrationInput, MediaType: "application/yaml",
			}))
			_, err := p.Apply(testutils.OnePatch("remove", selector, nil))
			if err == nil {
				t.Fatal("expected root remove to return an error")
			}
		})
	}
}

// ---- merge: JSONPath selector -----------------------------------------------

// TestIntegration_Merge_JSONPath_BumpsVersionAndAddsOrdId merges a new version
// and an x-sap-ord-id extension into the root via JSONPath "$", verifying that
// existing fields (name, title, description, tools, etc.) are preserved.
func TestIntegration_Merge_JSONPath_BumpsVersionAndAddsOrdId(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("merge_jsonpath_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{JSONPath: "$"},
			map[string]any{
				"version":      "2.0.0",
				"x-sap-ord-id": "sap.bookshop:mcpServerCard:catalogservice:v2",
			},
		)),
	)
}

// ---- update: JSONPath selector ----------------------------------------------

// TestIntegration_Update_JSONPath_ReplacesMeta replaces the _meta block entirely
// with an updated set of SAP-specific metadata keys, verifying the original
// "sap/category: TBD" entry is gone and new keys are present.
func TestIntegration_Update_JSONPath_ReplacesMeta(t *testing.T) {
	result := applyIntegration(t, testutils.OnePatch(
		"update",
		model.Selector{JSONPath: "$._meta"},
		map[string]any{
			"sap/category": "AI Tools",
			"sap/owner":    "team-ai-services",
		},
	))

	testutils.AssertDeepEquals(t, result, loadIntegrationExpected("update_jsonpath_expected.yaml"))
}

// ---- remove: JSONPath selector ----------------------------------------------

// TestIntegration_Remove_JSONPath_RemovesPrompts removes the prompts array from
// the document, verifying all other top-level keys are untouched.
func TestIntegration_Remove_JSONPath_RemovesPrompts(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("remove_jsonpath_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.prompts"},
			nil,
		)),
	)
}

// ---- merge: JSONPath targeting a nested tool --------------------------------

// TestIntegration_Merge_JSONPath_UpdatesToolDescription updates the description
// of the first tool (catalogservice-books-read) and marks it deprecated via a
// custom extension, verifying that name, inputSchema, and _meta are preserved.
func TestIntegration_Merge_JSONPath_UpdatesToolDescription(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("merge_jsonpath_tool_description_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{JSONPath: "$.tools[0]"},
			map[string]any{
				"description":  "Reads a book from the CatalogService. Deprecated: use v2 endpoint instead.",
				"x-deprecated": true,
			},
		)),
	)
}

// ---- update: JSONPath targeting a tool inputSchema --------------------------

// TestIntegration_Update_JSONPath_ReplacesToolInputSchema replaces the inputSchema
// of the books-create tool with a trimmed definition (ID + title only), verifying
// that descr, stock, author, texts, and localized are gone after the update.
func TestIntegration_Update_JSONPath_ReplacesToolInputSchema(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("update_jsonpath_tool_inputschema_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{JSONPath: "$.tools[1].inputSchema"},
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"ID":    map[string]any{"type": "integer", "description": "Unique identifier of the book"},
					"title": map[string]any{"type": "string", "maxLength": 111, "description": "Title of the book"},
				},
				"required":             []any{"title"},
				"additionalProperties": false,
				"$schema":              "http://json-schema.org/draft-07/schema#",
			},
		)),
	)
}

// ---- remove: JSONPath targeting the resources array -------------------------

// TestIntegration_Remove_JSONPath_RemovesResources removes the resources array
// entirely, verifying tools, prompts, and all other top-level keys are untouched.
func TestIntegration_Remove_JSONPath_RemovesResources(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("remove_jsonpath_resources_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"remove",
			model.Selector{JSONPath: "$.resources"},
			nil,
		)),
	)
}

// ---- merge: JSONPath targeting a nested resource ----------------------------

// TestIntegration_Merge_JSONPath_EnrichesResource enriches the first resource entry
// with a more detailed description and an x-license extension, verifying that
// uri, name, and mimeType are preserved.
func TestIntegration_Merge_JSONPath_EnrichesResource(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("merge_jsonpath_resource_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"merge",
			model.Selector{JSONPath: "$.resources[0]"},
			map[string]any{
				"description": "The official SAP logo in SVG format, suitable for embedding in documents.",
				"x-license":   "SAP Brand Guidelines",
			},
		)),
	)
}

// ---- update: JSONPath targeting remotes -------------------------------------

// TestIntegration_Update_JSONPath_ReplacesRemoteEndpoint replaces the first remote
// entry with a production endpoint definition, verifying the original example URL
// is gone after the update.
func TestIntegration_Update_JSONPath_ReplacesRemoteEndpoint(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("update_jsonpath_remotes_expected.yaml"),
		applyIntegration(t, testutils.OnePatch(
			"update",
			model.Selector{JSONPath: "$.remotes[0]"},
			map[string]any{
				"type": "streamable-http",
				"url":  "https://prod.example.com/mcp/v2",
			},
		)),
	)
}

// ---- multi-patch: realistic overlay sequence --------------------------------

// TestIntegration_MultiPatch_RealisticOverlaySequence applies a sequence of
// patches that mirrors a realistic MCP server card overlay:
//  1. merge $ → bump version to 1.1.0, add x-sap-ord-id
//  2. update $._meta → replace category and add owner
//  3. remove $.prompts → strip prompts
//  4. merge root → add x-sap-stateInfo
func TestIntegration_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	testutils.AssertDeepEquals(
		t,
		loadIntegrationExpected("multi_patch_expected.yaml"),
		applyIntegration(t, model.OverlayDefinition{
			Overlay: model.Overlay{Patches: []model.Patch{
				{
					Action:   "merge",
					Selector: &model.Selector{JSONPath: "$"},
					Data: map[string]any{
						"version":      "1.1.0",
						"x-sap-ord-id": "sap.bookshop:mcpServerCard:catalogservice:v1",
					},
				},
				{
					Action:   "update",
					Selector: &model.Selector{JSONPath: "$._meta"},
					Data: map[string]any{
						"sap/category": "AI Tools",
						"sap/owner":    "team-ai-services",
					},
				},
				{
					Action:   "remove",
					Selector: &model.Selector{JSONPath: "$.prompts"},
					Data:     nil,
				},
				{
					Action:   "merge",
					Selector: &model.Selector{Root: utils.Ptr(true)},
					Data: map[string]any{
						"x-sap-stateInfo": map[string]any{
							"state": "Active",
						},
					},
				},
			}},
		}),
	)
}
