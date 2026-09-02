package xml2json

import (
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

func Pinpoint(content Document, expression jp.Expr) (jp.Expr, bool, *errors.OverlayError) {
	located := expression.Locate(content, 2)

	if len(located) == 0 {
		return nil, false, errors.Create(errors.Severity_Warning, "no such element: %s", expression.String())
	}

	if len(located) > 1 {
		return nil, true, errors.Create(errors.Severity_Warning, "ambiguous expression: %s", expression.String())
	}

	return located[0], true, nil
}

func SetNodes(content Document, expression jp.Expr, nodes ...Node) *errors.OverlayError {
	pexpression, _, err := Pinpoint(content, expression.Child("nodes"))
	if err != nil {
		return err
	}

	return errors.Wrap(pexpression.SetOne(content, nodes), errors.Severity_Warning)
}

func AppendNodes(content Document, expression jp.Expr, nodes ...Node) *errors.OverlayError {
	pexpression, _, err := Pinpoint(content, expression.Child("nodes"))
	if err != nil {
		return err
	}

	return errors.Wrap(pexpression.SetOne(content, append(pexpression.First(content).([]Node), nodes...)), errors.Severity_Warning)
}

func PruneNodes(content Document, expression jp.Expr, predicate func(node Node) bool, del ...bool) *errors.OverlayError {
	deleteIfEmpty := len(del) > 0 && del[0]
	pexpression, _, err := Pinpoint(content, expression.Child("nodes"))
	if err != nil {
		return err
	}

	pruned := utils.Filter(pexpression.First(content).([]Node), predicate)

	if len(pruned) > 0 || !deleteIfEmpty {
		return errors.Wrap(pexpression.SetOne(content, pruned), errors.Severity_Warning)
	}

	return errors.Wrap(utils.Second(pexpression[:len(pexpression)-1].RemoveOne(content)), errors.Severity_Warning)
}
