package xml2json

import (
	"fmt"
	"strings"

	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

var rattributes = strings.NewReplacer("&", "&amp;", "<", "&lt;", `"`, "&quot;")
var relements = strings.NewReplacer("&", "&amp;", "<", "&lt;", `"`, "&quot;", ">", "&gt;")

type Document map[string]any

func NewDocument(nodes []Node, notations []string, declarations []string) Document {
	return map[string]any{
		"nodes":        nodes,
		"notations":    notations,
		"declarations": declarations,
	}
}

func (self Document) Nodes() []Node {
	return utils.SafeCast[[]Node](self["nodes"])
}

func (self Document) Notations() []string {
	return utils.SafeCast[[]string](self["notations"])
}

func (self Document) Declarations() []string {
	return utils.SafeCast[[]string](self["declarations"])
}

func (self Document) ToXML() string {
	var sb strings.Builder

	sb.WriteString(strings.Join(self.Declarations(), "\n"))
	sb.WriteString(strings.Join(self.Notations(), "\n"))

	return self.toXML(&sb, self.Nodes()).String()
}

func (self Document) attributes(node Node) string {
	return strings.Join(
		utils.Map(
			utils.Sort(utils.Keys(node.Attributes())),
			func(_ int, attribute string) string {
				return fmt.Sprintf(`%s="%s"`, attribute, rattributes.Replace(node.Attribute(attribute)))
			},
		),
		" ",
	)
}

func (self Document) element(sb *strings.Builder, node Node) {
	attrs := self.attributes(node)
	element := node.Name()
	children := node.Nodes()

	if len(children) == 0 {
		sb.WriteString(fmt.Sprintf("<%s%s%s/>", element, utils.Ternary(len(attrs) == 0, "", " "), attrs))
	} else if len(children) == 1 {
		if child := children[0]; child.Type() == "text" {
			value := child.Value()
			sb.WriteString(fmt.Sprintf("<%s%s%s>%s</%s>", element, utils.Ternary(len(attrs) == 0, "", " "), attrs, relements.Replace(strings.TrimSpace(value)), element))
		} else {
			sb.WriteString(fmt.Sprintf("<%s%s%s>", element, utils.Ternary(len(attrs) == 0, "", " "), attrs))
			self.toXML(sb, children)
			sb.WriteString(fmt.Sprintf("</%s>", element))
		}
	} else {
		sb.WriteString(fmt.Sprintf("<%s%s%s>", element, utils.Ternary(len(attrs) == 0, "", " "), attrs))
		self.toXML(sb, children)
		sb.WriteString(fmt.Sprintf("</%s>", element))
	}
}

func (self Document) toXML(sb *strings.Builder, nodes []Node) *strings.Builder {
	for _, node := range nodes {
		switch node.Type() {
		case "text":
			sb.WriteString(relements.Replace(node.Value()))
		case "cdata":
			sb.WriteString(fmt.Sprintf("<![CDATA[%s]]>", node.Value()))
		case "processing-instruction":
			sb.WriteString(node.Value())
		default:
			self.element(sb, node)
		}
	}

	return sb
}
