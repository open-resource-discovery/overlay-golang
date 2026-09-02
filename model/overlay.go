package model

import (
	"encoding/json"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

// Overlay is the root object of an ORD Overlay document.
// It describes an ordered set of patches to apply to a referenced resource
// definition file (e.g. OpenAPI, AsyncAPI, OData CSDL) without modifying
// the original source.
type Overlay struct {
	// Schema is the optional URI to the ORD Overlay JSON Schema, enabling
	// editor validation and code intelligence.
	Schema string `json:"$schema,omitempty"`

	// OrdOverlay is the version of the ORD Overlay specification. Currently
	// the only valid value is "0.1".
	OrdOverlay string `json:"ordOverlay"`

	// OrdID is the optional ORD ID of this overlay document.
	// Pattern: <namespace>:overlay:<localId>:<version>
	OrdID string `json:"ordId,omitempty"`

	// Description is an optional CommonMark (Markdown) description of the
	// overlay document itself.
	Description string `json:"description,omitempty"`

	// Perspective scopes where the overlay should be applied: one of
	// "system-type", "system-version", or "system-instance".
	Perspective string `json:"perspective,omitempty"`

	// DescribedSystemType links the overlay to a system type in the ORD
	// system landscape model.
	DescribedSystemType *SystemType `json:"describedSystemType,omitempty"`

	// DescribedSystemVersion links the overlay to a specific released system
	// version.
	DescribedSystemVersion *SystemVersion `json:"describedSystemVersion,omitempty"`

	// DescribedSystemInstance links the overlay to a specific system instance
	// (tenant).
	DescribedSystemInstance *SystemInstance `json:"describedSystemInstance,omitempty"`

	// Visibility controls which consumers can discover this overlay document.
	// One of "public", "internal", or "private".
	Visibility string `json:"visibility,omitempty"`

	// Target identifies the resource or definition file being patched.
	Target *Target `json:"target,omitempty"`

	// Patches is the ordered sequence of patches to apply. Required; at least
	// one patch must be present. Patches are applied in listed order.
	Patches []Patch `json:"patches"`

	// Meta holds arbitrary out-of-band metadata about the overlay document.
	// Never applied to the target document.
	Meta map[string]json.RawMessage `json:"meta,omitempty"`

	Url string
}

func (self Overlay) String() string {
	return string(utils.First(json.Marshal(self)))
}
