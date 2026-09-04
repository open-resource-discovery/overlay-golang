package edmx

import (
	"fmt"
	"strings"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

type AnnotationConverter byte

func (self AnnotationConverter) Convert(name string, value any) xml2json.Node {
	return xml2json.NewElementNode(
		"Annotation",
		self.asChildren(value),
		self.asAttributes(
			[]string{"Term", utils.First(strings.Cut(name[1:], "#"))},
			[]string{"Qualifier", utils.Second(strings.Cut(name[1:], "#"))},
			[]string{self.resolveTypeName(value), fmt.Sprint(value)},
		),
	)
}

func (self AnnotationConverter) resolveTypeName(value any) string {
	switch value.(type) {
	case bool:
		return "Bool"
	case string:
		return "String"
	case float32, float64:
		return "Float"
	case nil, []any, map[string]any:
		return ""
	case int, int8, int16, int32, int64:
		return "Int"
	default:
		panic(fmt.Sprintf("unsupported value type: %T", value))
	}
}

func (self AnnotationConverter) asScalar(value any) xml2json.Node {
	return xml2json.NewElementNode(
		utils.Ternary(
			value == nil,
			"Null",
			self.resolveTypeName(value),
		),
		utils.Ternary(
			value == nil,
			[]xml2json.Node{},
			[]xml2json.Node{xml2json.NewTextNode(fmt.Sprint(value))},
		),
		xml2json.NewAttributes(),
	)
}

func (self AnnotationConverter) isStructuralPath(value string) bool {
	return utils.OneOf(value, "$Path", "$EnumMember", "$PropertyPath", "$AnnotationPath", "$ModelElementPath", "$NavigationPropertyPath")
}

func (self AnnotationConverter) isConstantWrapper(value string) bool {
	return utils.OneOf(value, "$Int", "$Decimal", "$Float", "$Date", "$DateTimeOffset", "$Duration", "$Guid", "$TimeOfDay", "$Binary", "$Null")
}

func (self AnnotationConverter) asChildren(value any) []xml2json.Node {
	switch value.(type) {
	case []any:
		return []xml2json.Node{self.asCollection(value.([]any))}
	case map[string]any:
		return utils.Flatten(
			[][]xml2json.Node{
				self.asRecords(self.extractRecordProperties(value.(map[string]any))),
				self.asStructuralPaths(self.extractStructuralPaths(value.(map[string]any))),
				self.asConstantWrappers(self.extractConstantWrappers(value.(map[string]any))),
			},
		)
	default:
		return utils.Ternary(
			len(self.resolveTypeName(value)) != 0,
			[]xml2json.Node{},
			[]xml2json.Node{self.asScalar(value)},
		)
	}
}

func (self AnnotationConverter) asCollection(values []any) xml2json.Node {
	return xml2json.NewElementNode(
		"Collection",
		utils.Flatten(
			utils.Map(values, func(_ int, value any) []xml2json.Node {
				switch value.(type) {
				case []any:
					return []xml2json.Node{self.asCollection(value.([]any))}
				case map[string]any:
					return utils.Flatten(
						[][]xml2json.Node{
							self.asRecords(self.extractRecordProperties(value.(map[string]any))),
							self.asStructuralPaths(self.extractStructuralPaths(value.(map[string]any))),
							self.asConstantWrappers(self.extractConstantWrappers(value.(map[string]any))),
						},
					)
				default:
					return []xml2json.Node{self.asScalar(value)}
				}
			}),
		),
		xml2json.NewAttributes(),
	)
}

func (self AnnotationConverter) asRecords(data map[string]any) []xml2json.Node {
	if len(data) == 0 {
		return []xml2json.Node{}
	}

	return []xml2json.Node{
		xml2json.NewElementNode(
			"Record",
			utils.Map(
				utils.Filter(
					utils.Keys(data),
					func(key string) bool { return key != "$Type" },
				),
				func(_ int, key string) xml2json.Node {
					return self.asPropertyValue(key, data[key])
				},
			),
			xml2json.NewAttributes(
				utils.Ternary(
					!utils.ContainsKey(data, "$Type"),
					[]string{},
					[]string{"Type", fmt.Sprint(data["$Type"])},
				)...,
			),
		),
	}
}

func (self AnnotationConverter) asAttributes(parts ...[]string) xml2json.Attributes {
	return xml2json.NewAttributes(
		utils.Flatten(
			utils.Filter(
				parts,
				func(parts []string) bool {
					return len(parts[0]) > 0 && len(parts[1]) > 0
				},
			),
		)...,
	)
}

func (self AnnotationConverter) asPropertyValue(key string, value any) xml2json.Node {
	return xml2json.NewElementNode(
		"PropertyValue",
		self.asChildren(value),
		self.asAttributes(
			[]string{"Property", key},
			[]string{self.resolveTypeName(value), fmt.Sprint(value)},
		),
	)
}

func (self AnnotationConverter) asStructuralPaths(data map[string]any) []xml2json.Node {
	// first add nodes for all structured properties & constant wrappers
	return utils.Map(
		utils.Filter(
			utils.Keys(data),
			func(key string) bool {
				return self.isStructuralPath(key)
			},
		),
		func(_ int, key string) xml2json.Node {
			return xml2json.NewElementNode(
				key[1:], // drop the leading $ sign
				utils.Ternary(
					key == "$Null",
					[]xml2json.Node{},
					[]xml2json.Node{xml2json.NewTextNode(fmt.Sprint(data[key]))},
				),
				xml2json.NewAttributes(),
			)
		},
	)
}

func (self AnnotationConverter) asConstantWrappers(data map[string]any) []xml2json.Node {
	// first add nodes for all structured properties & constant wrappers
	return utils.Map(
		utils.Filter(
			utils.Keys(data),
			func(key string) bool {
				return self.isConstantWrapper(key)
			},
		),
		func(_ int, key string) xml2json.Node {
			return xml2json.NewElementNode(
				key[1:], // drop the leading $ sign
				utils.Ternary(
					key == "$Null",
					[]xml2json.Node{},
					[]xml2json.Node{xml2json.NewTextNode(fmt.Sprint(data[key]))},
				),
				xml2json.NewAttributes(),
			)
		},
	)
}

func (self AnnotationConverter) extractStructuralPaths(data map[string]any) map[string]any {
	return utils.Projection(
		data,
		utils.Filter(
			utils.Keys(data),
			func(key string) bool { return self.isStructuralPath(key) },
		),
	)
}

func (self AnnotationConverter) extractConstantWrappers(data map[string]any) map[string]any {
	return utils.Projection(
		data,
		utils.Filter(
			utils.Keys(data),
			func(key string) bool { return self.isConstantWrapper(key) },
		),
	)
}

func (self AnnotationConverter) extractRecordProperties(data map[string]any) map[string]any {
	return utils.Projection(
		data,
		utils.Filter(
			utils.Keys(data),
			func(key string) bool { return !self.isStructuralPath(key) && !self.isConstantWrapper(key) },
		),
	)
}
