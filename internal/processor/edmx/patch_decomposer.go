package edmx

import (
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type PatchDecomposer byte

func (self PatchDecomposer) Decompose(patch model.Patch) []model.Patch {
	if patch.Data == nil {
		return []model.Patch{patch}
	}

	result := make([]model.Patch, 0)
	data := utils.SafeCast[map[string]any](patch.Data)

	if annotations := utils.Filter(utils.Keys(data), func(s string) bool { return s[0] == '@' }); len(annotations) > 0 {
		result = append(result, utils.Clone(patch, func(p *model.Patch) { p.Data = utils.Projection(data, annotations) }))
	}

	if properties := utils.Filter(utils.Keys(data), func(s string) bool { return s[0] != '@' && s[0] != '$' }); len(properties) > 0 {
		result = append(
			result,
			utils.Map(
				properties,
				func(_ int, property string) model.Patch {
					return utils.Clone(patch, func(p *model.Patch) {
						p.Data = data[property]
						p.Selector.Parameter = utils.Ternary(len(patch.Selector.Operation) == 0, "", property)
						p.Selector.PropertyType = utils.Ternary(len(patch.Selector.Operation) == 0, property, "")
					})
				},
			)...,
		)
	}

	return result
}
