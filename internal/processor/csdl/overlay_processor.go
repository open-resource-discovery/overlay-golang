package csdl

import (
	"maps"
	"strings"

	"github.com/go-errors/errors"
	"github.com/huandu/go-clone"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/patching"
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
	return self.ApplyWithDiagnostics(od)
}

func (self *OverlayProcessor) ApplyWithDiagnostics(od model.OverlayDefinition, handlers ...model.DiagnosticHandler) (model.ResourceDefinition, error) {
	content := clone.Clone(self.content).(map[string]any)
	content, patchErr := patching.Run(
		content,
		od.Overlay.Patches,
		func(value map[string]any) map[string]any { return clone.Clone(value).(map[string]any) },
		patching.CheckJSONPathMatch[map[string]any],
		func(patch model.Patch, candidate map[string]any) (map[string]any, error) {
			decomposed, err := self.decompose(candidate, patch)
			if err != nil {
				return candidate, err
			}
			for _, decomposedPatch := range decomposed {
				candidate, err = self.apply(decomposedPatch, candidate)
				if err != nil {
					return candidate, err
				}
			}
			return candidate, nil
		},
		handlers...,
	)

	serialized, err := marshaller.Marshal("application/json", content)
	if err != nil {
		return model.ResourceDefinition{}, err
	}

	result := utils.Clone(self.definition, func(rd *model.ResourceDefinition) {
		rd.Content = serialized
		rd.Purpose = od.Purpose
		rd.Visibility = od.Overlay.Visibility
	})
	return result, patchErr
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
		return nil, errors.Errorf("removing the document root is not supported")
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

	// Semantic selectors decompose into annotation-leaf writes onto an
	// already-validated element; the leaf key itself does not pre-exist, so we
	// Set it. Raw jsonPath no-match is rejected earlier in decompose, so this
	// never fabricates a user-typo'd target.
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

	// See merge: annotation-leaf writes Set a not-yet-existing leaf; raw
	// jsonPath no-match is rejected earlier in decompose.
	for _, location := range utils.Ternary(!expression.Has(content), []jp.Expr{expression}, expression.Locate(content, 0)) {
		if err := location.Set(content, value); err != nil {
			return nil, err
		}
	}

	return content, nil
}

func (self *OverlayProcessor) decompose(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	if len(patch.Selector.JSONPath) > 0 || (patch.Selector.Root != nil && *patch.Selector.Root) {
		return self.decomposeSyntacticSelector(content, patch)
	}

	return self.decomposeSemanticSelector(content, patch)
}

func (self *OverlayProcessor) decomposeSemanticSelector(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	result := make([]model.Patch, 0)
	data := utils.SafeCast[map[string]any](patch.Data)
	isEnumTypeMemberSelector := len(patch.Selector.EnumType) > 0 && len(patch.Selector.PropertyType) > 0
	expression, err := self.Resolve(content, patch.Selector)
	if err != nil {
		return nil, err
	}

	// Updates for everything but enum types and enum type members shall be applied directly
	if patch.Action == "update" && len(patch.Selector.EnumType) == 0 {
		return append(result, utils.Clone(patch, func(p *model.Patch) {
			p.Selector = &model.Selector{JSONPath: expression.String()}
		})), nil
	}

	// Update/remove of an enum type member shall remove all of its annotations first
	if isEnumTypeMemberSelector && (patch.Action == "update" || (patch.Action == "remove" && patch.Data == nil)) {
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

func (self *OverlayProcessor) decomposeSyntacticSelector(content map[string]any, patch model.Patch) ([]model.Patch, error) {
	if patch.Data == nil || utils.OneOf(patch.Action, "merge", "update") {
		// A raw jsonPath merge/update that matches nothing is a no-op.
		if utils.OneOf(patch.Action, "merge", "update") && len(patch.Selector.JSONPath) > 0 {
			expression, err := jp.ParseString(patch.Selector.JSONPath)
			if err != nil {
				return nil, err
			}
			if !expression.Has(content) {
				return []model.Patch{}, nil
			}
		}

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
