package model

// Target identifies the resource or definition file being patched.
// At least one of OrdID, URL, CorrelationIDs, or DefinitionType must be set.
type Target struct {
	// OrdID is the ORD ID of the target resource (e.g. API Resource, Event
	// Resource, Data Product).
	OrdID string `json:"ordId,omitempty"`

	// URL points directly to the file being patched.
	URL string `json:"url,omitempty"`

	// CorrelationIDs reference the target resource in external registries.
	// Format: <namespace>:<type>:<localId>
	CorrelationIDs []string `json:"correlationIds,omitempty"`

	// DefinitionType is the type of the target definition being patched
	// (e.g. "openapi-v3", "asyncapi-v2", "edmx"). Should match the `type`
	// field of the referenced resource definition.
	DefinitionType string `json:"definitionType,omitempty"`

	// SystemInstance further scopes the target to a specific tenant when
	// Perspective is "system-instance".
	SystemInstance *SystemInstance `json:"systemInstance,omitempty"`
}
