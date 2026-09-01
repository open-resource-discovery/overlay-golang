package model

import (
	"encoding/json"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

// Patch is a single patch action applied to the element identified by Selector.
type Patch struct {
	// Description is an optional human-readable note explaining the purpose
	// of this patch. Purely informational; has no effect on application.
	Description string `json:"description,omitempty"`

	// Action is the patch operation to perform. One of "update", "remove",
	// or "merge". Required.
	Action string `json:"action"`

	// Selector identifies the element in the target to patch. Required.
	Selector *Selector `json:"selector"`

	// Data is the value used by the patch action. Required when Action is
	// "update" or "merge".
	Data any `json:"data,omitempty"`

	// Tags are string labels associated with the patched element for
	// classification and filtering. Informational only.
	Tags []string `json:"tags,omitempty"`

	// Meta holds arbitrary out-of-band metadata about this individual patch.
	// Never applied to the target document.
	Meta map[string]json.RawMessage `json:"meta,omitempty"`
}

func (self Patch) String() string {
	return string(utils.First(json.Marshal(self)))
}
