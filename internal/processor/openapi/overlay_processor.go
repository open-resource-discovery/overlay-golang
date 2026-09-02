package openapi

import (
	"github.com/huandu/go-clone"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/marshaller"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type OverlayProcessor struct {
	content    map[string]any
	definition model.ResourceDefinition
}

func NewOverlayProcessor(definition model.ResourceDefinition) *OverlayProcessor {
	return &OverlayProcessor{
		definition: definition,
		content:    utils.SafeCast[map[string]any](marshaller.MustUnmarshal(definition.MediaType, definition.Content)),
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
		rd.Content = marshaller.MustMarshal(self.definition.MediaType, content)
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
			expression, err := self.resolve(content, dpatch.Selector)
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

func (self *OverlayProcessor) resolve(content map[string]any, selector *model.Selector) (jp.Expr, *errors.OverlayError) {
	if selector.Root != nil && *selector.Root {
		return jputils.Root(), nil
	}

	if len(selector.JSONPath) > 0 {
		return jputils.Parse(selector.JSONPath)
	}

	if len(selector.Operation) > 0 {
		expression := jputils.Expr(
			"$",
			"paths",
			"*",
			jputils.Eq("@.operationId", selector.Operation),
			utils.Ternary(
				len(selector.Parameter) == 0,
				jputils.Expr(),
				jputils.Expr("parameters", jputils.Eq("@.name", selector.Parameter)),
			),
		)

		return expression, utils.Third(jputils.Pinpoint(content, expression))
	}

	return nil, errors.Create(errors.Severity_Warning, "unsupported selector: %+v", selector)
}

func (self *OverlayProcessor) decompose(content map[string]any, patch model.Patch) ([]model.Patch, *errors.OverlayError) {
	if patch.Data == nil || utils.OneOf(patch.Action, "merge", "update") {
		return []model.Patch{patch}, nil
	}

	result := make([]model.Patch, 0)
	jsonpath, err := self.resolve(content, patch.Selector)
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to resolve selector %+v", patch.Selector)
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
