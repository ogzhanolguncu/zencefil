package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicLexer(t *testing.T) {
	content := "Hello, {{ name }}! {{ if is_admin }} You are an admin.{{ endif }}"
	tokens := New(content).Tokenize()
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
		{Type: KEYWORD, Value: "endif"},
		{Type: CLOSE_CURLY, Value: "}}"},
	}
	require.Equal(t, expected, tokens)
}

func TestLexerWithoutText(t *testing.T) {
	content := "{{ name }}"
	tokens := New(content).Tokenize()
	expected := []Token{
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "name"},
		{Type: CLOSE_CURLY, Value: "}}"},
	}
	require.Equal(t, expected, tokens)
}

func TestComplexTemplate(t *testing.T) {
	content := `
<html>
<body>
<h1>Welcome, {{ name }}!</h1>
{{ if loggedIn }}
  <p>Your tasks:</p>
  <ul>
  {{ for task in tasks }}
    <li>{{ task }}</li>
  {{ endfor }}
  </ul>
{{ else }}
  <p>Please log in to see your tasks.</p>
{{ endif }}
<footer>Copyright {{ year }}</footer>
</body>
</html>
`
	tokens := New(content).Tokenize()
	expected := []Token{
		{Type: TEXT, Value: "\n<html>\n<body>\n<h1>Welcome, "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "name"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "!</h1>\n"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "if"},
		{Type: IDENTIFIER, Value: "loggedIn"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n  <p>Your tasks:</p>\n  <ul>\n  "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "for"},
		{Type: IDENTIFIER, Value: "task"},
		{Type: KEYWORD, Value: "in"},
		{Type: IDENTIFIER, Value: "tasks"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n    <li>"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "task"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "</li>\n  "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "endfor"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n  </ul>\n"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "else"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n  <p>Please log in to see your tasks.</p>\n"},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: KEYWORD, Value: "endif"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "\n<footer>Copyright "},
		{Type: OPEN_CURLY, Value: "{{"},
		{Type: IDENTIFIER, Value: "year"},
		{Type: CLOSE_CURLY, Value: "}}"},
		{Type: TEXT, Value: "</footer>\n</body>\n</html>\n"},
	}
	require.Equal(t, expected, tokens)
}
