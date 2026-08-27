package xml2json

import "github.com/open-resource-discovery/overlay-golang/internal/common/utils"

type Attributes map[string]string

func NewAttributes(values ...string) Attributes {
	return utils.AsMap(values...)
}

func (self Attributes) Has(name string) bool {
	return utils.ContainsKey(self, name)
}

func (self Attributes) Get(name string) string {
	return self[name]
}
