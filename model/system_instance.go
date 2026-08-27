package model

// SystemInstance represents a concrete, running instance of a system type
// (typically a tenant).
type SystemInstance struct {
	// BaseURL is the base URL of the system instance. Must not have a trailing
	// slash.
	BaseURL string `json:"baseUrl,omitempty"`

	// LocalID is the local ID for the instance as known by the described
	// system (usually a tenant ID).
	LocalID string `json:"localId,omitempty"`

	// CorrelationIDs link this instance to external systems of record.
	CorrelationIDs []string `json:"correlationIds,omitempty"`
}
