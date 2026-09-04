package pointers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-errors/errors"
	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
	"github.com/open-resource-discovery/overlay-golang/internal/common/xml2json"
)

var resolvers = (func() *struct {
	ParseQualifiedName       func(value string) (namespace string, name string, parameters []string)
	ResolveNamespace         func(content xml2json.Document, expression jp.Expr) string
	ResolveAnnotationsTarget func(content xml2json.Document, expression jp.Expr) string
} {
	name := func(value string) string {
		return utils.First(utils.Pop(strings.Split(strings.Split(value, "(")[0], ".")))
	}
	namespace := func(value string) string {
		return strings.Join(utils.Second(utils.Pop(strings.Split(strings.Split(value, "(")[0], "."))), ".")
	}
	parameters := func(value string) []string {
		if !utils.First(regexp.MatchString(`.+\(.*\)$`, value)) {
			return nil
		}

		return utils.Filter(
			utils.Map(
				strings.Split(value[strings.Index(value, "(")+1:len(value)-1], ","),
				func(_ int, value string) string { return strings.TrimSpace(value) },
			),
			func(parameter string) bool { return len(parameter) > 0 },
		)
	}
	signature := func(element xml2json.Node) string {
		return fmt.Sprintf(
			"%s(%s)",
			element.Attribute("Name"),
			utils.Join(
				",",
				utils.Map(
					utils.Filter(element.Nodes(), func(node xml2json.Node) bool { return node.Name() == "Parameter" }),
					func(_ int, node xml2json.Node) string { return node.Attribute("Type") },
				)...,
			),
		)
	}

	return &(struct {
		ParseQualifiedName       func(value string) (string, string, []string)
		ResolveNamespace         func(content xml2json.Document, expression jp.Expr) string
		ResolveAnnotationsTarget func(content xml2json.Document, expression jp.Expr) string
	}{
		ParseQualifiedName: func(value string) (string, string, []string) {
			return namespace(value), name(value), parameters(value)
		},
		ResolveNamespace: func(content xml2json.Document, expression jp.Expr) string {
			element := expression.First(content).(xml2json.Node)

			switch name := element.Name(); name {
			case "Schema":
				return element.Attribute("Namespace")
			case "Member", "Property", "Parameter", "EntitySet", "ReturnType", "FunctionImport":
				return expression[:len(expression)-4].First(content).(xml2json.Node).Attribute("Namespace")
			case "Action", "Function", "EnumType", "EntityType", "ComplexType", "EntityContainer":
				return expression[:len(expression)-2].First(content).(xml2json.Node).Attribute("Namespace")
			default:
				panic(errors.Errorf("unexpected element: %s", name))
			}
		},
		ResolveAnnotationsTarget: func(content xml2json.Document, expression jp.Expr) string {
			element := expression.First(content).(xml2json.Node)

			switch name := element.Name(); name {
			case "Schema":
				// <namespace>
				return element.Attribute("Namespace")
			case "Action", "Function":
				// <namespace>.<action>
				// <namespace>.<function>
				return utils.Join(
					".",
					expression[:len(expression)-2].First(content).(xml2json.Node).Attribute("Namespace"),
					signature(element),
				)
			case "Parameter", "ReturnType":
				// <namespace>.<action>/<parameter>
				// <namespace>.<action>/$ReturnType
				// <namespace>.<function>/<parameter>
				// <namespace>.<function>/$ReturnType
				return utils.Join(
					"/",
					utils.Join(
						".",
						expression[:len(expression)-4].First(content).(xml2json.Node).Attribute("Namespace"),
						signature(expression[:len(expression)-2].First(content).(xml2json.Node)),
					),
					utils.Ternary(name == "ReturnType", "$ReturnType", element.Attribute("Name")),
				)
			case "Member", "Property", "EntitySet", "FunctionImport":
				// <namespace>.<enum-type>/<member>
				// <namespace>.<entity-type>/<property>
				// <namespace>.<complex-type>/<property>
				// <namespace>.<entity-container>/<entity-set>
				// <namespace>.<entity-container>/<function-import>
				return utils.Join(
					"/",
					utils.Join(
						".",
						expression[:len(expression)-4].First(content).(xml2json.Node).Attribute("Namespace"),
						expression[:len(expression)-2].First(content).(xml2json.Node).Attribute("Name"),
					),
					element.Attribute("Name"),
				)
			case "EnumType", "EntityType", "ComplexType", "EntityContainer":
				// <namespace>.<enum-type>
				// <namespace>.<entity-type>
				// <namespace>.<complex-type>
				// <namespace>.<entity-container>
				return utils.Join(
					".",
					expression[:len(expression)-2].First(content).(xml2json.Node).Attribute("Namespace"),
					element.Attribute("Name"),
				)
			default:
				panic(errors.Errorf("unexpected element: %s", name))
			}
		},
	})
})()
