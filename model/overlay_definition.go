package model

import (
	"encoding/json"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

type OverlayDefinition struct {
	Purpose string

	Overlay Overlay
}

func (self OverlayDefinition) String() string {
	return string(utils.First(json.Marshal(self)))
}
