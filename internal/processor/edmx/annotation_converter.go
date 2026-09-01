package edmx

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	xml2json "github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

type AnnotationConverter byte

func annotationAttributes(name string, values ...string) xml2json.Attributes {
	term, qualifier, qualified := strings.Cut(strings.TrimPrefix(name, "@"), "#")
	attributes := []string{"Term", term}
	if qualified {
		attributes = append(attributes, "Qualifier", qualifier)
	}
	return xml2json.NewAttributes(append(attributes, values...)...)
}

func (self AnnotationConverter) Convert(name string, value any) (xml2json.Node, error) {
	switch value.(type) {
	case []any:
		child, err := self.asCollectionElement(value.([]any))
		if err != nil {
			return xml2json.Node{}, err
		}
		return xml2json.NewElementNode(
			"Annotation",
			[]xml2json.Node{child},
			annotationAttributes(name),
		), nil
	case map[string]any:
		child, err := self.asRecordElement(value.(map[string]any))
		if err != nil {
			return xml2json.Node{}, err
		}
		return xml2json.NewElementNode(
			"Annotation",
			[]xml2json.Node{child},
			annotationAttributes(name),
		), nil
	default:
		typeName, err := self.resolveTypeName(value)
		if err != nil {
			return xml2json.Node{}, err
		}
		return xml2json.NewElementNode(
			"Annotation",
			[]xml2json.Node{},
			annotationAttributes(name, typeName, fmt.Sprint(value)),
		), nil
	}
}

func (self AnnotationConverter) resolveTypeName(value any) (string, error) {
	switch value.(type) {
	case bool:
		return "Bool", nil
	case string:
		return "String", nil
	case float32, float64:
		return "Float", nil
	case int, int8, int16, int32, int64:
		return "Int", nil
	default:
		return "", errors.Errorf("unsupported annotation value type: %T", value)
	}
}

func (self AnnotationConverter) asCollectionElement(values []any) (xml2json.Node, error) {
	nodes := make([]xml2json.Node, 0, len(values))
	for _, value := range values {
		switch value.(type) {
		case []any:
			node, err := self.asCollectionElement(value.([]any))
			if err != nil {
				return xml2json.Node{}, err
			}
			nodes = append(nodes, node)
		case map[string]any:
			node, err := self.asRecordElement(value.(map[string]any))
			if err != nil {
				return xml2json.Node{}, err
			}
			nodes = append(nodes, node)
		default:
			typeName, err := self.resolveTypeName(value)
			if err != nil {
				return xml2json.Node{}, err
			}
			nodes = append(nodes, xml2json.NewElementNode(
				typeName,
				[]xml2json.Node{xml2json.NewTextNode(fmt.Sprint(value))},
				xml2json.NewAttributes(),
			))
		}
	}

	return xml2json.NewElementNode("Collection", nodes, xml2json.NewAttributes()), nil
}

func (self AnnotationConverter) asRecordElement(data map[string]any) (xml2json.Node, error) {
	keys := utils.Keys(data)
	nodes := make([]xml2json.Node, 0, len(keys))
	for _, key := range keys {
		node, err := self.asPropertyValueElement(key, data[key])
		if err != nil {
			return xml2json.Node{}, err
		}
		nodes = append(nodes, node)
	}

	return xml2json.NewElementNode("Record", nodes, xml2json.NewAttributes()), nil
}

func (self AnnotationConverter) asPropertyValueElement(key string, value any) (xml2json.Node, error) {
	switch value.(type) {
	case []any:
		child, err := self.asCollectionElement(value.([]any))
		if err != nil {
			return xml2json.Node{}, err
		}
		return xml2json.NewElementNode(
			"PropertyValue",
			[]xml2json.Node{child},
			xml2json.NewAttributes("Property", key),
		), nil
	case map[string]any:
		child, err := self.asRecordElement(value.(map[string]any))
		if err != nil {
			return xml2json.Node{}, err
		}
		return xml2json.NewElementNode(
			"PropertyValue",
			[]xml2json.Node{child},
			xml2json.NewAttributes("Property", key),
		), nil
	default:
		typeName, err := self.resolveTypeName(value)
		if err != nil {
			return xml2json.Node{}, err
		}
		return xml2json.NewElementNode(
			"PropertyValue",
			[]xml2json.Node{},
			xml2json.NewAttributes("Property", key, typeName, fmt.Sprint(value)),
		), nil
	}
}
