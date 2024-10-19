package parser

import (
	"testing"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expected         []Node
		allowPrettyPrint bool
	}{
		{
			name:    "Simple if statement",
			content: "Hello, {{ if is_admin }} You are an admin. {{ endif }}",
			expected: []Node{
				{Type: TEXT_NODE, Value: "Hello, "},
				{Type: IF_NODE, Value: "is_admin", Children: []Node{
					{Type: TEXT_NODE, Value: " You are an admin. "},
				}},
			},
		},
		{
			name:    "If-else statement",
			content: "Hello, {{ if is_admin }} You are an admin. {{ else }} You are not an admin. {{ endif }}",
			expected: []Node{
				{Type: TEXT_NODE, Value: "Hello, "},
				{
					Type:  IF_NODE,
					Value: "is_admin",
					Children: []Node{
						{Type: TEXT_NODE, Value: " You are an admin. "},
						{Type: TEXT_NODE, Value: " You are not an admin. "},
					},
				},
			},
		},
		{
			name:    "If-else statement",
			content: `Hello, {{ if is_admin }} You are an admin. {{ if is_super_admin}} SuperAdminIsComing {{ if is_super_super_admin}} Yessssss! {{endif}} {{endif}} {{ else }} You are not an admin. {{ endif }}`,
			expected: []Node{
				{Type: TEXT_NODE, Value: "Hello, "},
				{
					Type:  IF_NODE,
					Value: "is_admin",
					Children: []Node{
						{Type: TEXT_NODE, Value: " You are an admin. "},
						{
							Type:  IF_NODE,
							Value: "is_super_admin",
							Children: []Node{
								{Type: TEXT_NODE, Value: " SuperAdminIsComing "},
								{
									Type:  IF_NODE,
									Value: "is_super_super_admin",
									Children: []Node{
										{Type: TEXT_NODE, Value: " Yessssss! "},
									},
								},
								{Type: WHITESPACE_NODE, Value: " "},
							},
						},
						{Type: WHITESPACE_NODE, Value: " "},
						{Type: TEXT_NODE, Value: " You are not an admin. "},
					},
				},
			},
			allowPrettyPrint: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := New(lexer.New(tt.content).Tokenize()).Parse()
			if tt.allowPrettyPrint {
				PrettifyAST(ast)
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, ast)
		})
	}
}
