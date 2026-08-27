package csdl

import (
	"math"
	"strings"

	"github.com/go-errors/errors"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type Expressions byte

func (self Expressions) Resolve(document map[string]any, selector *model.Selector) (jp.Expr, error) {
	if selector.Root != nil && *selector.Root {
		return jputils.Root(), nil
	}

	if len(selector.JSONPath) > 0 {
		return jp.ParseString(selector.JSONPath)
	}

	if len(selector.Operation) > 0 {
		if selector.ReturnType != nil && *selector.ReturnType {
			return self.OperationReturnType(document, selector.Operation)
		}

		if len(selector.Parameter) > 0 {
			return self.OperationParameter(document, selector.Operation, selector.Parameter)
		}

		return self.Operation(document, selector.Operation)
	}

	if len(selector.EntityType) > 0 {
		if len(selector.PropertyType) > 0 {
			return self.EntityTypeProperty(document, selector.EntityType, selector.PropertyType)
		}

		return self.EntityType(document, selector.EntityType)
	}

	if len(selector.ComplexType) > 0 {
		if len(selector.PropertyType) > 0 {
			return self.ComplexTypeProperty(document, selector.ComplexType, selector.PropertyType)
		}

		return self.ComplexType(document, selector.ComplexType)
	}

	if len(selector.EnumType) > 0 {
		if len(selector.PropertyType) > 0 {
			return self.EnumTypeMember(document, selector.EnumType, selector.PropertyType)
		}

		return self.EnumType(document, selector.EnumType)
	}

	if len(selector.EntitySet) > 0 {
		return self.EntitySet(document, selector.EntitySet)
	}

	if len(selector.Namespace) > 0 {
		return self.Namespace(document, selector.Namespace)
	}

	return nil, errors.Errorf("unsupported selector: %+v", selector)
}

func (self Expressions) EnumType(document any, fqname string) (jp.Expr, error) {
	namespace, name := self.fqsplit(fqname)
	pexpression, _, err := jputils.Pinpoint(document, jputils.Expr(
		"$",
		utils.Ternary(len(namespace) > 0, namespace, "*"),
		name,
	))

	if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Kind"] != "EnumType" {
		return nil, errors.Errorf("unexpected element found: %v", pexpression.First(document))
	}

	return pexpression, nil
}

func (self Expressions) Namespace(document any, namespace string) (jp.Expr, error) {
	if expression := jputils.Expr("$", namespace); expression.Has(document) {
		return expression, nil
	} else {
		return nil, errors.Errorf("no such element: %s", expression.String())
	}
}

func (self Expressions) Operation(document any, fqname string) (jp.Expr, error) {
	namespace, name := self.fqsplit(fqname)
	pexpression, _, err := jputils.Pinpoint(document, jputils.Expr(
		"$",
		utils.Ternary(len(namespace) > 0, namespace, "*"),
		name,
	))

	if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).([]any); !ok || len(node) != 1 {
		return nil, errors.Errorf("ambiguous expression: %s", pexpression.String())
	}

	// Per OData CSDL JSON spec, overloaded functions/actions are stored as arrays.
	if node, ok := pexpression.Nth(0).First(document).(map[string]any); !ok || !utils.OneOf(node["$Kind"], "Action", "Function") {
		return nil, errors.Errorf("unexpected element found: %v", pexpression.First(document))
	}

	return pexpression.Nth(0), nil
}

