package edmx

import (
	"github.com/go-errors/errors"
	"github.com/huandu/go-clone"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	xml2json "github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type OverlayProcessor struct {
	PatchDecomposer
	AnnotationConverter

	content    xml2json.Document
	definition model.ResourceDefinition
}

func NewOverlayProcessor(definition model.ResourceDefinition) (*OverlayProcessor, error) {
	parsed, err := marshaller.Unmarshal("application/xml", definition.Content)
	if err != nil {
		return nil, err
	}

	return &OverlayProcessor{
		content:    parsed.(xml2json.Document),
		definition: definition,
	}, nil
}

func (self *OverlayProcessor) Apply(definition model.OverlayDefinition) (model.ResourceDefinition, error) {
	document := clone.Clone(self.content).(xml2json.Document) // Create a copy of the content to avoid mutating the original definition

	for _, patch := range definition.Overlay.Patches {
		for _, decomposed := range self.Decompose(patch) {
			pointer, err := NewPointer(document, decomposed.Selector)
			if err != nil {
				return model.ResourceDefinition{}, err
			}

			document, err = self.reconcile(document, pointer)
			if err != nil {
				return model.ResourceDefinition{}, err
			}

			if document, err = self.process(decomposed, document, pointer); err != nil {
				return model.ResourceDefinition{}, err
			}
		}
	}

	serialized, err := marshaller.Marshal("application/xml", document)
	if err != nil {
		return model.ResourceDefinition{}, err
	}

	return utils.Clone(self.definition, func(rd *model.ResourceDefinition) {
		rd.Content = serialized
		rd.Purpose = definition.Purpose
		rd.Visibility = definition.Overlay.Visibility
	}), nil
}

func (self *OverlayProcessor) asAnnotations(data map[string]any) ([]xml2json.Node, error) {
	nodes := make([]xml2json.Node, 0, len(data))
	for _, annotation := range utils.Keys(data) {
		node, err := self.Convert(annotation, data[annotation])
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (self *OverlayProcessor) reconcile(content xml2json.Document, pointer Pointer) (xml2json.Document, error) {
	expression := jputils.Expr(pointer.Element(), "nodes", jputils.Eq("@.name", "Annotation"))
	annotations := utils.Map(expression.Get(content), func(_ int, a any) xml2json.Node { return a.(xml2json.Node) })

	if len(annotations) == 0 {
		return content, nil
	}

	if _, err := expression.Remove(content); err != nil {
		return nil, err
	}

	return self.merge(content, pointer, annotations)
}

func (self *OverlayProcessor) process(patch model.Patch, content xml2json.Document, pointer Pointer) (xml2json.Document, error) {
	switch patch.Action {
	case "remove":
		return self.remove(content, pointer, utils.SafeCast[map[string]any](patch.Data))
	case "merge":
		annotations, err := self.asAnnotations(utils.SafeCast[map[string]any](patch.Data))
		if err != nil {
			return nil, err
		}
		return self.merge(content, pointer, annotations)
	case "update":
		annotations, err := self.asAnnotations(utils.SafeCast[map[string]any](patch.Data))
		if err != nil {
			return nil, err
		}
		return self.update(content, pointer, annotations)
	default:
		return nil, errors.Errorf("unsupported patch action: %s", patch.Action)
	}
}

func (self *OverlayProcessor) remove(content xml2json.Document, pointer Pointer, data map[string]any) (xml2json.Document, error) {
	if expression := pointer.Annotations(); !expression.Has(content) {
		return content, nil
	} else {
		return xml2json.PruneNodes(
			content,
			expression,
			func(node xml2json.Node) bool { return !utils.ContainsKey(data, annotationName(node)) },
			true,
		)
	}

}

func (self *OverlayProcessor) merge(content xml2json.Document, pointer Pointer, annotations []xml2json.Node) (xml2json.Document, error) {
	if expression := pointer.Annotations(); expression.Has(content) {
		// Replace annotations with the same term and qualifier, while preserving
		// other qualifiers of the same term.
		incoming := make(map[string]bool, len(annotations))
		for _, node := range annotations {
			incoming[annotationIdentity(node)] = true
		}

		content, err := xml2json.PruneNodes(
			content,
			expression,
			func(node xml2json.Node) bool { return !incoming[annotationIdentity(node)] },
		)
		if err != nil {
			return content, err
		}

		return xml2json.AppendNodes(content, expression, annotations...)
	}

	return xml2json.AppendNodes(
		content,
		pointer.Schema(),
		xml2json.NewElementNode("Annotations", annotations, xml2json.NewAttributes("Target", pointer.Target())),
	)
}

func annotationIdentity(node xml2json.Node) string {
	return node.Attribute("Term") + "\x00" + node.Attribute("Qualifier")
}

func annotationName(node xml2json.Node) string {
	name := "@" + node.Attribute("Term")
	if qualifier := node.Attribute("Qualifier"); qualifier != "" {
		name += "#" + qualifier
	}
	return name
}

func (self *OverlayProcessor) update(content xml2json.Document, pointer Pointer, annotations []xml2json.Node) (xml2json.Document, error) {
	if expression := pointer.Annotations(); expression.Has(content) {
		return xml2json.SetNodes(content, expression, annotations...)
	}

	return xml2json.AppendNodes(
		content,
		pointer.Schema(),
		xml2json.NewElementNode("Annotations", annotations, xml2json.NewAttributes("Target", pointer.Target())),
	)
}
