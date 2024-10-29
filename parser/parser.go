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
	EXPRESSION_NODE
	OP_EQUALS
	OP_NOT_EQUALS
	OP_AND
	OP_OR
	OP_LT
	OP_GT
	OP_LTE
	OP_GTE
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
		"EXPRESSION_NODE",
		"OP_EQUALS",
		"OP_NOT_EQUALS",
		"OP_AND",
		"OP_OR",
		"OP_LT",
		"OP_GT",
		"OP_LTE",
		"OP_GTE",
		"STRING_LITERAL_NODE",
		"NUMBER_LITERAL_NODE",
		"IF_NODE",
		"THEN_BRANCH",
		"ELIF_BRANCH",
		"ELIF_ITEM",
		"ELSE_BRANCH",
		"FOR_NODE",
		"ITERATOR_ITEM",
		"ITERATEE_ITEM",
		"FOR_BODY",
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

func NewIfNode(condition *string, thenBranch, elifBranch, elseBranch Node) Node {
	var children []Node

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
		Value:    condition,
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
			} else if p.match(lexer.IDENTIFIER) {
				varNode, err := p.parseVar()
				if err != nil {
					return nil, fmt.Errorf("error parsing variable statement: %w", err)
				}
				nodes = append(nodes, varNode)
			} else {
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unrecognized token: %v, they should start with -> '{{'", p.peek())
		}
	}
}

func (p *Parser) parseVar() (Node, error) {
	identifier := p.previous().Value
	// If there are no expression in the curlies, it's an variable node we can bail.
	if p.check(lexer.CLOSE_CURLY) {
		p.advance() // Consume the closing curly
		return Node{Type: VARIABLE_NODE, Value: &identifier}, nil
	}

	var expressionNodes []Node
	expressionNodes = append(expressionNodes, Node{Type: VARIABLE_NODE, Value: &identifier})

	for !p.check(lexer.CLOSE_CURLY) {
		currToken := p.peek()
		if operator, exists := lexer.Operators[currToken.Value]; exists {
			p.advance()
			switch operator {
			case lexer.AMPERSAND:
				expressionNodes = append(expressionNodes, Node{Type: OP_AND, Value: &currToken.Value})
			case lexer.PIPE:
				expressionNodes = append(expressionNodes, Node{Type: OP_OR, Value: &currToken.Value})
			case lexer.EQ:
				expressionNodes = append(expressionNodes, Node{Type: OP_EQUALS, Value: &currToken.Value})
			case lexer.NEQ:
				expressionNodes = append(expressionNodes, Node{Type: OP_NOT_EQUALS, Value: &currToken.Value})
			case lexer.GT:
				expressionNodes = append(expressionNodes, Node{Type: OP_GT, Value: &currToken.Value})
			case lexer.LT:
				expressionNodes = append(expressionNodes, Node{Type: OP_LT, Value: &currToken.Value})
			case lexer.GTE:
				expressionNodes = append(expressionNodes, Node{Type: OP_GTE, Value: &currToken.Value})
			case lexer.LTE:
				expressionNodes = append(expressionNodes, Node{Type: OP_LTE, Value: &currToken.Value})
			}
			continue
		}

		p.advance()
		currToken = p.previous()

		switch currToken.Type {
		case lexer.STRING:
			str := strings.Trim(currToken.Value, "'")
			expressionNodes = append(expressionNodes, Node{Type: STRING_LITERAL_NODE, Value: &str})
		case lexer.NUMBER:
			expressionNodes = append(expressionNodes, Node{Type: NUMBER_LITERAL_NODE, Value: &currToken.Value})
		case lexer.IDENTIFIER:
			expressionNodes = append(expressionNodes, Node{Type: VARIABLE_NODE, Value: &currToken.Value})
		}
	}

	p.advance() // Consume the closing curly
	return Node{Type: EXPRESSION_NODE, Children: expressionNodes}, nil
}

// parseIf parses an if-else construct in the template.
// It handles:
//  1. The condition of the if statement
//  2. The 'then' block, which may contain nested templates
//  3. An optional 'else' block, also potentially containing nested templates
//  4. The 'endif' terminator
//
// Each block is parsed as a separate template, allowing for nested if-else constructs.
// Returns a Node representing the entire if-else structure.
func (p *Parser) parseIf() (Node, error) {
	condition, err := p.expectIfIdentifier()
	if err != nil {
		return Node{}, err
	}

	if err := p.expectCloseCurly(); err != nil {
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

	return NewIfNode(&condition,
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
	p.advance() // {{
	p.advance() // elif

	condition, err := p.expectIfIdentifier()
	if err != nil {
		return Node{}, err
	}

	if err := p.expectCloseCurly(); err != nil {
		return Node{}, err
	}

	block, err := p.parseBlock()
	if err != nil {
		return Node{}, fmt.Errorf("error parsing elif block: %w", err)
	}

	return NewNode(ELIF_ITEM, &condition, block...), nil
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
