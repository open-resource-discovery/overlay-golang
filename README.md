[![REUSE status](https://api.reuse.software/badge/github.com/open-resource-discovery/overlay-golang)](https://api.reuse.software/info/github.com/open-resource-discovery/overlay-golang)
[![CI](https://img.shields.io/github/actions/workflow/status/open-resource-discovery/overlay-golang/ci.yml?label=CI)](https://github.com/open-resource-discovery/overlay-golang/actions/workflows/ci.yml)

# ORD Overlay Golang Library

## About this project

Library for applying [Open Resource Discovery Overlay](https://open-resource-discovery.org/spec-v1/interfaces/OrdOverlay) patches to API resource definition files. It lets you enrich or adapt OpenAPI, OData CSDL/EDMX, CSN, A2A Agent Card, and generic JSON/YAML documents without modifying the original sources.

## Requirements and Setup

Requires Go 1.26 or later.

### Installation

```sh
go get github.tools.sap/ORD/goverlay
```

### Quick start

```go
import (
    overlays "github.tools.sap/ORD/goverlay"
    "github.tools.sap/ORD/goverlay/model"
)

definition := model.ResourceDefinition{
    OrdID:          "sap.example:apiResource:CustomerAPI:v1",
    DefinitionType: "openapi-v3",
    MediaType:      "application/json",
    Content:        `{ ... }`, // raw document content
}

patches := []model.OverlayDefinition{
    {
        Overlay: model.Overlay{
            OrdOverlay: "0.1",
            Target: &model.Target{OrdID: "sap.example:apiResource:CustomerAPI:v1"},
            Patches: []model.Patch{
                {
                    Action:   "merge",
                    Selector: &model.Selector{Operation: "getCustomer"},
                    Data: map[string]any{
                        "x-sap-visibility": "public",
                    },
                },
            },
        },
    },
}

results, err := overlays.Apply(definition, patches)
if err != nil {
    log.Fatal(err)
}
// results[i].Content holds the patched document for each applicable overlay
```

`Apply` returns one `model.ResourceDefinition` per overlay whose `Target` matched the definition. Each result's `Content` field holds the patched document as a string.

#### Checking applicability without applying

```go
if overlays.IsApplicable(definition, overlayDef) {
    // overlay targets this definition
}
```

An overlay is applicable when every non-empty field of its `Target` (`OrdID`, `URL`, `DefinitionType`, `Perspective`) matches the corresponding field on the definition.

### Supported document types

The processor is selected automatically from `ResourceDefinition.DefinitionType`, with `MediaType` as fallback:

| DefinitionType | Format | Notes |
|---|---|---|
| `openapi-v2`, `openapi-v3`, `openapi-v3.1+` | JSON or YAML | Structural + operation-level selectors |
| `csdl-json` | JSON | OData CSDL JSON — entity type, complex type, enum, entity set, operation, namespace selectors |
| `edmx` | XML | OData EDMX — annotation-focused; entity type, complex type, enum, entity set, operation selectors |
| `sap-csn-interop-effective-v1` | JSON | SAP CSN Interop — entity type and definition selectors |
| `a2a-agent-card` | JSON | A2A Agent Card — skill (operation) selector |
| *(any)* | `application/json` | Generic JSON — root and JSONPath selectors |
| *(any)* | `application/yaml` | Generic YAML — root and JSONPath selectors |

### Overlay model

#### `ResourceDefinition`

Represents the API resource definition file to be patched.

```go
type ResourceDefinition struct {
    OrdID          string
    URL            string
    DefinitionType string  // selects the processor; see table above
    MediaType      string  // fallback processor selection
    Content        string  // raw document text (JSON, YAML, or XML)
    Perspective    string  // "system-type" | "system-version" | "system-instance"
    Visibility     string
    Purpose        string
    DescribedSystemType     *SystemType
    DescribedSystemVersion  *SystemVersion
    DescribedSystemInstance *SystemInstance
}
```

#### `OverlayDefinition`

Wraps an `Overlay` (a `Target` and a list of `Patch` operations).

```go
type OverlayDefinition struct {
    Overlay Overlay
}

type Overlay struct {
    OrdOverlay  string   // overlay spec version, e.g. "0.1"
    Target      *Target
    Perspective string
    Patches     []Patch
}
```

#### `Target`

Identifies which resource definition to patch. At least one field must be set.

```go
type Target struct {
    OrdID          string   // ORD ID of the target resource
    URL            string   // direct URL to the definition file
    CorrelationIDs []string // external registry references
    DefinitionType string   // e.g. "openapi-v3", "edmx"
    SystemInstance *SystemInstance
}
```

#### `Patch`

A single patch operation applied to the element identified by `Selector`.

```go
type Patch struct {
    Action      string    // "merge" | "update" | "remove"
    Selector    *Selector
    Data        any       // payload for merge/update; nil-keyed map for targeted remove
    Description string    // informational only
    Tags        []string
    Meta        map[string]json.RawMessage
}
```

#### Actions

| Action | Behaviour |
|---|---|
| `merge` | Deep-merges `Data` into the selected node. Map keys are merged recursively; arrays are appended. Creates the node if it does not exist. |
| `update` | Fully replaces the selected node with `Data`. |
| `remove` | Deletes the selected node. When `Data` is a `map[string]any` with `nil` values, only those specific keys are deleted from the node rather than the node itself. |

#### `Selector`

Exactly one field should be set per patch. Concept-level selectors are preferred over `JSONPath` — they are resilient to format changes and document restructuring.

| Field | Targets |
|---|---|
| `Root *bool` | The root of the document (`true` required) |
| `JSONPath string` | Any node addressed by a JSONPath expression (must start with `$`) |
| `Operation string` | An operation by ID: `operationId` (OpenAPI), `skills[].id` (A2A Agent Card), namespace-qualified name (OData) |
| `EntityType string` | OData EntityType by namespace-qualified name, e.g. `"ODataDemo.Product"` |
| `ComplexType string` | OData ComplexType by namespace-qualified name |
| `EnumType string` | OData EnumType by namespace-qualified name |
| `PropertyType string` | Property or enum member by unqualified name — must be combined with `EntityType`, `ComplexType`, or `EnumType` |
| `EntitySet string` | OData EntitySet inside the EntityContainer |
| `Namespace string` | OData schema/namespace element |
| `Operation` + `Parameter string` | A named parameter on the identified operation |
| `Operation` + `ReturnType *bool` | The return type of the identified operation |

> **Note:** `Root` and `JSONPath` selectors are not supported for the `edmx` processor. Use concept-level selectors (EntityType, Operation, etc.) instead.

### Running the tests

Run vet:

```sh
go vet -tags unit,integration ./...
```

Run tests:

```sh
# Run unit tests only
go test -tags unit ./...

# Run integration tests only
go test -tags integration ./...

# Run unit and integration tests together
go vet -tags unit,integration ./...
```

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/open-resource-discovery/overlay-golang/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](https://github.com/open-resource-discovery/.github/blob/main/CONTRIBUTING.md).

## Security / Disclosure
If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/open-resource-discovery/.github/blob/main/SECURITY.md) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/open-resource-discovery/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2026 SAP SE or an SAP affiliate company and overlay-golang contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/open-resource-discovery/overlay-golang).
