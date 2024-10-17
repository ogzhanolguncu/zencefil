package parser

import (
	"testing"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	content := "Hello, {{ if is_admin }} You are an admin. {{ endif }}"
	ast, err := New(lexer.New(content).Tokenize()).Parse()
	require.NoError(t, err)
	expectedAST := []Node{
		{Type: TextNode, Value: "Hello, ", Children: []Node(nil)},
		{Type: IfNode, Value: "is_admin", Children: []Node{
			{Type: TextNode, Value: " You are an admin. ", Children: []Node(nil)},
		}},
	}
	require.Equal(t, expectedAST, ast)
}

func TestParserWithIfElse(t *testing.T) {
	content := "Hello, {{ if is_admin }} You are an admin. {{ else }} You are not an admin. {{ endif }}"
	ast, err := New(lexer.New(content).Tokenize()).Parse()
	require.NoError(t, err)
	expectedAST := []Node{
		{Type: TextNode, Value: "Hello, ", Children: []Node(nil)},
		{
			Type:  IfNode,
			Value: "is_admin",
			Children: []Node{
				{Type: TextNode, Value: " You are an admin. ", Children: []Node(nil)},
				{Type: TextNode, Value: " You are not an admin. ", Children: []Node(nil)},
			},
		},
	}
	require.Equal(t, expectedAST, ast)
}
