package xml2json

import (
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/open-resource-discovery/overlay-golang/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

func Convert(xml string) (Document, *errors.OverlayError) {
	parsed, err := xmlquery.Parse(strings.NewReader(xml))
	if err != nil {
		return nil, errors.Wrap(err, errors.Severity_Error)
	}

	return NewDocument(nodes(parsed), notations(parsed), declarations(parsed)), nil
}

func nodes(parent *xmlquery.Node) []Node {
	result := make([]Node, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case xmlquery.TextNode:
			if strings.TrimSpace(child.Data) != "" {
				result = append(result, NewTextNode(child.Data))
			}
		case xmlquery.ElementNode:
			result = append(result, NewElementNode(name(child.Prefix, child.Data), nodes(child), attributes(child.Attr...)))
		case xmlquery.CharDataNode:
			result = append(result, NewCdataNode(child.Data))
		case xmlquery.ProcessingInstruction:
			result = append(result, NewProcessingInstructionNode(child.OutputXML(true)))
		}
	}

	return result
}

func name(namespace, local string) string {
	return utils.Join(":", namespace, local)
}

func notations(parent *xmlquery.Node) []string {
	result := make([]string, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.NotationNode {
			result = append(result, child.OutputXML(true))
		}
	}

	return result
}

func declarations(parent *xmlquery.Node) []string {
	result := make([]string, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.DeclarationNode {
			result = append(result, child.OutputXML(true))
		}
	}

	return result
}

func attributes(attributes ...xmlquery.Attr) map[string]string {
	return NewAttributes(
		utils.Flatten(
			utils.Map(
				attributes,
				func(_ int, a xmlquery.Attr) []string {
					return []string{name(a.Name.Space, a.Name.Local), a.Value}
				},
			),
		)...,
	)
}
