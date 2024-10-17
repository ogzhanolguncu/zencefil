package parser

import (
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

// Hello, {{ name }}! {{ if is_admin }} You are an admin. {{ elseif is_customer}} You are a customer {{ else }} But not a super admin. {{ endif }}
func (p *Parser) Parse(bail string) []Node {

	var nodes []Node
	for !p.isAtEnd() {
		token := p.peek()
		switch token.Type {
		case lexer.TEXT:
			nodes = append(nodes, NewTextNode(p.advance().Value))
		case lexer.OPEN_CURLY:
			p.advance() // consume '{'
			token = p.advance()
			switch token.Type {
			case lexer.IDENTIFIER:
				nodes = append(nodes, NewVariableNode(token.Value))
				p.expectCloseCurly()
			case lexer.KEYWORD:
				if token.Value == "if" {
					nodes = append(nodes, p.parseIfStatement())
				} else if token.Value == "else" {
					return nodes
				} else if token.Value == "endif" {
					return nodes
				}
			}
		}
	}
	return nodes
}

func (p *Parser) parseIfStatement() Node {
	condition := p.advance()
	if condition.Type != lexer.IDENTIFIER {
		panic("Expected condition after 'if'")
	}
	p.expectCloseCurly()
	parsedThen := p.Parse("else")
	p.advance()
	parsedElse := p.Parse("endif")
	p.advance()
	return NewIfNode(condition.Value, parsedThen, parsedElse)
}

func (p *Parser) expectCloseCurly() {
	if p.peek().Type != lexer.CLOSE_CURLY {
		panic("Expected close curly")
	}
	p.advance()
}

func (p *Parser) advance() lexer.Token {
	if p.isAtEnd() {
		return p.Tokens[len(p.Tokens)-1]
	}
	token := p.Tokens[p.currentPosition]
	p.currentPosition++
	return token
}

func (p *Parser) peek() lexer.Token {
	if p.isAtEnd() {
		return p.Tokens[len(p.Tokens)-1]
	}
	return p.Tokens[p.currentPosition]
}

func (p *Parser) isAtEnd() bool {
	return p.currentPosition >= len(p.Tokens)
}
