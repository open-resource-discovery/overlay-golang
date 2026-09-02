package csdl

import (
	"strings"

	"github.com/huandu/go-clone"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/errors"
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

func NewOverlayProcessor(definition model.ResourceDefinition) *OverlayProcessor {
	return &OverlayProcessor{
		definition: definition,
		content:    marshaller.MustUnmarshal("application/json", definition.Content).(map[string]any),
	}
}

func (self *OverlayProcessor) Apply(od model.OverlayDefinition) (model.ResourceDefinition, *errors.OverlayError) {
	content := clone.Clone(self.content).(map[string]any)
	aggregated := utils.Reduce(
		od.Overlay.Patches,
		nil,
		func(result *errors.OverlayError, patch model.Patch) *errors.OverlayError {
			return errors.Append(result, errors.WrapPrefix(self.apply(patch, content), "failed to apply patch %+v", patch))
		},
	)

	return utils.Clone(self.definition, func(rd *model.ResourceDefinition) {
		rd.Purpose = od.Purpose
		rd.Visibility = od.Overlay.Visibility
		rd.Content = marshaller.MustMarshal("application/json", content)
	}), aggregated
}

func (self *OverlayProcessor) apply(patch model.Patch, content map[string]any) *errors.OverlayError {
	decomposed, err := self.decompose(content, patch)
	if err != nil {
		return err
	}

	return utils.Reduce(
		decomposed,
		nil,
		func(result *errors.OverlayError, dpatch model.Patch) *errors.OverlayError {
			expression, err := self.Resolve(content, dpatch.Selector)
			if err != nil {
				return errors.Append(result, errors.WrapPrefix(err, "failed to resolve selector %+v", dpatch.Selector))
			}

			switch dpatch.Action {
			case "merge":
				return errors.Append(result, self.merge(content, expression, dpatch.Data))
			case "remove":
				return errors.Append(result, self.remove(content, expression))
			case "update":
				return errors.Append(result, self.update(content, expression, dpatch.Data))
			default:
				return errors.Append(result, errors.Create(errors.Severity_Warning, "unsupported patch action: %s", dpatch.Action))
			}
		},
	)
}

func (self *OverlayProcessor) remove(content map[string]any, expression jp.Expr) *errors.OverlayError {
	if jputils.IsRoot(expression) {
		return errors.Create(errors.Severity_Warning, "removing the document root is not supported")
	}

	return utils.Reduce(
		jputils.SortedLocations(expression, content),
		nil,
		func(result *errors.OverlayError, expression jp.Expr) *errors.OverlayError {
			return errors.Append(result, utils.Second(expression.Remove(content)))
		},
	)
}

func (self *OverlayProcessor) merge(content map[string]any, expression jp.Expr, value any) *errors.OverlayError {
	if jputils.IsRoot(expression) {
		utils.Overwrite(content, utils.SafeCast[map[string]any](utils.DeepMerge(content, value)))
		return nil
	}

	return utils.Reduce(
		// allows for create or update behavior here
		utils.Ternary(!expression.Has(content), []jp.Expr{expression}, jputils.SortedLocations(expression, content)),
		nil,
		func(result *errors.OverlayError, expression jp.Expr) *errors.OverlayError {
			return errors.Append(result, expression.Set(content, utils.DeepMerge(expression.First(content), value)))
		},
	)
}

func (self *OverlayProcessor) update(content map[string]any, expression jp.Expr, value any) *errors.OverlayError {
	if jputils.IsRoot(expression) {
		utils.Overwrite(content, utils.SafeCast[map[string]any](value))
		return nil
	}

	return utils.Reduce(
		// allows for create or update behavior here
		utils.Ternary(!expression.Has(content), []jp.Expr{expression}, jputils.SortedLocations(expression, content)),
		nil,
		func(result *errors.OverlayError, expression jp.Expr) *errors.OverlayError {
			return errors.Append(result, expression.Set(content, value))
		},
	)
}

func (self *OverlayProcessor) decompose(content map[string]any, patch model.Patch) ([]model.Patch, *errors.OverlayError) {
	if len(patch.Selector.JSONPath) > 0 || (patch.Selector.Root != nil && *patch.Selector.Root) {
		return self.decomposeSyntacticSelector(content, patch)
	}

	return self.decomposeSemanticSelector(content, patch)
}

func (self *OverlayProcessor) decomposeSemanticSelector(content map[string]any, patch model.Patch) ([]model.Patch, *errors.OverlayError) {
	result := make([]model.Patch, 0)
	data := utils.SafeCast[map[string]any](patch.Data)
	isEnumTypeMemberSelector := len(patch.Selector.EnumType) > 0 && len(patch.Selector.PropertyType) > 0
	expression, err := self.Resolve(content, patch.Selector)
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to resolve selector %+v", patch.Selector)
	}

	// Remove of an enum type member or update shall remove all existing annotations first
	if patch.Action == "update" || (patch.Action == "remove" && patch.Data == nil && isEnumTypeMemberSelector) {
		prefix := utils.Ternary(!isEnumTypeMemberSelector, "", patch.Selector.PropertyType) + "@"

		for _, candidate := range jputils.Expr(expression, "*").Locate(content, 0) {
			if child, ok := candidate[len(candidate)-1].(jp.Child); ok && strings.HasPrefix(string(child), prefix) {
				result = append(result, utils.Clone(patch, func(p *model.Patch) {
					p.Data = nil
					p.Action = "remove"
					p.Selector = &model.Selector{JSONPath: candidate.String()}
				}))
			}
		}
	}

	// Simple removes can now be applied directly
	if patch.Action == "remove" && patch.Data == nil {
		return append(result, utils.Clone(patch, func(p *model.Patch) {
			p.Selector = &model.Selector{
				JSONPath: utils.Ternary(
					isEnumTypeMemberSelector,
					expression.Child(patch.Selector.PropertyType),
					expression,
				).String(),
			}
		})), nil
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

func (self *OverlayProcessor) decomposeSyntacticSelector(content map[string]any, patch model.Patch) ([]model.Patch, *errors.OverlayError) {
	if patch.Data == nil || utils.OneOf(patch.Action, "merge", "update") {
		return []model.Patch{patch}, nil
	}

	result := make([]model.Patch, 0)
	jsonpath, err := self.Resolve(content, patch.Selector)
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to resolve selector %+v", patch.Selector)
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
