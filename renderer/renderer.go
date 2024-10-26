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
	if context == nil {
		context = make(map[string]interface{})
	}
	return &Renderer{
		Context: context,
		AST:     ast,
	}
}

type RenderError struct {
	Message string
	Node    parser.Node
}

func (e *RenderError) Error() string {
	return fmt.Sprintf("render error: %s", e.Message)
}

func (r *Renderer) Render() (string, error) {
	return r.renderNodes(r.AST)
}

func (r *Renderer) renderNodes(nodes []parser.Node) (string, error) {
	var sb strings.Builder
	for _, node := range nodes {
		rendered, err := r.renderNode(node)
		if err != nil {
			return "", err
		}
		sb.WriteString(rendered)
	}
	return sb.String(), nil
}

func (r *Renderer) renderNode(node parser.Node) (string, error) {
	switch node.Type {
	case parser.TEXT_NODE, parser.WHITESPACE_NODE:
		if node.Value == nil {
			return "", &RenderError{Message: "text node has nil value", Node: node}
		}
		return *node.Value, nil

	case parser.VARIABLE_NODE:
		if node.Value == nil {
			return "", &RenderError{Message: "variable node has nil value", Node: node}
		}
		variable, found := r.variableLookup(*node.Value)
		if !found {
			return "", &RenderError{
				Message: fmt.Sprintf("variable '%s' not found in context", *node.Value),
				Node:    node,
			}
		}
		return fmt.Sprintf("%v", variable), nil

	case parser.IF_NODE:
		return r.renderIfNode(node)

	case parser.FOR_NODE:
		return r.renderForNode(node)

	default:
		return "", &RenderError{
			Message: fmt.Sprintf("unknown node type: %v", node.Type),
			Node:    node,
		}
	}
}

func (r *Renderer) renderForNode(node parser.Node) (string, error) {
	var iteratee string
	var iterator []interface{}
	var sb strings.Builder

	for _, forNode := range node.Children {
		switch forNode.Type {
		case parser.ITERATEE_ITEM:
			if forNode.Value == nil {
				return "", &RenderError{Message: "iteratee item has nil value", Node: forNode}
			}
			iteratee = *forNode.Value

		case parser.ITERATOR_ITEM:
			if forNode.Value == nil {
				return "", &RenderError{Message: "iterator item has nil value", Node: forNode}
			}
			variable, found := r.variableLookup(*forNode.Value)
			if !found {
				return "", &RenderError{
					Message: fmt.Sprintf("iterator variable '%s' not found in context", *forNode.Value),
				}
			}
			var ok bool
			iterator, ok = variable.([]interface{})
			if !ok {
				return "", &RenderError{
					Message: fmt.Sprintf("iterator must be a slice, got %T", variable),
				}
			}
		}
	}

	for _, forNode := range node.Children {
		if forNode.Type == parser.FOR_BODY {
			// Store original value to restore after loop
			originalValue, hadOriginal := r.Context[iteratee]

			for _, item := range iterator {
				r.Context[iteratee] = item
				rendered, err := r.renderNodes(forNode.Children)
				if err != nil {
					return "", &RenderError{
						Message: fmt.Sprintf("error in for loop: %v", err),
						Node:    node,
					}
				}
				sb.WriteString(rendered)
			}

			// Restore original context
			if hadOriginal {
				r.Context[iteratee] = originalValue
			} else {
				delete(r.Context, iteratee)
			}
		}
	}

	return sb.String(), nil
}

// renderIfNode handles rendering if/elif/else conditional blocks
func (r *Renderer) renderIfNode(node parser.Node) (string, error) {
	if node.Value == nil {
		return "", &RenderError{Message: "if node has nil value", Node: node}
	}

	condition, err := r.evaluateCondition(*node.Value)
	if err != nil {
		return "", err
	}

	if condition {
		return r.renderConditionalBranch(node.Children, parser.THEN_BRANCH)
	}

	// Check elif branches
	if elifResult, err := r.renderElifBranches(node.Children); err != nil {
		return "", err
	} else if elifResult != "" {
		return elifResult, nil
	}

	// If no conditions matched, try else branch
	return r.renderConditionalBranch(node.Children, parser.ELSE_BRANCH)
}

// renderElifBranches handles rendering elif branches
func (r *Renderer) renderElifBranches(nodes []parser.Node) (string, error) {
	for _, node := range nodes {
		if node.Type != parser.ELIF_BRANCH {
			continue
		}

		for _, elifNode := range node.Children {
			if elifNode.Value == nil {
				return "", &RenderError{Message: "elif node has nil value", Node: elifNode}
			}

			condition, err := r.evaluateCondition(*elifNode.Value)
			if err != nil {
				return "", err
			}

			if condition {
				return r.renderNodes(elifNode.Children)
			}
		}
	}
	return "", nil
}

// renderConditionalBranch renders a specific branch (then/else) of a conditional
func (r *Renderer) renderConditionalBranch(nodes []parser.Node, branchType parser.NodeType) (string, error) {
	for _, node := range nodes {
		if node.Type == branchType {
			return r.renderNodes(node.Children)
		}
	}
	return "", nil
}

// evaluateCondition evaluates a boolean condition from the context
func (r *Renderer) evaluateCondition(key string) (bool, error) {
	value, exists := r.variableLookup(key)
	if !exists {
		return false, &RenderError{
			Message: fmt.Sprintf("condition variable '%s' not found in context", key),
		}
	}

	boolVal, ok := value.(bool)
	if !ok {
		return false, &RenderError{
			Message: fmt.Sprintf("condition variable '%s' is not a boolean", key),
		}
	}

	return boolVal, nil
}

// variableLookup retrieves a value from the context
func (r *Renderer) variableLookup(key string) (interface{}, bool) {
	value, exists := r.Context[key]
	return value, exists
}
