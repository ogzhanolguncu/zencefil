package parser

import (
	"fmt"

	"github.com/ogzhanolguncu/zencefil/lexer"
)

type NodeType int

const (
	TextNode NodeType = iota
	VariableNode
	IfNode
	ForNode
)

func (tt NodeType) String() string {
	return [...]string{"TextNode", "VariableNode", "IfNode", "ForNode"}[tt]
}

type Node struct {
	Type     NodeType
	Value    string
	Children []Node
}

func NewTextNode(text string) Node {
	return Node{Type: TextNode, Value: text}
}

func NewVariableNode(name string) Node {
	return Node{Type: VariableNode, Value: name}
}

func NewIfNode(condition string, thenBranch, elseBranch []Node) Node {
	return Node{
		Type:     IfNode,
		Value:    condition,
		Children: append(thenBranch, elseBranch...),
	}
}

type Parser struct {
	currentPosition int
	Tokens          []lexer.Token
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{
		currentPosition: 0,
		Tokens:          tokens,
	}
}

// {Type: TEXT, Value: "Hello, "},
// {Type: OPEN_CURLY, Value: "{{"},
// {Type: KEYWORD, Value: "if"},
// {Type: IDENTIFIER, Value: "is_admin"},
// {Type: CLOSE_CURLY, Value: "}}"},
// {Type: TEXT, Value: " You are an admin."},
// {Type: OPEN_CURLY, Value: "{{"},
// {Type: KEYWORD, Value: "endif"},
// {Type: CLOSE_CURLY, Value: "}}"},

// Hello, {{ if is_admin }} You are an admin. {{ else }} But not a super admin. {{ endif }}
func (p *Parser) Parse() ([]Node, error) {
	var nodes []Node

	for !p.isAtEnd() {
		if p.match(lexer.TEXT) {
			nodes = append(nodes, NewTextNode(p.previous().Value))
		} else if p.match(lexer.OPEN_CURLY) {
			if p.match(lexer.KEYWORD) && p.previous().Value == "if" {
				ifNode, err := p.parseIf()
				if err != nil {
					return nil, fmt.Errorf("error parsing if statement: %w", err)
				}
				nodes = append(nodes, ifNode)
			} else {
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unexpected token: %v", p.peek())
		}
	}

	return nodes, nil
}

func (p *Parser) parseIf() (Node, error) {
	if !p.match(lexer.IDENTIFIER) {
		return Node{}, fmt.Errorf("expected condition after 'if', got %v", p.peek())
	}
	condition := p.previous().Value

	if !p.match(lexer.CLOSE_CURLY) {
		return Node{}, fmt.Errorf("expected '}}' after '%s', got %v", condition, p.peek())
	}

	thenBlock, err := p.parseBlock()
	if err != nil {
		return Node{}, fmt.Errorf("error parsing then block: %w", err)
	}

	var elseBlock []Node
	if p.match(lexer.OPEN_CURLY) && p.match(lexer.KEYWORD) && p.previous().Value == "else" {
		if !p.match(lexer.CLOSE_CURLY) {
			return Node{}, fmt.Errorf("expected '}}' after else")
		}
		elseBlock, err = p.parseBlock()
		if err != nil {
			return Node{}, fmt.Errorf("error parsing else block: %w", err)
		}
	} else {
		// We have to go two step back if 'else' is missing
		p.backup()
		p.backup()
	}

	// Check for endif
	if !p.match(lexer.OPEN_CURLY) {
		return Node{}, fmt.Errorf("expected '{{' before 'endif', got: %v", p.peek())
	}
	if !p.match(lexer.KEYWORD) || p.previous().Value != "endif" {
		return Node{}, fmt.Errorf("expected 'endif' to close if statement, got: %v", p.previous())
	}
	if !p.match(lexer.CLOSE_CURLY) {
		return Node{}, fmt.Errorf("expected '}}' after 'endif', got: %v", p.peek())
	}

	return NewIfNode(condition, thenBlock, elseBlock), nil
}

// Hello, {{ if is_admin }} You are an admin. {{ else }} But not a super admin. {{ endif }}

func (p *Parser) parseBlock() ([]Node, error) {
	var nodes []Node

	for !p.isAtEnd() &&
		!(p.check(lexer.OPEN_CURLY) &&
			p.checkNext(lexer.KEYWORD) &&
			(p.Tokens[p.currentPosition+1].Value == "else" || p.Tokens[p.currentPosition+1].Value == "endif")) {

		if p.match(lexer.TEXT) {
			nodes = append(nodes, NewTextNode(p.previous().Value))
		} else if p.match(lexer.OPEN_CURLY) {
			if p.match(lexer.KEYWORD) && (p.previous().Value == "if") {
				ifNode, err := p.parseIf()
				if err != nil {
					return nil, fmt.Errorf("error parsing if statement: %w", err)
				}
				nodes = append(nodes, ifNode)
			} else {
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unexpected token: %v", p.peek())
		}
	}
	return nodes, nil
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.Tokens[p.currentPosition].Type == t
}

func (p *Parser) checkNext(t lexer.TokenType) bool {
	if p.currentPosition+1 >= len(p.Tokens) {
		return false
	}
	return p.Tokens[p.currentPosition+1].Type == t
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.currentPosition++
	}
	return p.previous()
}

func (p *Parser) backup() {
	if !p.isAtBeginning() {
		p.currentPosition--
	}
}

func (p *Parser) previous() lexer.Token {
	return p.Tokens[p.currentPosition-1]
}

func (p *Parser) isAtEnd() bool {
	return p.currentPosition >= len(p.Tokens)
}
func (p *Parser) isAtBeginning() bool {
	return p.currentPosition == 0
}

// Render function to use the parsed nodes
func (p *Parser) peek() lexer.Token {
	if p.isAtEnd() {
		return lexer.Token{Type: -1, Value: "EOF"}
	}
	return p.Tokens[p.currentPosition]
}
