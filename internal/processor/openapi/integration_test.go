//go:build integration

package openapi

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/testutils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

// applyIntegration applies the given overlay to the integration input fixture
// and returns the parsed result document.
func applyIntegration(t *testing.T, fixture string, od model.OverlayDefinition) map[string]any {
	t.Helper()
	p, err := NewOverlayProcessor(model.ResourceDefinition{
		Content:   testutils.LoadFixture(fixture),
		MediaType: fmt.Sprintf("application/%s", strings.Split(filepath.Base(fixture), ".")[1]),
	})
	if err != nil {
		t.Fatalf("NewOverlayProcessor: %v", err)
	}

	rd, err := p.Apply(od)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	parsed, err := marshaller.Unmarshal(rd.MediaType, rd.Content)
	if err != nil {
		t.Fatalf("Unmarshal result: %v", err)
	}

	return parsed.(map[string]any)
}

// loadIntegrationExpected loads and parses an expected-output fixture.
func loadIntegrationExpected(path string) map[string]any {
	return testutils.UnmarshalFixture[map[string]any]("testdata/integration/" + path)
}

// ---- merge: root selector ---------------------------------------------------

// TestIntegration_Merge_Root_AddsStateInfo merges x-sap-stateInfo into the root
// document, verifying that deep-merge preserves all existing top-level keys while
// adding the new extension field.
func TestIntegration_Merge_Root_AddsStateInfo(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				testutils.UnmarshalFixture[map[string]any](fmt.Sprintf("testdata/integration/merge_root_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
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
		})
	}
}

// ---- update: root selector --------------------------------------------------

// TestIntegration_Update_Root_ReplacesEntireDocument replaces the whole document
// with a minimal OpenAPI stub, verifying nothing from the original survives.
func TestIntegration_Update_Root_ReplacesEntireDocument(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("update_root_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"update",
					model.Selector{Root: utils.Ptr(true)},
					map[string]any{
						"openapi": "3.1.0",
						"info":    map[string]any{"title": "Replaced", "version": "0.0.1"},
					},
				)),
			)
		})
	}
}

// ---- remove: root selector --------------------------------------------------

func TestIntegration_Remove_Root_ReturnsError(t *testing.T) {
	selectors := map[string]model.Selector{
		"root selector": {Root: utils.Ptr(true)},
		"root JSONPath": {JSONPath: "$"},
	}
	for _, format := range []string{"json", "yaml"} {
		for name, selector := range selectors {
			t.Run(format+"/"+name, func(t *testing.T) {
				p := testutils.AssertNoError(NewOverlayProcessor(model.ResourceDefinition{
					Content:   testutils.LoadFixture(fmt.Sprintf("testdata/petstore.%s", format)),
					MediaType: fmt.Sprintf("application/%s", format),
				}))
				_, err := p.Apply(testutils.OnePatch("remove", selector, nil))
				if err == nil {
					t.Fatal("expected root remove to return an error")
				}
			})
		}
	}
}

// ---- merge: JSONPath selector -----------------------------------------------

// TestIntegration_Merge_JSONPath_UpdatesInfoVersion bumps the API version and
// adds an x-sap-ord-id extension to $.info, verifying that existing fields
// (title, description, contact) are preserved while the version is overwritten.
func TestIntegration_Merge_JSONPath_UpdatesInfoVersion(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("merge_jsonpath_info_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"merge",
					model.Selector{JSONPath: "$.info"},
					map[string]any{
						"version":      "2.0.0",
						"x-sap-ord-id": "sap.petstore:apiResource:petstore:v2",
					},
				)),
			)
		})
	}
}

// ---- update: JSONPath selector ($.info) -------------------------------------

// TestIntegration_Update_JSONPath_ReplacesInfoEntirely replaces $.info with a
// trimmed object (title + version only), verifying that description and contact
// are gone after an update.
func TestIntegration_Update_JSONPath_ReplacesInfoEntirely(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("update_jsonpath_info_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"update",
					model.Selector{JSONPath: "$.info"},
					map[string]any{
						"title":   "Petstore",
						"version": "2.0.0",
					},
				)),
			)
		})
	}
}

// ---- update: JSONPath selector ($.servers) ----------------------------------

// TestIntegration_Update_JSONPath_ReplacesServers replaces the servers array
// with a single production-only entry, verifying the sandbox server is gone.
func TestIntegration_Update_JSONPath_ReplacesServers(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("update_jsonpath_servers_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"update",
					model.Selector{JSONPath: "$.servers"},
					[]any{
						map[string]any{
							"url":         "https://api.petstore.example/v2",
							"description": "Production v2",
						},
					},
				)),
			)
		})
	}
}

// ---- remove: JSONPath selector ----------------------------------------------

// TestIntegration_Remove_JSONPath_RemovesExternalDocs removes the externalDocs
// object from the document root, verifying other top-level keys are untouched.
func TestIntegration_Remove_JSONPath_RemovesExternalDocs(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("remove_jsonpath_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"remove",
					model.Selector{JSONPath: "$.externalDocs"},
					nil,
				)),
			)
		})
	}
}

// ---- merge: Operation selector ----------------------------------------------

