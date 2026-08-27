package model

// SystemType represents the abstract type of an application or service from
// an operational perspective.
type SystemType struct {
	// SystemNamespace is the unique identifier for the system type
	// (e.g. "sap.s4"). Pattern: two dot-separated lowercase alphanumeric
	// segments, max 32 characters.
	SystemNamespace string `json:"systemNamespace,omitempty"`

	// CorrelationIDs link this system type to external systems of record.
	CorrelationIDs []string `json:"correlationIds,omitempty"`
}
