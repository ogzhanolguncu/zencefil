package lexer

import (
	"log"
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
)

func (tt TokenType) String() string {
	return [...]string{"TEXT", "OPEN_CURLY", "CLOSE_CURLY", "IDENTIFIER", "KEYWORD", "WHITESPACE"}[tt]
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
				// Handle accumulated text before the tag
				if sb.Len() > 0 {
					text := sb.String()
					l.Tokens = append(l.Tokens, Token{Value: text, Type: TEXT})
				}
				sb.Reset()

				// Handle opening tag
				l.advance() // consume the second '{'
				l.Tokens = append(l.Tokens, Token{Value: "{{", Type: OPEN_CURLY})
				l.mode = TagMode
			} else {
				sb.WriteRune(char)
			}

		case TagMode:
			if unicode.IsSpace(char) {
				if sb.Len() > 0 {
					text := sb.String()
					if isKeyword(text) {
						l.Tokens = append(l.Tokens, Token{Value: text, Type: KEYWORD})
					} else {
						l.Tokens = append(l.Tokens, Token{Value: text, Type: IDENTIFIER})
					}
					sb.Reset()
				}
				// Skip whitespace in tag mode
			} else if char == '}' {
				peek, _ := l.peek()
				if peek == '}' {
					if sb.Len() > 0 {
						text := sb.String()
						if isKeyword(text) {
							l.Tokens = append(l.Tokens, Token{Value: text, Type: KEYWORD})
						} else {
							l.Tokens = append(l.Tokens, Token{Value: text, Type: IDENTIFIER})
						}
						sb.Reset()
					}

					l.advance() // consume the second '}'
					l.Tokens = append(l.Tokens, Token{Value: "}}", Type: CLOSE_CURLY})
					l.mode = TextMode
				} else {
					sb.WriteRune(char)
				}
			} else {
				sb.WriteRune(char)
			}
		}
	}

	// Handle any remaining text
	if sb.Len() > 0 {
		text := sb.String()
		if l.mode == TextMode {
			l.Tokens = append(l.Tokens, Token{Value: text, Type: TEXT})
		} else {
			log.Printf("Warning: Unclosed tag at end of input")
			if isKeyword(text) {
				l.Tokens = append(l.Tokens, Token{Value: text, Type: KEYWORD})
			} else {
				l.Tokens = append(l.Tokens, Token{Value: text, Type: IDENTIFIER})
			}
		}
	}

	return l.Tokens
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

func isKeyword(word string) bool {
	return keywords[word]
}
