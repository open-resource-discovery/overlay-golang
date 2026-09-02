package jputils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"golang.org/x/exp/constraints"
)

func Parse(value string) (jp.Expr, *errors.OverlayError) {
	result, err := jp.ParseString(value)

	return result, errors.Wrap(err, errors.Severity_Warning)
}

func IsRoot(expression jp.Expr) bool {
	return len(expression) == 1 && expression[0] == jp.Root('$')
}

func Root() jp.Expr {
	return jp.Expr{jp.Root('$')}
}

func Frag(value string) jp.Frag {
	switch value {
	case "@":
		return jp.At('@')
	case "$":
		return jp.Root('$')
	case "*":
		return jp.Wildcard('*')
	default:
		if utils.First(regexp.MatchString(`^\[\d+]$`, value)) {
			return jp.Nth(utils.First(strconv.Atoi(value)))
		}

		return jp.Child(value)
	}
}

func Expr(values ...any) jp.Expr {
	result := make([]jp.Frag, 0, len(values))

	for _, node := range values {
		switch node.(type) {
		case string:
			result = append(result, Frag(utils.SafeCast[string](node)))
		case jp.Expr, []jp.Frag:
			result = append(result, utils.SafeCast[jp.Expr](node)...)
		case *jp.Equation:
			result = append(result, utils.SafeCast[*jp.Equation](node).Filter())
		default:
			result = append(result, node.(jp.Frag))
		}
	}

	return result
}

func Const(value any) *jp.Equation {
	switch value.(type) {
	case bool:
		return jp.ConstBool(value.(bool))
	case string:
		return jp.ConstString(value.(string))
	case float32, float64:
		return jp.ConstFloat(utils.SafeCast[float64](value))
	case int, int8, int16, int32, int64:
		return jp.ConstInt(utils.SafeCast[int64](value))
	default:
		panic(fmt.Sprintf("unsupported constant type: %v", reflect.TypeOf(value)))
	}
}

func Eq[T bool | string | constraints.Float | constraints.Integer](expression string, value T) *jp.Equation {
	return jp.Eq(
		jp.Get(
			Expr(
				utils.Map(
					strings.Split(expression, "."),
					func(_ int, node string) any { return node },
				)...,
			),
		),
		Const(value),
	)
}

func And(first *jp.Equation, second *jp.Equation, rest ...*jp.Equation) *jp.Equation {
	return utils.Reduce(rest, jp.And(first, second), jp.And)
}

func Pinpoint(document any, expression jp.Expr) (jp.Expr, bool, *errors.OverlayError) {
	located := expression.Locate(document, 2)

	if IsRoot(expression) {
		return expression, true, nil
	}

	if len(located) == 0 {
		return nil, false, errors.Create(errors.Severity_Warning, "no such element: %s", expression.String())
	}

	if len(located) > 1 {
		return nil, true, errors.Create(errors.Severity_Warning, "ambiguous expression: %s", expression.String())
	}

	return located[0], true, nil
}
