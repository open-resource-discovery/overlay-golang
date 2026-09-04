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
	case []any, map[string]any:
		return ""
	case bool:
		return "Bool"
	case string:
		return "String"
	case float32, float64:
		return "Float"
	case int, int8, int16, int32, int64:
		return "Int"
	default:
		panic(fmt.Sprintf("unsupported value type: %T", value))
	}
}

func (self AnnotationConverter) asChildren(value any) []xml2json.Node {
	switch value.(type) {
	case []any:
		return []xml2json.Node{self.asCollectionElement(value.([]any))}
	case map[string]any:
		return []xml2json.Node{self.asRecordElement(value.(map[string]any))}
	default:
		return []xml2json.Node{}
	}
}

func (self AnnotationConverter) asCollectionElement(values []any) xml2json.Node {
	return xml2json.NewElementNode(
		"Collection",
		utils.Map(values, func(_ int, value any) xml2json.Node {
			switch value.(type) {
			case []any:
				return self.asCollectionElement(value.([]any))
			case map[string]any:
				return self.asRecordElement(value.(map[string]any))
			default:
				return xml2json.NewElementNode(
					self.resolveTypeName(value),
					[]xml2json.Node{xml2json.NewTextNode(fmt.Sprint(value))},
					xml2json.NewAttributes(),
				)
			}
		}),
		xml2json.NewAttributes(),
	)
}

func (self AnnotationConverter) asRecordElement(data map[string]any) xml2json.Node {
	return xml2json.NewElementNode(
		"Record",
		utils.Map(
			utils.Keys(data),
			func(_ int, key string) xml2json.Node {
				return self.asPropertyValueElement(key, data[key])
			},
		),
		xml2json.NewAttributes(),
	)
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

func (self AnnotationConverter) asPropertyValueElement(key string, value any) xml2json.Node {
	return xml2json.NewElementNode(
		"PropertyValue",
		self.asChildren(value),
		self.asAttributes(
			[]string{"Property", key},
			[]string{self.resolveTypeName(value), fmt.Sprint(value)},
		),
	)
}
