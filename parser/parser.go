package parser

import (
	"fmt"
	"strings"

	"github.com/ogzhanolguncu/zencefil/lexer"
)

type NodeType int

const (
	TEXT_NODE NodeType = iota
	VARIABLE_NODE
	OBJECT_ACCESS_NODE
	OBJECT_ACCESOR
	EXPRESSION_NODE
	OP_EQUALS
	OP_NOT_EQUALS
	OP_AND
	OP_OR
	OP_LT
	OP_GT
	OP_LTE
	OP_GTE
	OP_BANG
	OP_NULL_COALESCE
	RPAREN
	LPAREN
	OPEN_BRACKET
	CLOSE_BRACKET
	STRING_LITERAL_NODE
	NUMBER_LITERAL_NODE
	IF_NODE
	THEN_BRANCH
	ELIF_BRANCH
	ELIF_ITEM
	ELSE_BRANCH
	FOR_NODE
	ITERATOR_ITEM
	ITERATEE_ITEM
	FOR_BODY
)

func (tt NodeType) String() string {
	return [...]string{
		"TEXT_NODE",
		"VARIABLE_NODE",
		"OBJECT_ACCESS_NODE", "OBJECT_ACCESOR",
		"EXPRESSION_NODE",
		"OP_EQUALS", "OP_NOT_EQUALS",
		"OP_AND", "OP_OR",
		"OP_LT", "OP_GT", "OP_LTE", "OP_GTE",
		"OP_BANG",
		"OP_NULL_COALESCE",
		"RPAREN", "LPAREN",
		"OPEN_BRACKET", "CLOSE_BRACKET",
		"STRING_LITERAL_NODE", "NUMBER_LITERAL_NODE",
		"IF_NODE", "THEN_BRANCH", "ELIF_BRANCH", "ELIF_ITEM", "ELSE_BRANCH",
		"FOR_NODE", "ITERATOR_ITEM", "ITERATEE_ITEM", "FOR_BODY",
	}[tt]
}

type Node struct {
	Value    *string
	Children []Node
	Type     NodeType
}

func NewNode(nodeType NodeType, value *string, children ...Node) Node {
	return Node{
		Type:     nodeType,
		Value:    value,
		Children: children,
	}
}

func NewIfNode(condition, thenBranch, elifBranch, elseBranch Node) Node {
	var children []Node

	// Condition is either variable or expression node
	children = append(children, condition)

	// Add THEN_BRANCH even if empty
	children = append(children, NewNode(THEN_BRANCH, nil, thenBranch.Children...))

	// Add ELIF_BRANCH if it has children
	if len(elifBranch.Children) > 0 {
		children = append(children, elifBranch)
	}

	// Add ELSE_BRANCH even if empty
	if len(elseBranch.Children) > 0 {
		children = append(children, NewNode(ELSE_BRANCH, nil, elseBranch.Children...))
	}

	return Node{
		Type:     IF_NODE,
		Children: children,
	}
}

func NewForNode(iterator, iteratee, body Node) Node {
	return Node{
		Type:     FOR_NODE,
		Children: []Node{iteratee, iterator, body},
	}
}

