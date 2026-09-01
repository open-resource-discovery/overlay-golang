package edmx

import (
	"strings"

	"github.com/huandu/go-clone"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/overlay-golang/model"
)

var odataV2EDMNamespaces = map[string]struct{}{
	"http://schemas.microsoft.com/ado/2006/04/edm": {},
	"http://schemas.microsoft.com/ado/2007/05/edm": {},
	"http://schemas.microsoft.com/ado/2008/09/edm": {},
}

type OverlayProcessor struct {
	PatchDecomposer
	AnnotationConverter

	content    xml2json.Document
	definition model.ResourceDefinition
}

func NewOverlayProcessor(definition model.ResourceDefinition) *OverlayProcessor {
	return &OverlayProcessor{
		definition: definition,
		content:    marshaller.MustUnmarshal("application/xml", definition.Content).(xml2json.Document),
	}
}

func (self *OverlayProcessor) Apply(definition model.OverlayDefinition) (model.ResourceDefinition, *errors.OverlayError) {
	if isODataV2(self.content) {
		return model.ResourceDefinition{}, errors.Create(errors.Severity_Error,
			"applying an ORD Overlay to OData v2 EDMX is not supported; provide an OData v4 EDMX document that describes the OData v2 API and apply the overlay to that document",
		)
	}

	content := clone.Clone(self.content).(xml2json.Document) // Create a copy of the content to avoid mutating the original definition
	aggregated := utils.Reduce(
		definition.Overlay.Patches,
		nil,
		func(result *errors.OverlayError, patch model.Patch) *errors.OverlayError {
			return errors.Append(result, errors.WrapPrefix(self.apply(patch, content), "failed to apply patch %+v", patch))
		},
	)

	return utils.Clone(self.definition, func(rd *model.ResourceDefinition) {
		rd.Purpose = definition.Purpose
		rd.Visibility = definition.Overlay.Visibility
		rd.Content = marshaller.MustMarshal("application/xml", content)
	}), aggregated
}

func (self *OverlayProcessor) apply(patch model.Patch, document xml2json.Document) *errors.OverlayError {
	return utils.Reduce(
		self.Decompose(patch),
		nil,
		func(result *errors.OverlayError, dpatch model.Patch) *errors.OverlayError {
			pointer, err := NewPointer(document, dpatch.Selector)
			if err != nil {
				return errors.Append(result, err)
			}

			err = self.reconcile(document, pointer)
			if err != nil {
				return errors.Append(result, err)
			}

			return errors.Append(result, self.process(dpatch, document, pointer))
		},
	)
}

func isODataV2(document xml2json.Document) bool {
	for _, dataService := range jputils.Expr(
		"$",
		"nodes",
		jputils.Eq("@.name", "edmx:Edmx"),
		"nodes",
		jputils.Eq("@.name", "edmx:DataServices"),
	).Locate(document, 0) {
		for name, value := range dataService.First(document).(xml2json.Node).Attributes() {
			if strings.HasSuffix(name, ":DataServiceVersion") && strings.HasPrefix(value, "2.") {
				return true
			}
		}
	}

	for _, schema := range jputils.Expr(
		"$",
		"nodes",
		jputils.Eq("@.name", "edmx:Edmx"),
		"nodes",
		jputils.Eq("@.name", "edmx:DataServices"),
		"nodes",
		jputils.Eq("@.name", "Schema"),
	).Locate(document, 0) {
		if _, ok := odataV2EDMNamespaces[schema.First(document).(xml2json.Node).Attribute("xmlns")]; ok {
			return true
		}
	}

	return false
}

func (self *OverlayProcessor) asAnnotations(data map[string]any) []xml2json.Node {
	return utils.Map(
		utils.Keys(data),
		func(_ int, annotation string) xml2json.Node {
			return self.Convert(annotation, data[annotation])
		},
	)
}

func (self *OverlayProcessor) reconcile(content xml2json.Document, pointer Pointer) *errors.OverlayError {
	expression := jputils.Expr(pointer.Element(), "nodes", jputils.Eq("@.name", "Annotation"))
	annotations := utils.Map(expression.Get(content), func(_ int, a any) xml2json.Node { return a.(xml2json.Node) })

	if len(annotations) == 0 {
		return nil
	}

	if _, err := expression.Remove(content); err != nil {
		return errors.Wrap(err, errors.Severity_Warning)
	}

	return self.merge(content, pointer, annotations)
}

func (self *OverlayProcessor) process(patch model.Patch, content xml2json.Document, pointer Pointer) *errors.OverlayError {
	switch patch.Action {
	case "remove":
		return self.remove(content, pointer, utils.SafeCast[map[string]any](patch.Data))
	case "merge":
		return self.merge(content, pointer, self.asAnnotations(utils.SafeCast[map[string]any](patch.Data)))
	case "update":
		return self.update(content, pointer, self.asAnnotations(utils.SafeCast[map[string]any](patch.Data)))
	default:
		return errors.Create(errors.Severity_Warning, "unsupported patch action: %s", patch.Action)
	}
}

func (self *OverlayProcessor) remove(content xml2json.Document, pointer Pointer, data map[string]any) *errors.OverlayError {
	if expression := pointer.Annotations(); !expression.Has(content) {
		return nil
	} else {
		return xml2json.PruneNodes(
			content,
			expression,
			func(node xml2json.Node) bool { return !utils.ContainsKey(data, "@"+node.Attribute("Term")) },
			true,
		)
	}
}

func (self *OverlayProcessor) merge(content xml2json.Document, pointer Pointer, annotations []xml2json.Node) *errors.OverlayError {
	if expression := pointer.Annotations(); expression.Has(content) {
		return xml2json.AppendNodes(content, expression, annotations...)
	}

	return xml2json.AppendNodes(
		content,
		pointer.Schema(),
		xml2json.NewElementNode("Annotations", annotations, xml2json.NewAttributes("Target", pointer.Target())),
	)
}

func (self *OverlayProcessor) update(content xml2json.Document, pointer Pointer, annotations []xml2json.Node) *errors.OverlayError {
	if expression := pointer.Annotations(); expression.Has(content) {
		return xml2json.SetNodes(content, expression, annotations...)
	}

	return xml2json.AppendNodes(
		content,
		pointer.Schema(),
		xml2json.NewElementNode("Annotations", annotations, xml2json.NewAttributes("Target", pointer.Target())),
	)
}
