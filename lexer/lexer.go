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
	WHITESPACE
)

// Lets us pretify the enums when printing
func (tt TokenType) String() string {
	return [...]string{"TEXT", "OPEN_CURLY", "CLOSE_CURLY", "IDENTIFIER", "KEYWORD"}[tt]
}

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	Tokens  []Token
	crrPos  int
	rawText string
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

// ```
// Hello, {{ name }}! {{ if is_admin }} You are an admin.{{ end }}
// ```
func (l *Lexer) Tokenize() []Token {
	var sb strings.Builder
	for {
		char, ok := l.advance()
		if !ok {
			break // End of input
		}

		switch l.mode {
		case TextMode:
			if char == '{' {
				peek, _ := l.peek()
				if peek == '{' {
					//Append literal
					l.addToken(sb, TEXT)
					sb.Reset()

					//Append openning tag
					l.advance() // consume the second '}'
					var sbAlt strings.Builder
					sbAlt.WriteString("{{")
					l.addToken(sbAlt, OPEN_CURLY)

					//Switch to tag mode
					l.switchMode()
				}
			} else {
				sb.WriteRune(char)
			}
		case TagMode:
			// Handles the case where there is a space between keyword and identifiers
			if unicode.IsSpace(char) {
				if sb.Len() > 0 {
					if isKeyword(sb.String()) {
						l.addToken(sb, KEYWORD)
					} else {
						l.addToken(sb, IDENTIFIER)
					}
					sb.Reset()
				}
				// Handles keyword or identifier before closing the tag
			} else if char == '}' {
				peek, _ := l.peek()
				if peek == '}' {
					if sb.Len() > 0 {
						if isKeyword(sb.String()) {
							l.addToken(sb, KEYWORD)
						} else {
							l.addToken(sb, IDENTIFIER)
						}
						sb.Reset()
					}
					//Append closing tag
					l.advance() // consume the second '}'
					var sbAlt strings.Builder
					sbAlt.WriteString("}}")
					l.addToken(sbAlt, CLOSE_CURLY)

					//Switch to text mode
					l.switchMode()
				} else {
					sb.WriteRune(char)
				}

			} else {
				sb.WriteRune(char)
			}

		}
	}

	// Emit any remaining text
	if sb.Len() > 0 {
		if l.mode == TextMode {
			l.addToken(sb, TEXT)
		} else {
			// Handle unclosed tag
			log.Printf("Warning: Unclosed tag at end of input")
			if isKeyword(sb.String()) {
				l.addToken(sb, KEYWORD)
			} else {
				l.addToken(sb, IDENTIFIER)
			}
		}
	}

	return l.Tokens
}

func (l *Lexer) switchMode() {
	if l.mode == TextMode {
		l.mode = TagMode
	} else {
		l.mode = TextMode
	}
}

// Consumes a char
func (l *Lexer) advance() (rune, bool) {
	if l.crrPos >= len(l.rawText) {
		return 0, false
	}
	r := rune(l.rawText[l.crrPos])
	l.crrPos++
	return r, true
}

// Checks a char without consuming it
func (l *Lexer) peek() (rune, bool) {
	if l.crrPos >= len(l.rawText) {
		return 0, false
	}
	return rune(l.rawText[l.crrPos]), true
}

// Adds token to token slice
func (l *Lexer) addToken(value strings.Builder, tokenType TokenType) {
	if tokenType == TEXT && strings.TrimSpace(value.String()) == "" {
		l.Tokens = append(l.Tokens, Token{Value: value.String(), Type: WHITESPACE})
	} else {
		l.Tokens = append(l.Tokens, Token{Value: value.String(), Type: tokenType})
	}
}

// ----- Helper Functions -----

func isKeyword(word string) bool {
	return keywords[word]
}
