package csdl

import (
	"maps"
	"strings"

	"github.com/go-errors/errors"
	"github.com/huandu/go-clone"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type OverlayProcessor struct {
	Expressions

	content    map[string]any
	definition model.ResourceDefinition
}

func NewOverlayProcessor(definition model.ResourceDefinition) (*OverlayProcessor, error) {
	parsed, err := marshaller.Unmarshal("application/json", definition.Content)
	if err != nil {
		return nil, err
	}

	return &OverlayProcessor{
		content:    parsed.(map[string]any),
		definition: definition,
	}, nil
}

func (self *OverlayProcessor) Apply(od model.OverlayDefinition) (model.ResourceDefinition, error) {
	content, err := clone.Clone(self.content).(map[string]any), error(nil)

	for _, patch := range od.Overlay.Patches {
		decomposed, err := self.decompose(content, patch)
		if err != nil {
			return model.ResourceDefinition{}, err
		}

		for _, decomposed := range decomposed {
			if content, err = self.apply(decomposed, content); err != nil {
				return model.ResourceDefinition{}, err
			}
		}
	}

	serialized, err := marshaller.Marshal("application/json", content)
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
	expression, err := self.Resolve(content, patch.Selector)
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
		return make(map[string]any), nil
	}

	for _, location := range expression.Locate(content, 0) {
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

func (self *OverlayProcessor) decompose(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	if patch.Action == "remove" && patch.Data == nil && len(patch.Selector.EnumType) > 0 && len(patch.Selector.PropertyType) > 0 {
		return self.decomposeEnumMemberRemove(content, patch)
	}

	if patch.Data == nil || len(patch.Selector.JSONPath) > 0 || (patch.Selector.Root != nil && *patch.Selector.Root) {
		return self.decomposeSyntacticSelector(content, patch)
	}

	return self.decomposeSemanticSelector(content, patch)
}

func (self *OverlayProcessor) decomposeEnumMemberRemove(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	expression, err := self.Resolve(content, patch.Selector)
	if err != nil {
		return nil, err
	}

	result := make([]model.Patch, 0)
	prefix := patch.Selector.PropertyType + "@"
	for _, candidate := range jputils.Expr(expression, "*").Locate(content, 0) {
		child, ok := candidate[len(candidate)-1].(jp.Child)
		if !ok {
			continue
		}

		name := string(child)
		if name != patch.Selector.PropertyType && !strings.HasPrefix(name, prefix) {
			continue
		}

		result = append(result, utils.Clone(patch, func(p *model.Patch) {
			p.Selector = &model.Selector{JSONPath: candidate.String()}
		}))
	}

	return result, nil
}

func (self *OverlayProcessor) decomposeSemanticSelector(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	result := make([]model.Patch, 0)
	data := utils.SafeCast[map[string]any](patch.Data)
	expression, err := self.Resolve(content, patch.Selector)
	if err != nil {
		return nil, err
	}

	if patch.Action == "update" {
		// Updates for everything but an enum type member can be applied directly
		if len(patch.Selector.EnumType) == 0 || len(patch.Selector.PropertyType) == 0 {
			return append(result, utils.Clone(patch, func(p *model.Patch) {
				p.Selector = &model.Selector{JSONPath: expression.String()}
			})), nil
		}

		// Special handling for update of an enum type member - remove all existing annotations first
		for _, candidate := range jputils.Expr(expression, "*").Locate(content, 0) {
			if child, ok := candidate[len(candidate)-1].(jp.Child); ok && strings.HasPrefix(string(child), patch.Selector.PropertyType+"@") {
				result = append(result, utils.Clone(patch, func(p *model.Patch) {
					p.Data = nil
					p.Action = "remove"
					p.Selector = &model.Selector{JSONPath: candidate.String()}
				}))
			}
		}
	}

	// Decompose the remaining annotations into separate patches, each with a selector that includes the annotation name
	for _, annotation := range utils.Filter(utils.Keys(data), func(s string) bool { return s[0] == '@' }) {
		result = append(result, utils.Clone(
			patch,
			func(p *model.Patch) {
				p.Data = data[annotation]
				p.Selector = &model.Selector{
					// Prepend EnumType member names before the annotation name
					// See: https://docs.oasis-open.org/odata/odata-csdl-json/v4.01/odata-csdl-json-v4.01.html#sec_EnumerationTypeMember
					JSONPath: jputils.Expr(expression, utils.Ternary(len(patch.Selector.EnumType) == 0, "", patch.Selector.PropertyType)+annotation).String(),
				}
			},
		))
	}

	// Decompose the remaining properties into separate patches, each with a selector that includes the property name
	for _, property := range utils.Filter(utils.Keys(data), func(s string) bool { return s[0] != '@' && s[0] != '$' }) {
		decomposed, err := self.decomposeSemanticSelector(content, utils.Clone(patch, func(p *model.Patch) {
			p.Data = data[property]
			p.Selector.Parameter = utils.Ternary(len(patch.Selector.Operation) == 0, "", property)
			p.Selector.PropertyType = utils.Ternary(len(patch.Selector.Operation) == 0, property, "")
		}))
		if err != nil {
			return nil, err
		}

		result = append(result, decomposed...)
	}

	return result, nil
}

func (self *OverlayProcessor) decomposeSyntacticSelector(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	if patch.Data == nil || utils.OneOf(patch.Action, "merge", "update") {
		return []model.Patch{patch}, nil
	}

	result := make([]model.Patch, 0)
	jsonpath, err := self.Resolve(content, patch.Selector)
	if err != nil {
		return nil, err
	}

	for key, value := range utils.SafeCast[map[string]any](patch.Data) {
		if value != nil && !utils.CanCast[map[string]any](value) {
			continue // ignore these keys for now
		}

		children, err := self.decomposeSyntacticSelector(content, utils.Clone(patch, func(p *model.Patch) {
			p.Data = value
			p.Selector = &model.Selector{JSONPath: jputils.Expr(jsonpath, key).String()}
		}))
		if err != nil {
			return nil, err
		}

		result = append(result, children...)
	}

	return result, nil
}
