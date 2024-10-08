package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	content := "Hello, {{ name }}! {{ if is_admin }} You are an admin.{{ end }}"
	tokens := NewLexer(content).Tokenize()

	expected := []Token{
		{Type: TEXT, Value: "Hello, "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "name"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "! "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "if"},
		{Type: IDENTIFIER, Value: "is_admin"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: " You are an admin."},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "end"},
		{Type: CLOSE_CURLY, Value: "}}"},
	}

	require.Equal(t, expected, tokens)
}
func TestMoreComplexLexer(t *testing.T) {
	content := `
	<html>
	<body>
	<h1>Welcome, {{ name }}!</h1>
	{{ if loggedIn }}
	 <p>Your tasks:</p>
	 <ul>
	 {{ for task in tasks }}
	  <li>{{ task }}</li>
	 {{ end }}
	   </ul>
	   {{ else }}
		<p>Please log in to see your tasks.</p>
	{{ end }}
		<footer>Copyright {{ year }}</footer>
	</body>
	</html>
	`
	tokens := NewLexer(content).Tokenize()

	expected := []Token{
		{Type: TEXT, Value: "\n\t<html>\n\t<body>\n\t<h1>Welcome, "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "name"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "!</h1>\n\t"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "if"},
		{Type: IDENTIFIER, Value: "loggedIn"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n\t <p>Your tasks:</p>\n\t <ul>\n\t "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "for"},
		{Type: IDENTIFIER, Value: "task"},
		{Type: KEYWORD, Value: "in"},
		{Type: IDENTIFIER, Value: "tasks"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n\t  <li>"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "task"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "</li>\n\t "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "end"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n\t   </ul>\n\t   "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "else"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n\t\t<p>Please log in to see your tasks.</p>\n\t"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "end"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n\t\t<footer>Copyright "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "year"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "</footer>\n\t</body>\n\t</html>\n\t"},
	}

	require.Equal(t, expected, tokens)
}
