package model

// Selector identifies the element in the target document to patch.
// Exactly one field must be set per patch.
//
// Prefer concept-level selectors (Operation, EntityType, etc.) over
// JSONPath where possible — they are resilient to format changes.
type Selector struct {
	// Root targets the root of the target document. Must be true when set.
	// Equivalent to JSONPath "$" but preferred for clarity.
	Root *bool `json:"root,omitempty"`

	// JSONPath is a generic structural fallback selector. Must start with "$".
	JSONPath string `json:"jsonPath,omitempty"`

	// Operation targets a concept-level operation by its identifier:
	//   - OpenAPI: operationId
	//   - MCP: tools[].name
	//   - A2A Agent Card: skills[].id
	//   - OData: namespace-qualified Action or Function name
	Operation string `json:"operation,omitempty"`

	// EntityType targets an OData EntityType by its namespace-qualified name
	// (e.g. "OData.Demo.Customer"). Also used for CSN Interop definitions.
	EntityType string `json:"entityType,omitempty"`

	// ComplexType targets an OData ComplexType by its namespace-qualified
	// name (e.g. "OData.Demo.Address").
	ComplexType string `json:"complexType,omitempty"`

	// EnumType targets an OData EnumType by its namespace-qualified name
	// (e.g. "OData.Demo.OrderStatus").
	EnumType string `json:"enumType,omitempty"`

	// PropertyType targets a property, navigation property, or enum member
	// by its unqualified name. Must be combined with exactly one of
	// EntityType, ComplexType, or EnumType as parent context.
	PropertyType string `json:"propertyType,omitempty"`

	// EntitySet targets an OData EntitySet inside the EntityContainer
	// (e.g. "Customers"). Use for Capabilities annotations.
	EntitySet string `json:"entitySet,omitempty"`

	// Namespace targets an OData schema/namespace element by its fully
	// qualified namespace value (e.g. "com.example.OrderService").
	Namespace string `json:"namespace,omitempty"`

	// Parameter targets a parameter by name on an operation. Operation must
	// also be set to identify the owning operation.
	Parameter string `json:"parameter,omitempty"`

	// ReturnType targets the return type of the operation identified by
	// Operation. Must be true when set.
	ReturnType *bool `json:"returnType,omitempty"`
}
