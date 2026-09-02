package model

import (
	"encoding/json"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

// SystemVersion describes a specific released version of a system type.
type SystemVersion struct {
	// Version follows Semantic Versioning 2.0.0 (e.g. "1.2.3", "2024.8.0").
	Version string `json:"version,omitempty"`

	// Title is a human-readable title for the system version
	// (e.g. "SAP S/4HANA Cloud 2408").
	Title string `json:"title,omitempty"`

	// CorrelationIDs link this system version to external systems of record.
	CorrelationIDs []string `json:"correlationIds,omitempty"`
}

func (self SystemVersion) String() string {
	return string(utils.First(json.Marshal(self)))
}
