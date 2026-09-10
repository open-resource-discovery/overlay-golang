package edmx

import (
	"slices"
	"strings"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type PatchDecomposer byte

func (self PatchDecomposer) Decompose(patch model.Patch) []model.Patch {
	if patch.Data == nil {
		return []model.Patch{patch}
	}

	result := make([]model.Patch, 0)
	data := self.preprocess(patch, utils.SafeCast[map[string]any](patch.Data))
	annotations := utils.Filter(utils.Keys(data), func(key string) bool { return key[0] == '@' })
	properties := utils.Filter(utils.Keys(data), func(key string) bool { return key[0] != '@' && key[0] != '$' })

	if len(annotations) > 0 {
		if patch.Action == "merge" {
			// Implement replace behavior for annotations by removing them first and then adding them back with the new values.
			result = append(result, utils.Clone(patch, func(p *model.Patch) {
				p.Action = "remove"
				p.Data = utils.Remap(
					utils.Projection(data, annotations),
					func(key string, _ any) (string, any) { return key, nil },
				)
			}))
		}

		result = append(result, utils.Clone(patch, func(p *model.Patch) { p.Data = utils.Projection(data, annotations) }))
	}

	if patch.Action == "update" && len(annotations) == 0 {
		result = append(result, utils.Clone(patch, func(p *model.Patch) {
			p.Data = nil
			p.Action = "remove"
		}))
	}

	if len(properties) > 0 {
		result = append(
			result,
			utils.Flatten(
				utils.Map(
					utils.Sort(properties),
					func(_ int, property string) []model.Patch {
						return self.Decompose(utils.Clone(patch, func(p *model.Patch) {
							p.Data = data[property]
							p.Selector.Parameter = utils.Ternary(len(patch.Selector.Operation) == 0, "", property)
							p.Selector.PropertyType = utils.Ternary(len(patch.Selector.Operation) == 0, property, "")
						}))
					},
				),
			)...,
		)
	}

	return result
}

func (self PatchDecomposer) preprocess(patch model.Patch, data map[string]any) map[string]any {
	isEnumTypeSelector := len(patch.Selector.EnumType) > 0 && len(patch.Selector.PropertyType) == 0

	return utils.Remap(data, func(key string, value any) (string, any) {
		if !isEnumTypeSelector || slices.Index([]rune(key), '@') < 1 { // either no @ or at the start
			return key, value
		}

		// split enum type member annotations from {<member>@<annotation>: <value>} to {<member>: {<annotation>: <value>}}
		parts := strings.SplitN(key, "@", 2)
		return parts[0], map[string]any{"@" + parts[1]: value}
	})
}
