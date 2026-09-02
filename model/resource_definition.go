package model

import (
	"encoding/json"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

// ResourceDefinition represents an ORD API resource definition file
// (e.g. OpenAPI, AsyncAPI, OData CSDL, MCP/A2A Agent Card).
type ResourceDefinition struct {
	URL                     string
	OrdID                   string
	Purpose                 string
	Content                 string
	MediaType               string
	Visibility              string
	Perspective             string
	DefinitionType          string
	DescribedSystemType     *SystemType
	DescribedSystemVersion  *SystemVersion
	DescribedSystemInstance *SystemInstance
}

func (self ResourceDefinition) String() string {
	return string(utils.First(json.Marshal(self)))
}
