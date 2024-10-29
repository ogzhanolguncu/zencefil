package parser

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func PrettifyAST(nodes []Node) {
	var sb strings.Builder
	prettifyNodes(&sb, nodes, 0)
	fmt.Printf("%%\n%s\n", sb.String())
}

func prettifyNodes(sb *strings.Builder, nodes []Node, indent int) {
	for _, node := range nodes {
		sb.WriteString(strings.Repeat("  ", indent))

		// Color for node type
		nodeTypeColor := color.New(color.FgCyan, color.Bold).SprintFunc()

		// Color for node value
		var nodeValueColor func(a ...interface{}) string
		switch node.Type {
		case TEXT_NODE:
			nodeValueColor = color.New(color.FgGreen).SprintFunc()
		case VARIABLE_NODE:
			nodeValueColor = color.New(color.FgYellow).SprintFunc()

		case IF_NODE:
			nodeValueColor = color.New(color.FgMagenta).SprintFunc()
		case THEN_BRANCH:
			nodeValueColor = color.New(color.FgMagenta).SprintFunc()
		case ELIF_BRANCH:
			nodeValueColor = color.New(color.FgMagenta).SprintFunc()
		case ELIF_ITEM:
			nodeValueColor = color.New(color.FgMagenta).SprintFunc()
		case ELSE_BRANCH:
			nodeValueColor = color.New(color.FgMagenta).SprintFunc()

		case FOR_NODE:
			nodeValueColor = color.New(color.FgBlue).SprintFunc()
		case ITERATEE_ITEM:
			nodeValueColor = color.New(color.FgBlue).SprintFunc()
		case ITERATOR_ITEM:
			nodeValueColor = color.New(color.FgBlue).SprintFunc()
		case FOR_BODY:
			nodeValueColor = color.New(color.FgBlue).SprintFunc()

		default:
			nodeValueColor = color.New(color.FgWhite).SprintFunc()
		}

		// Always print the node type
		if node.Value != nil {
			fmt.Fprintf(sb, "%s: %s\n",
				nodeTypeColor(node.Type),
				nodeValueColor(strings.ReplaceAll(strings.ReplaceAll(*node.Value, "\n", "\\n"), "\t", "\\t")))
		} else {
			fmt.Fprintf(sb, "%s: \n", nodeTypeColor(node.Type))
		}

		if len(node.Children) > 0 {
			prettifyNodes(sb, node.Children, indent+1)
		}
	}
}
