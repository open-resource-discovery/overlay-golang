package yaml

import (
	"maps"

	"github.com/go-errors/errors"
	"github.com/huandu/go-clone"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type OverlayProcessor struct {
	content    map[string]any
	definition model.ResourceDefinition
}

func NewOverlayProcessor(definition model.ResourceDefinition) (*OverlayProcessor, error) {
	parsed, err := marshaller.Unmarshal("application/yaml", definition.Content)
	if err != nil {
		return nil, err
	}

	return &OverlayProcessor{
		content:    utils.SafeCast[map[string]any](parsed),
		definition: definition,
	}, nil
}

func (self *OverlayProcessor) Apply(od model.OverlayDefinition) (model.ResourceDefinition, error) {
	content, err := utils.SafeCast[map[string]any](clone.Clone(self.content)), error(nil)

	for _, patch := range od.Overlay.Patches {
		decomposed, err := self.decompose(content, patch)
		if err != nil {
			return model.ResourceDefinition{}, err
		}

		for _, dpatch := range decomposed {
			if content, err = self.apply(dpatch, content); err != nil {
				return model.ResourceDefinition{}, err
			}
		}
	}

	serialized, err := marshaller.Marshal("application/yaml", content)
	if err != nil {
		return model.ResourceDefinition{}, err
	}

	return utils.Clone(self.definition, func(rd *model.ResourceDefinition) {
		rd.Content = serialized
		rd.Purpose = od.Purpose
		rd.Visibility = od.Overlay.Visibility
	}), nil
}

func (self *OverlayProcessor) apply(patch model.Patch, content map[string]any) (map[string]any, error) {
	expression, err := self.resolve(patch.Selector)
	if err != nil {
		return nil, err
	}

	switch patch.Action {
	case "merge":
		return self.merge(content, expression, patch.Data)
	case "remove":
		return self.remove(content, expression)
	case "update":
		return self.update(content, expression, patch.Data)
	default:
		return nil, errors.Errorf("unsupported patch action: %s", patch.Action)
	}
}

func (self *OverlayProcessor) remove(content map[string]any, expression jp.Expr) (map[string]any, error) {
	if jputils.IsRoot(expression) {
		return nil, errors.Errorf("removing the document root is not supported")
	}

	locations := expression.Locate(content, 0)
	if len(locations) == 0 {
		return nil, errors.Errorf("no such element: %s", expression.String())
	}

	for _, location := range locations {
		if _, err := location.Remove(content); err != nil {
			return nil, err
		}
	}

	return content, nil
}

func (self *OverlayProcessor) merge(content map[string]any, expression jp.Expr, value any) (map[string]any, error) {
	if jputils.IsRoot(expression) {
		return utils.SafeCast[map[string]any](utils.DeepMerge(content, value)), nil
	}

	// allows for create or update behavior here
	for _, location := range utils.Ternary(!expression.Has(content), []jp.Expr{expression}, expression.Locate(content, 0)) {
		if err := location.Set(content, utils.DeepMerge(location.First(content), value)); err != nil {
			return nil, err
		}
	}

	return content, nil
}

func (self *OverlayProcessor) update(content map[string]any, expression jp.Expr, value any) (map[string]any, error) {
	if jputils.IsRoot(expression) {
		return maps.Clone(utils.SafeCast[map[string]any](value)), nil
	}

	// allows for create or update behavior here
	for _, location := range utils.Ternary(!expression.Has(content), []jp.Expr{expression}, expression.Locate(content, 0)) {
		if err := location.Set(content, value); err != nil {
			return nil, err
		}
	}

	return content, nil
}

func (self *OverlayProcessor) resolve(selector *model.Selector) (jp.Expr, error) {
	if selector.Root != nil && *selector.Root {
		return jputils.Root(), nil
	}

	if len(selector.JSONPath) > 0 {
		return jp.ParseString(selector.JSONPath)
	}

	return nil, errors.Errorf("unsupported selector: %+v", selector)
}

func (self *OverlayProcessor) decompose(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	if patch.Data == nil || utils.OneOf(patch.Action, "merge", "update") {
		return []model.Patch{patch}, nil
	}

	result := make([]model.Patch, 0)
	jsonpath, err := self.resolve(patch.Selector)
	if err != nil {
		return nil, err
	}

	for key, value := range utils.SafeCast[map[string]any](patch.Data) {
		if value != nil && !utils.CanCast[map[string]any](value) {
			continue // ignore these keys for now
		}

		children, err := self.decompose(content, utils.Clone(patch, func(p *model.Patch) {
			p.Data = value
			p.Selector = &model.Selector{JSONPath: jsonpath.Child(key).String()}
		}))
		if err != nil {
			return nil, err
		}

		result = append(result, children...)
	}

	return result, nil
}
