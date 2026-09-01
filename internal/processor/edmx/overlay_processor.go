package edmx

import (
	"github.com/go-errors/errors"
	"github.com/huandu/go-clone"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/patching"
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
	return self.ApplyWithDiagnostics(definition)
}

func (self *OverlayProcessor) ApplyWithDiagnostics(definition model.OverlayDefinition, handlers ...model.DiagnosticHandler) (model.ResourceDefinition, error) {
	document := clone.Clone(self.content).(xml2json.Document)
	document, patchErr := patching.Run(
		document,
		definition.Overlay.Patches,
		func(value xml2json.Document) xml2json.Document { return clone.Clone(value).(xml2json.Document) },
		nil,
		func(patch model.Patch, candidate xml2json.Document) (xml2json.Document, error) {
			var err error
			for _, decomposed := range self.Decompose(patch) {
				pointer, pointerErr := NewPointer(candidate, decomposed.Selector)
				if pointerErr != nil {
					return candidate, pointerErr
				}

				candidate, err = self.reconcile(candidate, pointer)
				if err != nil {
					return candidate, err
				}

				candidate, err = self.process(decomposed, candidate, pointer)
				if err != nil {
					return candidate, err
				}
			}
			return candidate, nil
		},
		handlers...,
	)

	serialized, err := marshaller.Marshal("application/xml", document)
	if err != nil {
		return model.ResourceDefinition{}, err
	}

	result := utils.Clone(self.definition, func(rd *model.ResourceDefinition) {
		rd.Content = serialized
		rd.Purpose = definition.Purpose
		rd.Visibility = definition.Overlay.Visibility
	})
	return result, patchErr
}

func (self *OverlayProcessor) asAnnotations(data map[string]any) []xml2json.Node {
	return utils.Map(
		utils.Keys(data),
		func(_ int, annotation string) xml2json.Node {
			return self.Convert(annotation, data[annotation])
		},
	)
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
		return self.merge(content, pointer, self.asAnnotations(utils.SafeCast[map[string]any](patch.Data)))
	case "update":
		return self.update(content, pointer, self.asAnnotations(utils.SafeCast[map[string]any](patch.Data)))
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
			func(node xml2json.Node) bool { return !utils.ContainsKey(data, "@"+node.Attribute("Term")) },
			true,
		)
	}

}

func (self *OverlayProcessor) merge(content xml2json.Document, pointer Pointer, annotations []xml2json.Node) (xml2json.Document, error) {
	if expression := pointer.Annotations(); expression.Has(content) {
		return xml2json.AppendNodes(content, expression, annotations...)
	}

	return xml2json.AppendNodes(
		content,
		pointer.Schema(),
		xml2json.NewElementNode("Annotations", annotations, xml2json.NewAttributes("Target", pointer.Target())),
	)
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
