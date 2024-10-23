package parser

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ogzhanolguncu/zencefil/lexer"
)

type NodeType int

const (
	TEXT_NODE NodeType = iota
	VARIABLE_NODE

	IF_NODE
	THEN_BRANCH
	ELIF_BRANCH
	ELIF_ITEM
	ELSE_BRANCH

	FOR_NODE
	WHITESPACE_NODE
)

func (tt NodeType) String() string {
	return [...]string{"TEXT_NODE", "VARIABLE_NODE", "IF_NODE", "THEN_BRANCH", "ELIF_BRANCH", "ELIF_ITEM", "ELSE_BRANCH", "FOR_NODE", "WHITESPACE_NODE"}[tt]
}

type Node struct {
	Type     NodeType
	Value    *string
	Children []Node
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
		} else if p.match(lexer.WHITESPACE) {
			// TODO: If there are more than one space I should count them as one.
			prevVal := p.previous().Value
			nodes = append(nodes, NewNode(WHITESPACE_NODE, &prevVal))
		} else if p.match(lexer.OPEN_CURLY) {
			if p.match(lexer.KEYWORD) && p.previous().Value == "if" {
				IfNode, err := p.parseIf()
				if err != nil {
					return nil, fmt.Errorf("error parsing if statement: %w", err)
				}
				nodes = append(nodes, IfNode)
			} else if p.match(lexer.IDENTIFIER) {
				prevVal := p.previous().Value
				nodes = append(nodes, NewNode(VARIABLE_NODE, &prevVal))
				p.advance() // consume '}}' of variable node
			} else {
				// TODO: Later this will also handle 'for' and 'identifier' token
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unrecognized token: %v, they should start with -> '{{'", p.peek())
		}
	}
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
		} else if p.match(lexer.WHITESPACE) {
			prevVal := p.previous().Value
			nodes = append(nodes, NewNode(WHITESPACE_NODE, &prevVal))
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
					// TODO: handle later
				default:
					panic("Unknown KEYWORD")
				}
			} else if p.match(lexer.IDENTIFIER) {
				prevVal := p.previous().Value
				nodes = append(nodes, NewNode(VARIABLE_NODE, &prevVal))
				p.advance() // consume '}}' of variable node
			} else {
				// TODO: Later this will also handle 'for'  token
				return nil, fmt.Errorf("unexpected token after '{{': %v", p.peek())
			}
		} else {
			return nil, fmt.Errorf("unexpected token: %v", p.peek())
		}
	}
	return nodes, nil
}

func (p *Parser) isBlockEnd() bool {
	return p.isElseKeyword() || p.isElifKeyword() || p.isEndIfKeyword()
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

func (p *Parser) expectIfIdentifier() (string, error) {
	if !p.match(lexer.IDENTIFIER) {
		return "", fmt.Errorf("expected condition after 'if', got %v", p.peek())
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

func (p *Parser) expectCloseCurly() error {
	if !p.match(lexer.CLOSE_CURLY) {
		return fmt.Errorf("expected '}}', got %v", p.peek())
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

/// -------- Prettify Nodes --------

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
			nodeValueColor = color.New(color.FgHiCyan).SprintFunc()
		case ELSE_BRANCH:
			nodeValueColor = color.New(color.FgMagenta).SprintFunc()
		case FOR_NODE:
			nodeValueColor = color.New(color.FgBlue).SprintFunc()
		case WHITESPACE_NODE:
			nodeValueColor = color.New(color.FgWhite).SprintFunc()
		default:
			nodeValueColor = color.New(color.FgWhite).SprintFunc()
		}

		// Always print the node type
		if node.Value != nil {
			sb.WriteString(fmt.Sprintf("%s: %s\n",
				nodeTypeColor(node.Type),
				nodeValueColor(strings.ReplaceAll(strings.ReplaceAll(*node.Value, "\n", "\\n"), "\t", "\\t"))))
		} else {
			sb.WriteString(fmt.Sprintf("%s: \n", nodeTypeColor(node.Type)))
		}

		if len(node.Children) > 0 {
			prettifyNodes(sb, node.Children, indent+1)
		}
	}
}
