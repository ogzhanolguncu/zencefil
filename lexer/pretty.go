package lexer

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func PrettyPrintTokens(tokens []Token) string {
	var sb strings.Builder
	prettifyTokens(&sb, tokens, 0)
	return sb.String()
}

func prettifyTokens(sb *strings.Builder, tokens []Token, indent int) {
	tokenTypeColor := color.New(color.FgCyan, color.Bold).SprintFunc()

	for _, token := range tokens {
		sb.WriteString(strings.Repeat("  ", indent))

		var tokenValueColor func(a ...interface{}) string
		switch token.Type {
		case TEXT:
			tokenValueColor = color.New(color.FgGreen).SprintFunc()
		case IDENTIFIER:
			tokenValueColor = color.New(color.FgYellow).SprintFunc()
		case KEYWORD:
			tokenValueColor = color.New(color.FgMagenta).SprintFunc()
		case NUMBER:
			tokenValueColor = color.New(color.FgBlue).SprintFunc()
		case STRING:
			tokenValueColor = color.New(color.FgGreen).SprintFunc()
		case OPEN_CURLY, CLOSE_CURLY:
			tokenValueColor = color.New(color.FgRed).SprintFunc()
		case PIPE, AMPERSAND, GT, LT, GTE, LTE, EQ, NEQ, BANG, LPAREN, RPAREN:
			tokenValueColor = color.New(color.FgYellow).SprintFunc()
		default:
			tokenValueColor = color.New(color.FgWhite).SprintFunc()
		}

		formattedValue := strings.ReplaceAll(strings.ReplaceAll(token.Value, "\n", "\\n"), "\t", "\\t")

		// Print the token type and value
		fmt.Fprintf(sb, "%s: %s\n",
			tokenTypeColor(token.Type),
			tokenValueColor(formattedValue))
	}
}

// Helper function to use the pretty printer
func (l *Lexer) PrettyPrint() string {
	return PrettyPrintTokens(l.Tokens)
}
