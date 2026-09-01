package jputils

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"golang.org/x/exp/constraints"
)

func IsRoot(expression jp.Expr) bool {
	return len(expression) == 1 && expression[0] == jp.Root('$')
}

func Root() jp.Expr {
	return jp.Expr{jp.Root('$')}
}

// RemoveAll removes every location selected against the original document.
// Descendants are removed before ancestors, and array siblings are removed
// from the highest index down so earlier removals do not shift later targets.
func RemoveAll(document any, expression jp.Expr) error {
	locations := expression.Locate(document, 0)
	sort.SliceStable(locations, func(i, j int) bool {
		left := locations[i]
		right := locations[j]
		if len(left) != len(right) {
			return len(left) > len(right)
		}

		if 0 < len(left) && left[:len(left)-1].String() == right[:len(right)-1].String() {
			leftIndex, leftIsIndex := left[len(left)-1].(jp.Nth)
			rightIndex, rightIsIndex := right[len(right)-1].(jp.Nth)
			if leftIsIndex && rightIsIndex {
				return leftIndex > rightIndex
			}
		}

		return left.String() > right.String()
	})

	removed := make(map[string]struct{}, len(locations))
	for _, location := range locations {
		key := location.String()
		if _, exists := removed[key]; exists {
			continue
		}
		removed[key] = struct{}{}

		if _, err := location.Remove(document); err != nil {
			return err
		}
	}

	return nil
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

func Pinpoint(document any, expression jp.Expr) (jp.Expr, bool, error) {
	located := expression.Locate(document, 2)

	if IsRoot(expression) {
		return expression, true, nil
	}

	if len(located) == 0 {
		return nil, false, errors.Errorf("no such element: %s", expression.String())
	}

	if len(located) > 1 {
		return nil, true, errors.Errorf("ambiguous expression: %s", expression.String())
	}

	return located[0], true, nil
}
