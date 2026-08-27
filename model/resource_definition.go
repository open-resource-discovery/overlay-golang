package model

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
