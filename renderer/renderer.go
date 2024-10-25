package renderer

import (
	"fmt"
	"strings"

	"github.com/ogzhanolguncu/zencefil/parser"
)

type Renderer struct {
	Context map[string]interface{}
	AST     []parser.Node
}

func New(ast []parser.Node, context map[string]interface{}) *Renderer {
	return &Renderer{
		Context: context,
		AST:     ast,
	}
}

func (r *Renderer) Render() (string, error) {
	var sb strings.Builder
	for _, node := range r.AST {
		switch node.Type {
		case parser.TEXT_NODE:
			sb.WriteString(*node.Value)
		case parser.VARIABLE_NODE:
			variable, found := r.variableLookup(*node.Value)
			if !found {
				return "", fmt.Errorf("couldn't find variable '%s' in context", *node.Value)
			}
			sb.WriteString(fmt.Sprintf("%v", variable))
		case parser.IF_NODE:
			variable, found := r.variableLookup(*node.Value)
			if !found {
				return "", fmt.Errorf("couldn't find variable '%s' in context", *node.Value)
			}

			// Safe type assertion with ok check
			boolVal, ok := variable.(bool)
			if !ok {
				return "", fmt.Errorf("variable '%s' is not a boolean", *node.Value)
			}

			if boolVal {
				for _, thenItem := range node.Children {
					if thenItem.Type == parser.THEN_BRANCH {
						for _, child := range thenItem.Children {
							if child.Type == parser.TEXT_NODE {
								sb.WriteString(*child.Value)
							}
						}
					}
				}
			} else {
				// Handle else branch later
			}
		}
	}
	return sb.String(), nil
}

func (r *Renderer) variableLookup(key string) (interface{}, bool) {
	value, exists := r.Context[key]
	return value, exists
}