type Parser struct {
	tokens []lexer.Token
	crrPos int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() ([]Node, error) {
	var nodes []Node

	for {
		if p.isAtEnd() {
			return nodes, nil
		}

		if p.isBlockEnd() {
			return nil, fmt.Errorf("malformed tokens. 'else' or 'endif' cannot be used without 'if'")
		}

		if p.match(lexer.TEXT) {
			prevVal := p.previous().Value
			nodes = append(nodes, NewNode(TEXT_NODE, &prevVal))
		} else if p.match(lexer.OPEN_CURLY) {
			if p.match(lexer.KEYWORD) {
				prevVal := p.previous().Value
				switch prevVal {
				case "if":
					IfNode, err := p.parseIf()
					if err != nil {
						return nil, fmt.Errorf("error parsing if statement: %w", err)
					}
					nodes = append(nodes, IfNode)
				case "for":
					forNode, err := p.parseFor()
					if err != nil {
						return nil, fmt.Errorf("error parsing for statement: %w", err)
					}
					nodes = append(nodes, forNode)
				default:
					panic("Unknown KEYWORD")
				}
			} else if p.check(lexer.IDENTIFIER) || p.check(lexer.LPAREN) || p.check(lexer.BANG) {
				exprNode, err := p.parseExpression()
				if err != nil {
					return nil, fmt.Errorf("error parsing expression: %w", err)
				}
				nodes = append(nodes, exprNode)
			} else {
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unrecognized token: %v, they should start with -> '{{'", p.peek())
		}
	}
}

func (p *Parser) parseExpression() (Node, error) {
	var nodes []Node
	for !p.isAtEnd() {
		if p.check(lexer.CLOSE_CURLY) {
			p.advance() // consume closing curly
			break
		}

		switch p.peek().Type {
		case lexer.LPAREN:
			p.advance() // consume '('
			nestedExpr, err := p.parseExpression()
			if err != nil {
				return Node{}, err
			}
			nodes = append(nodes, nestedExpr)

		case lexer.RPAREN:
			p.advance() // consume ')'
			return Node{Type: EXPRESSION_NODE, Children: nodes}, nil

		case lexer.BANG:
			p.advance()
			val := p.previous().Value
			bangNode := Node{Type: OP_BANG, Value: &val}

			if p.check(lexer.LPAREN) {
				p.advance() // consume '('
				nestedExpr, err := p.parseExpression()
				if err != nil {
					return Node{}, err
				}
				nodes = append(nodes, bangNode, nestedExpr)
			} else if p.match(lexer.IDENTIFIER) {
				val := p.previous().Value
				nodes = append(nodes, bangNode, Node{Type: VARIABLE_NODE, Value: &val})
			}

		case lexer.IDENTIFIER:
			p.advance()
			val := p.previous().Value
			if p.check(lexer.OPEN_BRACKET) {
				p.advance() // Consume '['
				p.advance() // Consume 'string' token for objAccessor
				objAccessor := p.previous()
				if objAccessor.Type != lexer.STRING {
					return Node{}, fmt.Errorf("object accessor has to be STRING token, but its %v", objAccessor.Type)
				}
				objNode := Node{Type: OBJECT_ACCESS_NODE}
				objNode.Children = []Node{{Type: VARIABLE_NODE, Value: &val}, {Type: OBJECT_ACCESOR, Value: &objAccessor.Value}}
				p.advance() // Consume ']'

				nodes = append(nodes, objNode)
			} else {
				nodes = append(nodes, Node{Type: VARIABLE_NODE, Value: &val})
			}

		case lexer.STRING:
			p.advance()
			val := strings.Trim(p.previous().Value, "'")
			nodes = append(nodes, Node{Type: STRING_LITERAL_NODE, Value: &val})
		case lexer.NUMBER:
			p.advance()
			val := p.previous().Value
			nodes = append(nodes, Node{Type: NUMBER_LITERAL_NODE, Value: &val})

		default:
			// Check for operators
			if operator, exists := lexer.Operators[p.peek().Value]; exists {
				p.advance()
				val := p.previous().Value
				nodes = append(nodes, p.createOperatorNode(operator, &val))
			} else {
				return Node{}, fmt.Errorf("unexpected token in expression: %v", p.peek())
			}
		}
	}

	// If we only have one node and it's already an expression, return it directly
	if len(nodes) == 1 && nodes[0].Type == EXPRESSION_NODE {
		return nodes[0], nil
	}

	// If we have a single node that's not an operator, return it directly
	if len(nodes) == 1 && !isOperator(nodes[0].Type) {
		return nodes[0], nil
	}

	return Node{Type: EXPRESSION_NODE, Children: nodes}, nil
}

func isOperator(nodeType NodeType) bool {
	switch nodeType {
	case OP_EQUALS, OP_NOT_EQUALS, OP_AND, OP_OR, OP_LT, OP_GT,
		OP_LTE, OP_GTE, OP_BANG, OP_NULL_COALESCE:
		return true
	default:
		return false
	}
}

func (p *Parser) createOperatorNode(op lexer.TokenType, value *string) Node {
	opTypeMap := map[lexer.TokenType]NodeType{
		lexer.AMPERSAND:     OP_AND,
		lexer.PIPE:          OP_OR,
		lexer.EQ:            OP_EQUALS,
		lexer.NEQ:           OP_NOT_EQUALS,
		lexer.GT:            OP_GT,
		lexer.LT:            OP_LT,
		lexer.GTE:           OP_GTE,
		lexer.LTE:           OP_LTE,
		lexer.BANG:          OP_BANG,
		lexer.NULL_COALESCE: OP_NULL_COALESCE,
	}
	return Node{Type: opTypeMap[op], Value: value}
}

func (p *Parser) parseCondOrExpr() (Node, error) {
	identifier := p.previous().Value
	// If there are no expression in the curlies, it's an variable node so we can bail.
	if p.check(lexer.CLOSE_CURLY) {
		p.advance() // Consume the closing curly
		return Node{Type: VARIABLE_NODE, Value: &identifier}, nil
	}

	expr, err := p.parseExpression()
	if err != nil {
		return Node{}, err
	}
	return expr, nil
}

func (p *Parser) parseIf() (Node, error) {
	condition, err := p.parseCondOrExpr()
	if err != nil {
		return Node{}, err
	}

	thenBlock, err := p.parseBlock()
	if err != nil {
		return Node{}, fmt.Errorf("error parsing then block: %w", err)
	}

	var elifNodes []Node
	for p.isElifKeyword() {
		elifNode, err := p.parseElif()
		if err != nil {
			return Node{}, fmt.Errorf("error parsing elif block: %w", err)
		}
		elifNodes = append(elifNodes, elifNode)
	}

	var elseBlock Node
	if p.isElseKeyword() {
		elseBlock, err = p.parseElse()
		if err != nil {
			return Node{}, fmt.Errorf("error parsing else block: %w", err)
		}
	}

	if err := p.expectAndConsumeEndIf(); err != nil {
		return Node{}, err
	}

	return NewIfNode(condition,
		NewNode(THEN_BRANCH, nil, thenBlock...),
		NewNode(ELIF_BRANCH, nil, elifNodes...),
		elseBlock), nil
}

func (p *Parser) parseFor() (Node, error) {
	iteratee, err := p.expectForIteratee()
	iterateeNode := Node{Type: ITERATEE_ITEM, Value: &iteratee}
	if err != nil {
		return Node{}, err
	}

	if err := p.expectInKeyword(); err != nil {
		return Node{}, err
	}

	iterator, err := p.expectForIterator()
	iteratorNode := Node{Type: ITERATOR_ITEM, Value: &iterator}
	if err != nil {
		return Node{}, err
	}

	if err := p.expectCloseCurly(); err != nil {
		return Node{}, err
	}

	body, err := p.parseBlock()
	forBody := Node{Type: FOR_BODY, Children: body}
	if err != nil {
		return Node{}, fmt.Errorf("error parsing for body: %w", err)
	}

	if err := p.expectAndConsumeEndFor(); err != nil {
		return Node{}, err
	}

	return NewForNode(iteratorNode, iterateeNode, forBody), nil
}

func (p *Parser) parseElse() (Node, error) {
	p.advance()
	p.advance()
	if err := p.expectCloseCurly(); err != nil {
		return Node{}, err
	}
	elseBlock, err := p.parseBlock()
	if err != nil {
		return Node{}, err
	}
	return NewNode(ELSE_BRANCH, nil, elseBlock...), nil
}

func (p *Parser) parseElif() (Node, error) {
	p.advance() // consume {{
	p.advance() // consume elif

	var nodes []Node
	condition, err := p.parseCondOrExpr()
	if err != nil {
		return Node{}, err
	}
	nodes = append(nodes, condition)

	block, err := p.parseBlock()
	if err != nil {
		return Node{}, fmt.Errorf("error parsing elif block: %w", err)
	}
	nodes = append(nodes, block...)

	return NewNode(ELIF_ITEM, nil, nodes...), nil
}

func (p *Parser) parseBlock() ([]Node, error) {
	var nodes []Node

	for !p.isAtEnd() && !p.isBlockEnd() {
		if p.match(lexer.TEXT) {
			prevVal := p.previous().Value
			nodes = append(nodes, NewNode(TEXT_NODE, &prevVal))
		} else if p.match(lexer.OPEN_CURLY) {
			if p.match(lexer.KEYWORD) {
				switch p.previous().Value {
				case "if":
					ifNode, err := p.parseIf()
					if err != nil {
						return nil, fmt.Errorf("error parsing nested if statement: %w", err)
					}
					nodes = append(nodes, ifNode)
				case "for":
					forNode, err := p.parseFor()
					if err != nil {
						return nil, fmt.Errorf("error parsing nested for statement: %w", err)
					}
					nodes = append(nodes, forNode)
				default:
					panic("Unknown KEYWORD")
				}
			} else if p.match(lexer.IDENTIFIER) {
				prevVal := p.previous().Value
				nodes = append(nodes, NewNode(VARIABLE_NODE, &prevVal))
				p.advance() // consume '}}' of variable node
			} else {
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unexpected token: %v", p.peek())
		}
	}
	return nodes, nil
}

func (p *Parser) isBlockEnd() bool {
	return p.isElseKeyword() || p.isElifKeyword() || p.isEndIfKeyword() || p.isEndForKeyword()
}

func (p *Parser) isElifKeyword() bool {
	return p.check(lexer.OPEN_CURLY) && p.checkNext(lexer.KEYWORD) && p.tokens[p.crrPos+1].Value == "elif"
}

func (p *Parser) isElseKeyword() bool {
	return p.check(lexer.OPEN_CURLY) && p.checkNext(lexer.KEYWORD) && p.tokens[p.crrPos+1].Value == "else"
}

func (p *Parser) isEndIfKeyword() bool {
	return p.check(lexer.OPEN_CURLY) && p.checkNext(lexer.KEYWORD) && p.tokens[p.crrPos+1].Value == "endif"
}

func (p *Parser) isEndForKeyword() bool {
	return p.check(lexer.OPEN_CURLY) && p.checkNext(lexer.KEYWORD) && p.tokens[p.crrPos+1].Value == "endfor"
}

func (p *Parser) expectIfIdentifier() (string, error) {
	if !p.match(lexer.IDENTIFIER) {
		return "", fmt.Errorf("expected condition after 'if', got %v", p.peek())
	}
	return p.previous().Value, nil
}

func (p *Parser) expectElifIdentifier() (string, error) {
	if !p.match(lexer.IDENTIFIER) {
		return "", fmt.Errorf("expected condition after 'elif', got %v", p.peek())
	}
	return p.previous().Value, nil
}

func (p *Parser) expectForIteratee() (string, error) {
	if !p.match(lexer.IDENTIFIER) {
		return "", fmt.Errorf("expected iteratee after 'for', got %v", p.peek())
	}
	return p.previous().Value, nil
}

func (p *Parser) expectForIterator() (string, error) {
	if !p.match(lexer.IDENTIFIER) {
		return "", fmt.Errorf("expected iterator after 'in', got %v", p.peek())
	}
	return p.previous().Value, nil
}

func (p *Parser) expectAndConsumeEndIf() error {
	if !p.isEndIfKeyword() {
		return fmt.Errorf("expected '{{ endif }}' to close if statement, got: %v", p.peek())
	}
	p.advance() // {{
	p.advance() // endif
	return p.expectCloseCurly()
}

func (p *Parser) expectAndConsumeEndFor() error {
	if !p.isEndForKeyword() {
		return fmt.Errorf("expected '{{ endfor }}' to close for statement, got: %v", p.peek())
	}
	p.advance() // {{
	p.advance() // endif
	return p.expectCloseCurly()
}

func (p *Parser) expectCloseCurly() error {
	if !p.match(lexer.CLOSE_CURLY) {
		return fmt.Errorf("expected '}}', got %v", p.peek())
	}
	return nil
}

func (p *Parser) expectInKeyword() error {
	if p.match(lexer.KEYWORD) && p.previous().Value != "in" {
		return fmt.Errorf("expected 'in', got %v", p.peek())
	}
	return nil
}

// -------- HELPERS --------

// If given type matches current token, we consume it and move forward
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

// Consumes one token
func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.crrPos++
	}
	return p.previous()
}

// Similar to peek, but gives last token instead
func (p *Parser) previous() lexer.Token {
	return p.tokens[p.crrPos-1]
}

// If we are not at the end return next tokens type
func (p *Parser) checkNext(t lexer.TokenType) bool {
	return p.crrPos+1 < len(p.tokens) && p.tokens[p.crrPos+1].Type == t
}

// Check returns true if the current token matches the given type
func (p *Parser) check(t lexer.TokenType) bool {
	return !p.isAtEnd() && p.peek().Type == t
}

// Checks current token without consuming it
func (p *Parser) peek() lexer.Token {
	if p.isAtEnd() {
		return lexer.Token{Type: -1, Value: "EOF"}
	}
	return p.tokens[p.crrPos]
}

// Checks if we are at the end of token list
func (p *Parser) isAtEnd() bool {
	return p.crrPos >= len(p.tokens)
}
