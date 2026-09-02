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
	Target() string
	Element() jp.Expr
	Schema() jp.Expr
	Annotations() jp.Expr
}

func NewPointer(content xml2json.Document, selector *model.Selector) (Pointer, *errors.OverlayError) {
	if len(selector.EnumType) > 0 {
		return pointers.ForEnumType(content, selector)
	}

	if len(selector.Operation) > 0 {
		if selector.ReturnType != nil && *selector.ReturnType {
			return pointers.ForOperationReturnType(content, selector)
		}

		if len(selector.Parameter) > 0 {
			return pointers.ForOperationParameter(content, selector)
		}

		return pointers.ForOperation(content, selector)
	}

	if len(selector.EntitySet) > 0 {
		return pointers.ForEntitySet(content, selector)
	}

	if len(selector.EntityType) > 0 {
		return pointers.ForEntityType(content, selector)
	}

	if len(selector.ComplexType) > 0 {
		return pointers.ForComplexType(content, selector)
	}

	if len(selector.Namespace) > 0 {
		return pointers.ForNamespace(content, selector)
	}

	return nil, errors.Create(errors.Severity_Warning, "unsupported selector: %+v", selector)
}