func (self Expressions) EntitySet(document any, fqname string) (jp.Expr, error) {
	namespace, name := self.fqsplit(fqname)

	for _, candidate := range utils.Ternary(
		len(namespace) == 0,
		[]jp.Expr{
			// Try if fqname is missing <namespace>.<entity-container> and looks like '<entity-set>'
			jputils.Expr("$", "*", jputils.Eq("@.$Kind", "EntityContainer"), name),
		},
		append(
			[]jp.Expr{
				// Try if fqname is missing <entity-container> and looks like '<namespace>.<entity-set>'
				jputils.Expr("$", namespace, jputils.Eq("@.$Kind", "EntityContainer"), name),
			},
			utils.Ternary(
				strings.LastIndex(namespace, ".") < 0,
				[]jp.Expr{},
				[]jp.Expr{
					// Try if fqname is full and looks like '<namespace>.<entity-container>.<entity-set>'
					jputils.Expr("$", utils.First(self.fqsplit(namespace)), utils.Second(self.fqsplit(namespace)), name),
				},
			)...,
		),
	) {
		if pexpression, found, err := jputils.Pinpoint(document, candidate); !found {
			continue
		} else if err != nil {
			return nil, err
		} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Collection"] != true {
			return nil, errors.Errorf("unexpected element found: %v", pexpression.First(document))
		} else {
			return pexpression, nil
		}
	}

	return nil, errors.Errorf("entity set '%s' not found", fqname)
}

func (self Expressions) EntityType(document any, fqname string) (jp.Expr, error) {
	namespace, name := self.fqsplit(fqname)
	pexpression, _, err := jputils.Pinpoint(document, jputils.Expr("$", utils.Ternary(len(namespace) == 0, "*", namespace), name))

	if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Kind"] != "EntityType" {
		return nil, errors.Errorf("unexpected element found: %v", pexpression.First(document))
	}

	return pexpression, nil
}

func (self Expressions) ComplexType(document any, fqname string) (jp.Expr, error) {
	namespace, name := self.fqsplit(fqname)
	pexpression, _, err := jputils.Pinpoint(document, jputils.Expr("$", utils.Ternary(len(namespace) == 0, "*", namespace), name))

	if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Kind"] != "ComplexType" {
		return nil, errors.Errorf("unexpected element found: %v", pexpression.First(document))
	}

	return pexpression, nil
}

func (self Expressions) EnumTypeMember(document any, fqname string, member string) (jp.Expr, error) {
	parent, err := self.EnumType(document, fqname)
	if err != nil {
		return nil, err
	}

	if expression := parent.Child(member); !expression.Has(document) {
		return nil, errors.Errorf("no such element: %s", expression.String())
	}

	return parent, err // This is ok, see: https://docs.oasis-open.org/odata/odata-csdl-json/v4.01/odata-csdl-json-v4.01.html#sec_EnumerationTypeMember
}

func (self Expressions) OperationParameter(document any, fqname string, parameter string) (jp.Expr, error) {
	parent, err := self.Operation(document, fqname)
	if err != nil {
		return nil, err
	}

	expression := jputils.Expr(parent, "$Parameter", jputils.Eq("@.$Name", parameter))

	return expression, utils.Third(jputils.Pinpoint(document, expression))
}

func (self Expressions) OperationReturnType(document any, fqname string) (jp.Expr, error) {
	parent, err := self.Operation(document, fqname)
	if err != nil {
		return nil, err
	}

	expression := jputils.Expr(parent, "$ReturnType")

	return expression, utils.Third(jputils.Pinpoint(document, expression))
}

func (self Expressions) EntityTypeProperty(document any, fqname string, property string) (jp.Expr, error) {
	parent, err := self.EntityType(document, fqname)
	if err != nil {
		return nil, err
	}

	if expression := parent.Child(property); expression.Has(document) {
		return expression, nil
	} else {
		return nil, errors.Errorf("no such element: %s", expression.String())
	}
}

func (self Expressions) ComplexTypeProperty(document any, fqname string, property string) (jp.Expr, error) {
	parent, err := self.ComplexType(document, fqname)
	if err != nil {
		return nil, err
	}

	if expression := parent.Child(property); expression.Has(document) {
		return expression, nil
	} else {
		return nil, errors.Errorf("no such element: %s", expression.String())
	}
}

func (self Expressions) fqsplit(value string) (namespace string, name string) {
	return value[:int(math.Max(float64(0), float64(strings.LastIndex(value, "."))))],
		value[int(math.Max(float64(0), float64(strings.LastIndex(value, ".")+1))):]
}
