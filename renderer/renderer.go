package renderer

import (
	"fmt"
	"strings"

	"github.com/ogzhanolguncu/zencefil/parser"
)

// Renderer handles template rendering using an AST and context
type Renderer struct {
	Context map[string]interface{}
	AST     []parser.Node
}

// New creates a new Renderer instance
func New(ast []parser.Node, context map[string]interface{}) *Renderer {
	if context == nil {
		context = make(map[string]interface{})
	}
	return &Renderer{
		Context: context,
		AST:     ast,
	}
}

// RenderError represents a template rendering error
type RenderError struct {
	Message string
	Node    parser.Node
}

func (e *RenderError) Error() string {
	return fmt.Sprintf("render error: %s", e.Message)
}

// Render processes the AST and returns the rendered template
func (r *Renderer) Render() (string, error) {
	return r.renderNodes(r.AST)
}

// renderNodes handles rendering a slice of nodes
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

// renderNode handles rendering a single node
func (r *Renderer) renderNode(node parser.Node) (string, error) {
	switch node.Type {
	case parser.TEXT_NODE:
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

	default:
		return "", &RenderError{
			Message: fmt.Sprintf("unknown node type: %v", node.Type),
			Node:    node,
		}
	}
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
