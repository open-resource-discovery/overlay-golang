package csdl

import (
	"math"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/model"
)

type Expressions byte

func (self Expressions) Resolve(document map[string]any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	if patch.Selector.Root != nil && *patch.Selector.Root {
		return jputils.Root(), nil
	}

	if len(patch.Selector.JSONPath) > 0 {
		return jputils.Parse(patch.Selector.JSONPath)
	}

	if len(patch.Selector.Operation) > 0 {
		if patch.Selector.ReturnType != nil && *patch.Selector.ReturnType {
			return self.OperationReturnType(document, patch)
		}

		if len(patch.Selector.Parameter) > 0 {
			return self.OperationParameter(document, patch)
		}

		return self.Operation(document, patch)
	}

	if len(patch.Selector.EntityType) > 0 {
		if len(patch.Selector.PropertyType) > 0 {
			return self.EntityTypeProperty(document, patch)
		}

		return self.EntityType(document, patch)
	}

	if len(patch.Selector.ComplexType) > 0 {
		if len(patch.Selector.PropertyType) > 0 {
			return self.ComplexTypeProperty(document, patch)
		}

		return self.ComplexType(document, patch)
	}

	if len(patch.Selector.EnumType) > 0 {
		if len(patch.Selector.PropertyType) > 0 {
			return self.EnumTypeMember(document, patch)
		}

		return self.EnumType(document, patch)
	}

	if len(patch.Selector.EntitySet) > 0 {
		return self.EntitySet(document, patch)
	}

	if len(patch.Selector.Namespace) > 0 {
		return self.Namespace(document, patch)
	}

	return nil, errors.Create(errors.Severity_Warning, "unsupported selector: %+v", patch.Selector)
}

func (self Expressions) EnumType(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	namespace, name := self.fqsplit(patch.Selector.EnumType)
	pexpression, found, err := jputils.Pinpoint(document, jputils.Expr("$", utils.Ternary(len(namespace) > 0, namespace, "*"), name))

	if !found && patch.Action == "remove" {
		return pexpression, nil
	} else if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Kind"] != "EnumType" {
		return nil, errors.Create(errors.Severity_Warning, "unexpected element found: %v", pexpression.First(document))
	}

	return pexpression, nil
}

func (self Expressions) Namespace(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	expression, found, err := jputils.Pinpoint(document, jputils.Expr("$", patch.Selector.Namespace))

	return expression, utils.Ternary(!found && patch.Action == "remove", nil, err)
}

func (self Expressions) Operation(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	namespace, name := self.fqsplit(patch.Selector.Operation)
	pexpression, found, err := jputils.Pinpoint(document, jputils.Expr("$", utils.Ternary(len(namespace) > 0, namespace, "*"), name))

	if !found && patch.Action == "remove" {
		return pexpression, nil
	} else if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).([]any); !ok || len(node) != 1 {
		return nil, errors.Create(errors.Severity_Warning, "ambiguous expression: %s", pexpression.String())
	}

	// TODO - add support for resolving ambiguity based on function signature just like with EDMX

	// Per OData CSDL JSON spec, overloaded functions/actions are stored as arrays.
	if node, ok := pexpression.Nth(0).First(document).(map[string]any); !ok || !utils.OneOf(node["$Kind"], "Action", "Function") {
		return nil, errors.Create(errors.Severity_Warning, "unexpected element found: %v", pexpression.First(document))
	}

	return pexpression.Nth(0), nil
}

func (self Expressions) EntitySet(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	namespace, name := self.fqsplit(patch.Selector.EntitySet)
	candidates := utils.Ternary(
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
	)

	for _, candidate := range candidates {
		if pexpression, found, err := jputils.Pinpoint(document, candidate); !found {
			continue
		} else if err != nil {
			return nil, err
		} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Collection"] != true {
			return nil, errors.Create(errors.Severity_Warning, "unexpected element found: %v", pexpression.First(document))
		} else {
			return pexpression, nil
		}
	}

	return candidates[0], utils.Ternary(patch.Action == "remove", nil, errors.Create(errors.Severity_Warning, "entity set '%s' not found", patch.Selector.EntitySet))
}

func (self Expressions) EntityType(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	namespace, name := self.fqsplit(patch.Selector.EntityType)
	pexpression, found, err := jputils.Pinpoint(document, jputils.Expr("$", utils.Ternary(len(namespace) == 0, "*", namespace), name))

	if !found && patch.Action == "remove" {
		return pexpression, nil
	} else if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Kind"] != "EntityType" {
		return nil, errors.Create(errors.Severity_Warning, "unexpected element found: %v", pexpression.First(document))
	}

	return pexpression, nil
}

func (self Expressions) ComplexType(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	namespace, name := self.fqsplit(patch.Selector.ComplexType)
	pexpression, found, err := jputils.Pinpoint(document, jputils.Expr("$", utils.Ternary(len(namespace) == 0, "*", namespace), name))

	if !found && patch.Action == "remove" {
		return pexpression, nil
	} else if err != nil {
		return nil, err
	} else if node, ok := pexpression.First(document).(map[string]any); !ok || node["$Kind"] != "ComplexType" {
		return nil, errors.Create(errors.Severity_Warning, "unexpected element found: %v", pexpression.First(document))
	}

	return pexpression, nil
}

func (self Expressions) EnumTypeMember(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	parent, err := self.EnumType(document, patch)
	if err != nil {
		return nil, err
	}

	_, found, err := jputils.Pinpoint(document, parent.Child(patch.Selector.PropertyType))

	// This is ok, see: https://docs.oasis-open.org/odata/odata-csdl-json/v4.01/odata-csdl-json-v4.01.html#sec_EnumerationTypeMember
	return parent, utils.Ternary(!found && patch.Action == "remove", nil, err)
}

func (self Expressions) OperationParameter(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	parent, err := self.Operation(document, patch)
	if err != nil {
		return nil, err
	}

	expression := jputils.Expr(parent, "$Parameter", jputils.Eq("@.$Name", patch.Selector.Parameter))
	_, found, err := jputils.Pinpoint(document, expression)

	return expression, utils.Ternary(!found && patch.Action == "remove", nil, err)
}

func (self Expressions) OperationReturnType(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	parent, err := self.Operation(document, patch)
	if err != nil {
		return nil, err
	}

	expression, found, err := jputils.Pinpoint(document, jputils.Expr(parent, "$ReturnType"))

	return expression, utils.Ternary(!found && patch.Action == "remove", nil, err)
}

func (self Expressions) EntityTypeProperty(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	parent, err := self.EntityType(document, patch)
	if err != nil {
		return nil, err
	}

	expression, found, err := jputils.Pinpoint(document, parent.Child(patch.Selector.PropertyType))

	return expression, utils.Ternary(!found && patch.Action == "remove", nil, err)
}

func (self Expressions) ComplexTypeProperty(document any, patch model.Patch) (jp.Expr, *errors.OverlayError) {
	parent, err := self.ComplexType(document, patch)
	if err != nil {
		return nil, err
	}

	expression, found, err := jputils.Pinpoint(document, parent.Child(patch.Selector.PropertyType))

	return expression, utils.Ternary(!found && patch.Action == "remove", nil, err)
}

func (self Expressions) fqsplit(value string) (namespace string, name string) {
	return value[:int(math.Max(float64(0), float64(strings.LastIndex(value, "."))))],
		value[int(math.Max(float64(0), float64(strings.LastIndex(value, ".")+1))):]
}