// TestIntegration_Merge_Operation_UpdatesListPetsDescription updates the
// description of the listPets operation and adds an x-deprecated extension,
// verifying that operationId, tags, parameters, and responses are preserved.
func TestIntegration_Merge_Operation_UpdatesListPetsDescription(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("merge_operation_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"merge",
					model.Selector{Operation: "listPets"},
					map[string]any{
						"description":  "Returns a paginated list of all pets. Deprecated: use /v2/pets instead.",
						"x-deprecated": true,
					},
				)),
			)
		})
	}
}

// ---- update: Operation selector ---------------------------------------------

// TestIntegration_Update_Operation_ReplacesCreatePets replaces the createPets
// operation entirely with an updated definition that adds a requestBody and
// removes the 400 response, verifying full replacement semantics.
func TestIntegration_Update_Operation_ReplacesCreatePets(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("update_operation_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"update",
					model.Selector{Operation: "createPets"},
					map[string]any{
						"operationId": "createPets",
						"summary":     "Create a pet",
						"description": "Adds a new pet. Requires a valid request body.",
						"tags":        []any{"Pets"},
						"requestBody": map[string]any{
							"required": true,
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"$ref": "#/components/schemas/Pet"},
								},
							},
						},
						"responses": map[string]any{
							"201": map[string]any{"description": "Pet created successfully."},
						},
					},
				)),
			)
		})
	}
}

// ---- remove: Operation selector ---------------------------------------------

// TestIntegration_Remove_Operation_RemovesCreatePets removes the POST /pets
// operation entirely, verifying GET /pets and all /pets/{petId} operations survive.
func TestIntegration_Remove_Operation_RemovesCreatePets(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("remove_operation_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"remove",
					model.Selector{Operation: "createPets"},
					nil,
				)),
			)
		})
	}
}

// ---- merge: Operation + Parameter selector ----------------------------------

// TestIntegration_Merge_OperationWithParameter_EnrichsPetIdParam adds a schema
// example and updates the description on the petId path parameter of showPetById,
// verifying that other parameters and operation fields are unaffected.
func TestIntegration_Merge_OperationWithParameter_EnrichsPetIdParam(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("merge_operation_parameter_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"merge",
					model.Selector{Operation: "showPetById", Parameter: "petId"},
					map[string]any{
						"description": "The unique identifier of the pet. Must be a non-empty string.",
						"example":     "pet-42",
					},
				)),
			)
		})
	}
}

// ---- merge: JSONPath targeting a response -----------------------------------

// TestIntegration_Merge_JSONPath_UpdatesListPets200Response updates the
// description of the listPets 200 response via a JSONPath selector, verifying
// that content and other responses are preserved.
func TestIntegration_Merge_JSONPath_UpdatesListPets200Response(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("merge_jsonpath_response_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), testutils.OnePatch(
					"merge",
					model.Selector{JSONPath: "$.paths['/pets'].get.responses['200']"},
					map[string]any{
						"description": "A paginated list of pets. See x-next header for continuation.",
						"headers": map[string]any{
							"x-next": map[string]any{
								"description": "URL to the next page of results.",
								"schema":      map[string]any{"type": "string"},
							},
						},
					},
				)),
			)
		})
	}
}

// ---- multi-patch: realistic overlay sequence --------------------------------

// TestIntegration_MultiPatch_RealisticOverlaySequence applies a sequence of
// patches that mirrors a realistic API overlay:
//  1. merge $.info → bump version to 1.1.0, add x-sap-ord-id
//  2. update $.servers → replace with single production server
//  3. remove $.externalDocs → strip external documentation link
//  4. merge Operation "listPets" → add deprecation notice and tag
//  5. merge Operation "showPetById" + Parameter "petId" → add example
func TestIntegration_MultiPatch_RealisticOverlaySequence(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			testutils.AssertDeepEquals(
				t,
				loadIntegrationExpected(fmt.Sprintf("multi_patch_expected.%s", format)),
				applyIntegration(t, fmt.Sprintf("testdata/petstore.%s", format), model.OverlayDefinition{
					Overlay: model.Overlay{Patches: []model.Patch{
						{
							Action:   "merge",
							Selector: &model.Selector{JSONPath: "$.info"},
							Data: map[string]any{
								"version":      "1.1.0",
								"x-sap-ord-id": "sap.petstore:apiResource:petstore:v1",
							},
						},
						{
							Action:   "update",
							Selector: &model.Selector{JSONPath: "$.servers"},
							Data: []any{
								map[string]any{
									"url":         "https://api.petstore.example/v1",
									"description": "Production",
								},
							},
						},
						{
							Action:   "remove",
							Selector: &model.Selector{JSONPath: "$.externalDocs"},
							Data:     nil,
						},
						{
							Action:   "merge",
							Selector: &model.Selector{Operation: "listPets"},
							Data: map[string]any{
								"description":  "Returns a paginated list of all pets. Deprecated: use /v2/pets instead.",
								"x-deprecated": true,
							},
						},
						{
							Action:   "merge",
							Selector: &model.Selector{Operation: "showPetById", Parameter: "petId"},
							Data: map[string]any{
								"description": "The unique identifier of the pet. Must be a non-empty string.",
								"example":     "pet-42",
							},
						},
					}},
				}),
			)
		})
	}
}
