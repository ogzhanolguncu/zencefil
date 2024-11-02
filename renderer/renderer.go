package renderer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ogzhanolguncu/zencefil/parser"
)

// TODO: Object access
// TODO: iterating over object in for
// TODO: expression evaluation in variable
func hasHigherPrecedence(op1, op2 parser.NodeType) bool {
	precedence := map[parser.NodeType]int{
		parser.OP_BANG:       5,
		parser.OP_EQUALS:     4,
		parser.OP_NOT_EQUALS: 4,
		parser.OP_GT:         4,
		parser.OP_LT:         4,
		parser.OP_GTE:        4,
		parser.OP_LTE:        4,
		parser.OP_AND:        2,
		parser.OP_OR:         1,
	}
	return precedence[op1] > precedence[op2]
}

var operatorStringMap = map[parser.NodeType]string{
	parser.OP_AND:           "&&",
	parser.OP_OR:            "||",
	parser.OP_EQUALS:        "==",
	parser.OP_NOT_EQUALS:    "!=",
	parser.OP_GT:            ">",
	parser.OP_LT:            "<",
	parser.OP_GTE:           ">=",
	parser.OP_LTE:           "<=",
	parser.OP_BANG:          "!",
	parser.OP_NULL_COALESCE: "??",
}

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

	case parser.EXPRESSION_NODE:
		expr, err := r.evaluateExpression(node)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", expr), nil

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
	conditionNode := node.Children[0]
	if conditionNode.Type != parser.VARIABLE_NODE && conditionNode.Type != parser.EXPRESSION_NODE {
		return "", &RenderError{Message: "if node has nil condition", Node: node}
	}

	if conditionNode.Type == parser.VARIABLE_NODE {
		condition, err := r.evaluateCondition(*conditionNode.Value)
		if err != nil {
			return "", err
		}
		if condition {
			return r.renderConditionalBranch(node.Children, parser.THEN_BRANCH)
		}
	}

	if conditionNode.Type == parser.EXPRESSION_NODE {
		condition, err := r.evaluateExpression(conditionNode)
		if err != nil {
			return "", err
		}
		if isTruthy(condition) {
			return r.renderConditionalBranch(node.Children, parser.THEN_BRANCH)
		}
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
			// First element in the 'if', 'elif' nodes is condition node.
			// But, we have to delete the first item from node children to evaluate the rest of it.
			// Otherwise when we call 'renderNode', the condition will be treated as 'VARIABLE_NODE'.
			conditionNode := elifNode.Children[0]
			elifNode.Children = elifNode.Children[1:]

			if conditionNode.Type != parser.VARIABLE_NODE && conditionNode.Type != parser.EXPRESSION_NODE {
				return "", &RenderError{Message: "elif node has nil condition", Node: node}
			}

			if conditionNode.Type == parser.VARIABLE_NODE {
				condition, err := r.evaluateCondition(*conditionNode.Value)
				if err != nil {
					return "", err
				}

				if condition {
					return r.renderNodes(elifNode.Children)
				}
			}

			if conditionNode.Type == parser.EXPRESSION_NODE {
				condition, err := r.evaluateExpression(conditionNode)
				if err != nil {
					return "", err
				}

				if isTruthy(condition) {
					return r.renderNodes(elifNode.Children)
				}
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

func (r *Renderer) evaluateExpression(node parser.Node) (interface{}, error) {
	var operandStack []interface{}
	var operatorStack []parser.NodeType

	// Process nodes in original order
	for i := 0; i < len(node.Children); i++ {
		v := node.Children[i]

		switch v.Type {
		case parser.VARIABLE_NODE:
			if v.Value == nil {
				return false, fmt.Errorf("variable node has nil value")
			}
			value, exists := r.variableLookup(*v.Value)
			if !exists {
				return false, &RenderError{
					Message: fmt.Sprintf("variable '%s' not found in context", *v.Value),
					Node:    v,
				}
			}
			operandStack = append(operandStack, value)
			applyPendingBang(&operandStack, &operatorStack)

		case parser.EXPRESSION_NODE:
			value, err := r.evaluateExpression(v)
			if err != nil {
				return false, fmt.Errorf("failed to evaluate nested expression: %w", err)
			}
			operandStack = append(operandStack, value)
			applyPendingBang(&operandStack, &operatorStack)

		case parser.STRING_LITERAL_NODE:
			operandStack = append(operandStack, *v.Value)
			applyPendingBang(&operandStack, &operatorStack)

		case parser.NUMBER_LITERAL_NODE:
			if num, err := strconv.ParseFloat(*v.Value, 64); err == nil {
				operandStack = append(operandStack, num)
			} else {
				return false, fmt.Errorf("invalid number literal: %s", *v.Value)
			}
			applyPendingBang(&operandStack, &operatorStack)

		case parser.OP_BANG:
			operatorStack = append(operatorStack, parser.OP_BANG)

		case parser.OP_AND, parser.OP_OR, parser.OP_EQUALS, parser.OP_NOT_EQUALS,
			parser.OP_GT, parser.OP_LT, parser.OP_GTE, parser.OP_LTE:
			// Evaluate immediately if operator has higher or equal precedence
			for len(operatorStack) > 0 && hasHigherPrecedence(operatorStack[len(operatorStack)-1], v.Type) {
				err := evaluateTopOperator(&operandStack, &operatorStack)
				if err != nil {
					return false, err
				}
			}
			operatorStack = append(operatorStack, v.Type)
		}
	}

	// Evaluate remaining operators
	for len(operatorStack) > 0 {
		if err := evaluateTopOperator(&operandStack, &operatorStack); err != nil {
			return false, err
		}
	}

	if len(operandStack) != 1 {
		return false, fmt.Errorf("invalid expression: expected 1 final result, got %d", len(operandStack))
	}

	return operandStack[0], nil
}

func (r *Renderer) variableLookup(key string) (interface{}, bool) {
	value, exists := r.Context[key]
	return value, exists
}

// HELPERS
func isTruthy(v interface{}) bool {
	switch v := v.(type) {
	case nil:
		return false
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case float64:
		return v != 0
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}

// Helper function to convert various numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

func compareValues(a, b interface{}) int {
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)

	// If either is a string, do direct string comparison
	if aIsStr && bIsStr {
		if aStr == bStr {
			return 0
		}
		if aStr < bStr {
			return -1
		}
		return 1
	}

	// Try to convert both values to float64 for numeric comparison
	var aFloat, bFloat float64
	var aOk, bOk bool

	switch v := a.(type) {
	case float64:
		aFloat, aOk = v, true
	case int:
		aFloat, aOk = float64(v), true
	}

	switch v := b.(type) {
	case float64:
		bFloat, bOk = v, true
	case int:
		bFloat, bOk = float64(v), true
	}

	if aOk && bOk {
		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0
	}

	// For boolean values
	aBool, aIsBool := a.(bool)
	bBool, bIsBool := b.(bool)
	if aIsBool && bIsBool {
		if aBool == bBool {
			return 0
		}
		if aBool {
			return 1
		}
		return -1
	}

	// If all the comparisons fail treat them as strings and compare
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

func applyPendingBang(operandStack *[]interface{}, operatorStack *[]parser.NodeType) {
	if len(*operatorStack) > 0 && (*operatorStack)[len(*operatorStack)-1] == parser.OP_BANG {
		*operatorStack = (*operatorStack)[:len(*operatorStack)-1]
		lastIdx := len(*operandStack) - 1
		(*operandStack)[lastIdx] = !isTruthy((*operandStack)[lastIdx])
	}
}

func evaluateTopOperator(operandStack *[]interface{}, operatorStack *[]parser.NodeType) error {
	if len(*operatorStack) < 1 {
		return fmt.Errorf("invalid expression: no operator")
	}

	op := (*operatorStack)[len(*operatorStack)-1]
	*operatorStack = (*operatorStack)[:len(*operatorStack)-1]

	// Handle unary NOT operator
	if op == parser.OP_BANG {
		if len(*operandStack) < 1 {
			return fmt.Errorf("invalid expression: not enough operands for NOT operator")
		}
		lastIdx := len(*operandStack) - 1
		(*operandStack)[lastIdx] = !isTruthy((*operandStack)[lastIdx])
		return nil
	}

	// Handle binary operators
	if len(*operandStack) < 2 {
		return fmt.Errorf("invalid expression: not enough operands")
	}

	right := (*operandStack)[len(*operandStack)-1]
	left := (*operandStack)[len(*operandStack)-2]
	*operandStack = (*operandStack)[:len(*operandStack)-2]

	var result interface{}

	switch op {
	case parser.OP_AND:
		// If left is falsy, return left, otherwise return right
		if !isTruthy(left) {
			result = left
		} else {
			result = right
		}
	case parser.OP_OR:
		// If left is truthy, return left, otherwise return right
		if isTruthy(left) {
			result = left
		} else {
			result = right
		}
	case parser.OP_EQUALS:
		result = compareValues(left, right) == 0
	case parser.OP_NOT_EQUALS:
		result = compareValues(left, right) != 0
	case parser.OP_GT:
		result = compareValues(left, right) > 0
	case parser.OP_LT:
		result = compareValues(left, right) < 0
	case parser.OP_GTE:
		result = compareValues(left, right) >= 0
	case parser.OP_LTE:
		result = compareValues(left, right) <= 0
	default:
		return fmt.Errorf("unsupported operator: %v", op)
	}

	*operandStack = append(*operandStack, result)
	return nil
}
