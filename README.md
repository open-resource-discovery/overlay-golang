[![REUSE status](https://api.reuse.software/badge/github.com/open-resource-discovery/overlay-golang)](https://api.reuse.software/info/github.com/open-resource-discovery/overlay-golang)
[![CI](https://github.com/open-resource-discovery/overlay-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/open-resource-discovery/overlay-golang/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fopen-resource-discovery%2Foverlay-golang%2Freleases%2Flatest&query=%24.tag_name&label=Latest%20Release)](https://github.com/open-resource-discovery/overlay-golang/releases/latest)


# ORD Overlay Golang Library

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
It attempts every patch and applicable overlay, applies each patch atomically, and returns successfully produced partial results together with one aggregate error when any patch fails.
Pass an optional `model.DiagnosticHandler` to receive structured warnings and errors with overlay and patch locations.

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
| `merge` | Deep-merges `Data` into selected nodes. Map keys are merged recursively; arrays are appended. |
| `update` | Fully replaces the selected node with `Data`. |
| `remove` | Deletes the selected node. When `Data` is a `map[string]any` with `nil` values, only those specific keys are deleted from the node rather than the node itself. An unmatched selector is a successful no-op. |

#### `Selector`

Exactly one field should be set per patch.
Concept-level selectors are preferred over `JSONPath` because they are resilient to format changes and document restructuring.

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

#### Limitations and behavior notes

These reflect what the ORD Overlay spec currently defines and how this library implements it.
They are documented here so callers do not rely on unspecified behavior.

- **No-match / create-on-missing.** No selector creates a missing target.
  A concept-level `merge` or `update` whose selector matches nothing returns an error.
  A zero-match `JSONPath` patch is a successful no-op and produces a warning, regardless of action, because JSONPath selectors naturally match zero or more elements.
  An unmatched `remove` is also a successful no-op and produces a warning because the requested absence already holds.
  Do not rely on create-on-missing; a future version may reintroduce it behind a dedicated create action.
- **Portable JSONPath profile.** JSONPath follows RFC 9535.
  Every toolkit supports root, dot and quoted-bracket member selectors, non-negative array indices, object and array wildcards, homogeneous member or index selector lists, and forward array slices with omitted or non-negative bounds.
  Portable overlays use only this subset.
  Toolkits may accept additional non-portable extensions, but unsupported expressions are errors rather than zero matches.
- **Error aggregation.** Every patch is attempted, failed patches do not retain partial mutations, and later patches and overlays continue.
  `Apply` returns successfully produced partial results and one aggregate error when any error occurred.
  Consumers choose how to display diagnostics and whether warnings should affect their own process exit status.
- **The document root cannot be removed.** An omitted-data `remove` using `root` or `JSONPath: "$"` returns an error.
  Use `update` when the complete document must be replaced.
- **EDMX targets are annotation-only.** For the `edmx` (OData XML) processor, overlays add, replace, and remove annotations on existing structure.
  They cannot create new structural elements (EntityType, EntitySet, Property, Function/Action, EnumType, ComplexType); those must already exist in the source schema.
  Combined with the note above (no `Root`/`JSONPath` for `edmx`), there is no structural-authoring path.
  New structure belongs in the source CSDL/CDS, not in an overlay.
- **CSDL JSON enum members** are scalars, so member annotations are written as sibling keys on the enum type (`Read@Core.Description`), not merged into the member value.

These behaviors are specified by the [ORD Overlay specification clarification](https://github.com/open-resource-discovery/specification/pull/171).

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
