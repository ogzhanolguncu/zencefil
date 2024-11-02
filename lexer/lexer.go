package lexer

import (
	"strconv"
	"strings"
	"unicode"
)

var keywords = map[string]bool{
	"if":     true,
	"elif":   true,
	"else":   true,
	"for":    true,
	"in":     true,
	"endif":  true,
	"endfor": true,
}

var Operators = map[string]TokenType{
	"&&": AMPERSAND,
	"||": PIPE,
	">=": GTE,
	">":  GT,
	"<":  LT,
	"<=": LTE,
	"==": EQ,
	"!=": NEQ,
	"??": NULL_COALESCE,
	"!":  BANG,
	"(":  LPAREN,
	")":  RPAREN,
	"[":  OPEN_BRACKET,
	"]":  CLOSE_BRACKET,
}

type ReadMode int

const (
	TextMode ReadMode = iota
	TagMode
)

type TokenType int

const (
	TEXT TokenType = iota
	OPEN_CURLY
	CLOSE_CURLY
	IDENTIFIER
	KEYWORD
	PIPE
	AMPERSAND
	NUMBER
	STRING
	LTE
	GTE
	EQ
	NEQ
	GT
	LT
	RPAREN
	LPAREN
	OPEN_BRACKET
	CLOSE_BRACKET
	BANG
	NULL_COALESCE
)

func (tt TokenType) String() string {
	return [...]string{
		"TEXT",
		"OPEN_CURLY",
		"CLOSE_CURLY",
		"IDENTIFIER",
		"KEYWORD",
		"PIPE",
		"AMPERSAND",
		"NUMBER",
		"STRING",
		"LTE",
		"GTE",
		"EQ",
		"NEQ",
		"GT",
		"LT",
		"RPAREN",
		"LPAREN",
		"OPEN_BRACKET",
		"CLOSE_BRACKET",
		"BANG",
		"NULL_COALESCE",
	}[tt]
}

type Token struct {
	Value string
	Type  TokenType
}

type Lexer struct {
	rawText string
	Tokens  []Token
	crrPos  int
	mode    ReadMode
}

func New(content string) *Lexer {
	return &Lexer{
		crrPos:  0,
		Tokens:  nil,
		rawText: content,
		mode:    TextMode,
	}
}

func (l *Lexer) Tokenize() []Token {
	var sb strings.Builder
	for {
		char, ok := l.advance()
		if !ok {
			break
		}

		switch l.mode {
		case TextMode:
			peek, _ := l.peek()
			if char == '{' && peek == '{' {
				if sb.Len() > 0 {
					text := sb.String()
					l.Tokens = append(l.Tokens, Token{Value: text, Type: TEXT})
					sb.Reset()
				}
				l.advance() // consume the second '{'
				l.Tokens = append(l.Tokens, Token{Value: "{{", Type: OPEN_CURLY})
				l.mode = TagMode
			} else {
				sb.WriteRune(char)
			}

		case TagMode:
			if char == '\'' {
				// Start of a string literal
				sb.WriteRune(char)
				for {
					innerChar, ok := l.advance()
					if !ok {
						break
					}
					sb.WriteRune(innerChar)
					if innerChar == '\'' {
						// End of string literal found
						str := sb.String()
						content := strings.Trim(str, "'") // Remove surrounding quotes
						l.Tokens = append(l.Tokens, Token{Value: content, Type: STRING})
						sb.Reset()
						break
					}
				}
				continue
			}

			if unicode.IsSpace(char) {
				if sb.Len() > 0 {
					l.addToken(sb.String())
					sb.Reset()
				}
				continue
			}
			if char == '}' {
				peek, _ := l.peek()
				if peek == '}' {
					if sb.Len() > 0 {
						l.addToken(sb.String())
						sb.Reset()
					}
					l.advance() // consume the second '}'
					l.Tokens = append(l.Tokens, Token{Value: "}}", Type: CLOSE_CURLY})
					l.mode = TextMode
					continue
				}
			}

			// Check for two-character operators
			currentChar := string(char)
			peek, hasPeek := l.peek()
			if hasPeek {
				potentialOp := currentChar + string(peek)
				if tokenType, exists := Operators[potentialOp]; exists {
					if sb.Len() > 0 {
						l.addToken(sb.String())
						sb.Reset()
					}
					l.advance() // consume the second character
					l.Tokens = append(l.Tokens, Token{Value: potentialOp, Type: tokenType})
					continue
				}
			}

			// Check for single-character operators, e.g '!', '>','<'
			if tokenType, exists := Operators[currentChar]; exists {
				if sb.Len() > 0 {
					l.addToken(sb.String())
					sb.Reset()
				}
				l.Tokens = append(l.Tokens, Token{Value: currentChar, Type: tokenType})
				continue
			}

			sb.WriteRune(char)
		}
	}

	// Handle any remaining text
	if sb.Len() > 0 {
		if l.mode == TextMode {
			l.Tokens = append(l.Tokens, Token{Value: sb.String(), Type: TEXT})
		} else {
			l.addToken(sb.String())
		}
	}

	return l.Tokens
}

func (l *Lexer) addToken(text string) {
	if text == "" {
		return
	}

	switch {
	case keywords[text]:
		l.Tokens = append(l.Tokens, Token{Value: text, Type: KEYWORD})
	case isNumber(text):
		l.Tokens = append(l.Tokens, Token{Value: text, Type: NUMBER})
	case isString(text):
		str := strings.Trim(text, "'")
		l.Tokens = append(l.Tokens, Token{Value: str, Type: STRING})
	default:
		l.Tokens = append(l.Tokens, Token{Value: text, Type: IDENTIFIER})
	}
}

func (l *Lexer) advance() (rune, bool) {
	if l.crrPos >= len(l.rawText) {
		return 0, false
	}
	r := rune(l.rawText[l.crrPos])
	l.crrPos++
	return r, true
}

func (l *Lexer) peek() (rune, bool) {
	if l.crrPos >= len(l.rawText) {
		return 0, false
	}
	return rune(l.rawText[l.crrPos]), true
}

func isNumber(text string) bool {
	_, err := strconv.ParseFloat(text, 64)
	return err == nil
}

func isString(text string) bool {
	return strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'")
}
