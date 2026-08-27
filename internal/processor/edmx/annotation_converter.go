package edmx

import (
	"fmt"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	xml2json "github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

type AnnotationConverter byte

func (self AnnotationConverter) Convert(name string, value any) xml2json.Node {
	switch value.(type) {
	case []any:
		return xml2json.NewElementNode(
			"Annotation",
			[]xml2json.Node{self.asCollectionElement(value.([]any))},
			xml2json.NewAttributes("Term", name[1:]),
		)
	case map[string]any:
		return xml2json.NewElementNode(
			"Annotation",
			[]xml2json.Node{self.asRecordElement(value.(map[string]any))},
			xml2json.NewAttributes("Term", name[1:]),
		)
	default:
		return xml2json.NewElementNode(
			"Annotation",
			[]xml2json.Node{},
			xml2json.NewAttributes("Term", name[1:], self.resolveTypeName(value), fmt.Sprint(value)),
		)
	}
}

func (self AnnotationConverter) resolveTypeName(value any) string {
	switch value.(type) {
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

func (self AnnotationConverter) asPropertyValueElement(key string, value any) xml2json.Node {
	switch value.(type) {
	case []any:
		return xml2json.NewElementNode(
			"PropertyValue",
			[]xml2json.Node{self.asCollectionElement(value.([]any))},
			xml2json.NewAttributes("Property", key),
		)
	case map[string]any:
		return xml2json.NewElementNode(
			"PropertyValue",
			[]xml2json.Node{self.asRecordElement(value.(map[string]any))},
			xml2json.NewAttributes("Property", key),
		)
	default:
		return xml2json.NewElementNode(
			"PropertyValue",
			[]xml2json.Node{},
			xml2json.NewAttributes(
				"Property", key,
				self.resolveTypeName(value), fmt.Sprint(value),
			),
		)
	}
}
