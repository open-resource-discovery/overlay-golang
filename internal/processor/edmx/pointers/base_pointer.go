package pointers

import (
	"math"
	"strings"

	"github.com/go-errors/errors"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	xml2json "github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type PointerImpl struct {
	kind      string
	target    string
	element   jp.Expr
	document  xml2json.Document
	namespace string
}

func ForOperation(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, name, parameters := resolvers.ParseQualifiedName(selector.Operation)

	for _, candidate := range [][]any{
		{"Action", expressions.Action(namespace, name, parameters)},
		{"Function", expressions.Function(namespace, name, parameters)},
		{"EntityContainer.FunctionImport", expressions.FunctionImport(namespace, name)},
	} {
		if len(parameters) > 0 && "EntityContainer.FunctionImport" == candidate[0] {
			continue // If parameters are present - FunctionImport is not a valid candidate
		}

		if pexpression, found, err := xml2json.Pinpoint(document, utils.SafeCast[jp.Expr](candidate[1])); found && err != nil {
			return nil, err
		} else if found {
			return &PointerImpl{
				document:  document,
				kind:      utils.SafeCast[string](candidate[0]),
				element:   candidate[1].(jp.Expr),
				namespace: resolvers.ResolveNamespace(document, pexpression),
				target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
			}, nil
		}
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForEntityType(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, name, _ := resolvers.ParseQualifiedName(selector.EntityType)
	expression := expressions.EntityType(namespace, name, selector.PropertyType)

	if pexpression, found, err := xml2json.Pinpoint(document, expression); found && err != nil {
		return nil, err
	} else if found {
		return &PointerImpl{
			document:  document,
			element:   expression,
			kind:      utils.Ternary(len(selector.PropertyType) == 0, "EntityType", "EntityType.Property"),
			namespace: resolvers.ResolveNamespace(document, pexpression),
			target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
		}, nil
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForComplexType(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, name, _ := resolvers.ParseQualifiedName(selector.ComplexType)
	expression := expressions.ComplexType(namespace, name, selector.PropertyType)

	if pexpression, found, err := xml2json.Pinpoint(document, expression); found && err != nil {
		return nil, err
	} else if found {
		return &PointerImpl{
			document:  document,
			element:   expression,
			kind:      utils.Ternary(len(selector.PropertyType) == 0, "ComplexType", "ComplexType.Property"),
			namespace: resolvers.ResolveNamespace(document, pexpression),
			target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
		}, nil
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForEnumType(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, name, _ := resolvers.ParseQualifiedName(selector.EnumType)
	expression := expressions.EnumType(namespace, name, selector.PropertyType)

	if pexpression, found, err := xml2json.Pinpoint(document, expression); found && err != nil {
		return nil, err
	} else if found {
		return &PointerImpl{
			document:  document,
			element:   expression,
			kind:      utils.Ternary(len(selector.PropertyType) == 0, "EnumType", "EnumType.Member"),
			namespace: resolvers.ResolveNamespace(document, pexpression),
			target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
		}, nil
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForEntitySet(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, name, _ := resolvers.ParseQualifiedName(selector.EntitySet)
	index := strings.LastIndex(namespace, ".")

	for _, candidate := range [][]any{
		// selector.EntitySet looks like so: "<namespace>.<entity-container>.<entity-set>"
		{"EntityContainer.EntitySet", expressions.EntitySet(
			namespace[:utils.Ternary(index >= 0, index, len(namespace))],
			namespace[utils.Ternary(index >= 0, index+1, int(math.Max(0, float64(len(namespace)-1)))):],
			name,
		)},
		// selector.EntitySet looks like so: "<entity-set>" or "<namespace>.<entity-set>"
		{"EntityContainer.EntitySet", expressions.EntitySet(namespace, "", name)},
	} {
		if pexpression, found, err := xml2json.Pinpoint(document, candidate[1].(jp.Expr)); found && err != nil {
			return nil, err
		} else if found {
			return &PointerImpl{
				document:  document,
				kind:      candidate[0].(string),
				element:   candidate[1].(jp.Expr),
				namespace: resolvers.ResolveNamespace(document, pexpression),
				target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
			}, nil
		}
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForNamespace(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	expression := expressions.Schema(selector.Namespace)

	if pexpression, found, err := xml2json.Pinpoint(document, expression); found && err != nil {
		return nil, err
	} else if found {
		return &PointerImpl{
			document:  document,
			element:   expression,
			kind:      "Schema",
			namespace: resolvers.ResolveNamespace(document, pexpression),
			target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
		}, nil
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForOperationParameter(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, operation, parameters := resolvers.ParseQualifiedName(selector.Operation)

	for _, candidate := range [][]any{
		{"Action.Parameter", expressions.ActionParameter(namespace, operation, parameters, selector.Parameter)},
		{"Function.Parameter", expressions.FunctionParameter(namespace, operation, parameters, selector.Parameter)},
	} {
		if pexpression, found, err := xml2json.Pinpoint(document, candidate[1].(jp.Expr)); found && err != nil {
			return nil, err
		} else if found {
			return &PointerImpl{
				document:  document,
				kind:      candidate[0].(string),
				element:   candidate[1].(jp.Expr),
				namespace: resolvers.ResolveNamespace(document, pexpression),
				target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
			}, nil
		}
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func ForOperationReturnType(document xml2json.Document, selector *model.Selector) (*PointerImpl, error) {
	namespace, operation, parameters := resolvers.ParseQualifiedName(selector.Operation)

	for _, candidate := range [][]any{
		{"Action.ReturnType", expressions.ActionReturnType(namespace, operation, parameters)},
		{"Function.ReturnType", expressions.FunctionReturnType(namespace, operation, parameters)},
	} {
		if pexpression, found, err := xml2json.Pinpoint(document, candidate[1].(jp.Expr)); found && err != nil {
			return nil, err
		} else if found {
			return &PointerImpl{
				document:  document,
				kind:      candidate[0].(string),
				element:   candidate[1].(jp.Expr),
				namespace: resolvers.ResolveNamespace(document, pexpression),
				target:    resolvers.ResolveAnnotationsTarget(document, pexpression),
			}, nil
		}
	}

	return nil, errors.Errorf("no such element: %v", selector)
}

func (self *PointerImpl) Kind() string {
	return self.kind
}

func (self *PointerImpl) Target() string {
	return self.target
}

func (self *PointerImpl) Schema() jp.Expr {
	return expressions.Schema(self.namespace)
}

func (self *PointerImpl) Element() jp.Expr {
	return self.element.Locate(self.document, 1)[0]
}

func (self *PointerImpl) Annotations() jp.Expr {
	return expressions.Annotations(self.namespace, self.target)
}
