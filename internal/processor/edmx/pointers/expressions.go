package pointers

import (
	"fmt"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/jputils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

var expressions = (func() *struct {
	Schema             func(namespace string) jp.Expr
	Action             func(namespace string, name string, parameters []string) jp.Expr
	EnumType           func(namespace string, name string, member string) jp.Expr
	Function           func(namespace string, name string, parameters []string) jp.Expr
	EntitySet          func(namespace string, container string, name string) jp.Expr
	EntityType         func(namespace string, name string, property string) jp.Expr
	Annotations        func(namespace string, target string) jp.Expr
	ComplexType        func(namespace string, name string, property string) jp.Expr
	FunctionImport     func(namespace string, name string) jp.Expr
	ActionParameter    func(namespace string, operation string, parameters []string, parameter string) jp.Expr
	ActionReturnType   func(namespace string, operation string, parameters []string) jp.Expr
	FunctionParameter  func(namespace string, operation string, parameters []string, parameter string) jp.Expr
	FunctionReturnType func(namespace string, operation string, parameters []string) jp.Expr
} {
	_Schema := func(namespace string) jp.Expr {
		return jputils.Expr(
			"$",
			"nodes",
			jputils.Eq("@.name", "edmx:Edmx"),
			"nodes",
			jputils.Eq("@.name", "edmx:DataServices"),
			"nodes",
			utils.Ternary(
				len(namespace) == 0,
				jputils.Eq("@.name", "Schema"),
				jputils.And(
					jputils.Eq("@.name", "Schema"),
					jputils.Eq("@.attributes.Namespace", namespace),
				),
			),
		)
	}
	_Action := func(namespace string, name string, parameters []string) jp.Expr {
		return jputils.Expr(
			_Schema(namespace),
			"nodes",
			jputils.And(
				jputils.Eq("@.name", "Action"),
				jputils.Eq("@.attributes.Name", name),
				utils.Map(
					parameters,
					func(idx int, parameter string) *jp.Equation {
						return jputils.Eq(fmt.Sprintf("@.nodes.[%d].attributes.Type", idx), parameter)
					},
				)...,
			),
		)
	}
	_Function := func(namespace string, name string, parameters []string) jp.Expr {
		return jputils.Expr(
			_Schema(namespace),
			"nodes",
			jputils.And(
				jputils.Eq("@.name", "Function"),
				jputils.Eq("@.attributes.Name", name),
				utils.Map(
					parameters,
					func(idx int, parameter string) *jp.Equation {
						return jputils.Eq(fmt.Sprintf("@.nodes.[%d].attributes.Type", idx), parameter)
					},
				)...,
			),
		)
	}

	return &(struct {
		Schema             func(string) jp.Expr
		Action             func(namespace string, name string, parameters []string) jp.Expr
		EnumType           func(namespace string, name string, member string) jp.Expr
		Function           func(namespace string, name string, parameters []string) jp.Expr
		EntitySet          func(namespace string, container string, name string) jp.Expr
		EntityType         func(namespace string, name string, property string) jp.Expr
		Annotations        func(namespace string, target string) jp.Expr
		ComplexType        func(namespace string, name string, property string) jp.Expr
		FunctionImport     func(namespace string, name string) jp.Expr
		ActionParameter    func(namespace string, operation string, parameters []string, parameter string) jp.Expr
		ActionReturnType   func(namespace string, operation string, parameters []string) jp.Expr
		FunctionParameter  func(namespace string, operation string, parameters []string, parameter string) jp.Expr
		FunctionReturnType func(namespace string, operation string, parameters []string) jp.Expr
	}{
		Schema:   _Schema,
		Action:   _Action,
		Function: _Function,
		EnumType: func(namespace string, name string, member string) jp.Expr {
			return jputils.Expr(
				_Schema(namespace),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "EnumType"),
					jputils.Eq("@.attributes.Name", name),
				),
				utils.Ternary(
					len(member) == 0,
					jputils.Expr(),
					jputils.Expr(
						"nodes",
						jputils.And(
							jputils.Eq("@.name", "Member"),
							jputils.Eq("@.attributes.Name", member),
						),
					),
				),
			)
		},
		EntitySet: func(namespace string, container string, name string) jp.Expr {
			return jputils.Expr(
				_Schema(namespace),
				"nodes",
				utils.Ternary(
					len(container) == 0,
					jputils.Eq("@.name", "EntityContainer"),
					jputils.And(
						jputils.Eq("@.name", "EntityContainer"),
						jputils.Eq("@.attributes.Name", container),
					),
				),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "EntitySet"),
					jputils.Eq("@.attributes.Name", name),
				),
			)
		},
		EntityType: func(namespace string, name string, property string) jp.Expr {
			return jputils.Expr(
				_Schema(namespace),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "EntityType"),
					jputils.Eq("@.attributes.Name", name),
				),
				utils.Ternary(
					len(property) == 0,
					jputils.Expr(),
					jputils.Expr(
						"nodes",
						jputils.And(
							jputils.Eq("@.name", "Property"),
							jputils.Eq("@.attributes.Name", property),
						),
					),
				),
			)
		},
		Annotations: func(namespace string, target string) jp.Expr {
			return jputils.Expr(
				_Schema(namespace),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "Annotations"),
					jputils.Eq("@.attributes.Target", target),
				),
			)
		},
		ComplexType: func(namespace string, name string, property string) jp.Expr {
			return jputils.Expr(
				_Schema(namespace),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "ComplexType"),
					jputils.Eq("@.attributes.Name", name),
				),
				utils.Ternary(
					len(property) == 0,
					jputils.Expr(),
					jputils.Expr(
						"nodes",
						jputils.And(
							jputils.Eq("@.name", "Property"),
							jputils.Eq("@.attributes.Name", property),
						),
					),
				),
			)
		},
		FunctionImport: func(namespace string, name string) jp.Expr {
			return jputils.Expr(
				_Schema(namespace),
				"nodes",
				jputils.Eq("@.name", "EntityContainer"),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "FunctionImport"),
					jputils.Eq("@.attributes.Name", name),
				),
			)
		},
		ActionParameter: func(namespace string, operation string, parameters []string, parameter string) jp.Expr {
			return jputils.Expr(
				_Action(namespace, operation, parameters),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "Parameter"),
					jputils.Eq("@.attributes.Name", parameter),
				),
			)
		},
		ActionReturnType: func(namespace string, operation string, parameters []string) jp.Expr {
			return jputils.Expr(
				_Action(namespace, operation, parameters),
				"nodes",
				jputils.Eq("@.name", "ReturnType"),
			)
		},
		FunctionParameter: func(namespace string, operation string, parameters []string, parameter string) jp.Expr {
			return jputils.Expr(
				_Function(namespace, operation, parameters),
				"nodes",
				jputils.And(
					jputils.Eq("@.name", "Parameter"),
					jputils.Eq("@.attributes.Name", parameter),
				),
			)
		},
		FunctionReturnType: func(namespace string, operation string, parameters []string) jp.Expr {
			return jputils.Expr(
				_Function(namespace, operation, parameters),
				"nodes",
				jputils.Eq("@.name", "ReturnType"),
			)
		},
	})
})()
