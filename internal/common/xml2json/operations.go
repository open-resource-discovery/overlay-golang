package xml2json

import (
	"github.com/go-errors/errors"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

func Pinpoint(content Document, expression jp.Expr) (jp.Expr, bool, error) {
	located := expression.Locate(content, 2)

	if len(located) == 0 {
		return nil, false, errors.Errorf("no such element: %s", expression.String())
	}

	if len(located) > 1 {
		return nil, true, errors.Errorf("ambiguous expression: %s", expression.String())
	}

	return located[0], true, nil
}

func SetNodes(content Document, expression jp.Expr, nodes ...Node) (Document, error) {
	pexpression, _, err := Pinpoint(content, expression.Child("nodes"))
	if err != nil {
		return content, err
	}

	return content, pexpression.SetOne(content, nodes)
}

func AppendNodes(content Document, expression jp.Expr, nodes ...Node) (Document, error) {
	pexpression, _, err := Pinpoint(content, expression.Child("nodes"))
	if err != nil {
		return content, err
	}

	return content, pexpression.SetOne(content, append(pexpression.First(content).([]Node), nodes...))
}

func PruneNodes(content Document, expression jp.Expr, predicate func(node Node) bool, del ...bool) (Document, error) {
	deleteIfEmpty := len(del) > 0 && del[0]
	pexpression, _, err := Pinpoint(content, expression.Child("nodes"))
	if err != nil {
		return content, err
	}

	pruned := utils.Filter(pexpression.First(content).([]Node), predicate)

	if len(pruned) > 0 || !deleteIfEmpty {
		return content, pexpression.SetOne(content, pruned)
	}

	return content, utils.Second(pexpression[:len(pexpression)-1].RemoveOne(content))
}
