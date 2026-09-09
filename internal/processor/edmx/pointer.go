package edmx

import (
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/overlay-golang/internal/processor/edmx/pointers"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type Pointer interface {
	Kind() string
	IsNil() bool
	Target() string
	Schema() jp.Expr
	Element() jp.Expr
	Annotations() jp.Expr
}

func NewPointer(content xml2json.Document, patch model.Patch) (Pointer, *errors.OverlayError) {
	if len(patch.Selector.EnumType) > 0 {
		return pointers.ForEnumType(content, patch)
	}

	if len(patch.Selector.Operation) > 0 {
		if patch.Selector.ReturnType != nil && *patch.Selector.ReturnType {
			return pointers.ForOperationReturnType(content, patch)
		}

		if len(patch.Selector.Parameter) > 0 {
			return pointers.ForOperationParameter(content, patch)
		}

		return pointers.ForOperation(content, patch)
	}

	if len(patch.Selector.EntitySet) > 0 {
		return pointers.ForEntitySet(content, patch)
	}

	if len(patch.Selector.EntityType) > 0 {
		return pointers.ForEntityType(content, patch)
	}

	if len(patch.Selector.ComplexType) > 0 {
		return pointers.ForComplexType(content, patch)
	}

	if len(patch.Selector.Namespace) > 0 {
		return pointers.ForNamespace(content, patch)
	}

	return nil, errors.Create(errors.Severity_Warning, "unsupported selector: %+v", patch.Selector)
}
