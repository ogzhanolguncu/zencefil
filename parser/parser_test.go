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
		shouldError      bool
	}{
		{
			name:    "Simple if statement",
			content: "Hello, {{ name }}! {{ if is_admin }} You are an admin.{{ endif }} {{ surname }}",
			expected: []Node{
				{Type: TEXT_NODE, Value: "Hello, "},
				{Type: VARIABLE_NODE, Value: "name"},
				{Type: TEXT_NODE, Value: "! "},
				{Type: IF_NODE, Value: "is_admin", Children: []Node{
					{Type: TEXT_NODE, Value: " You are an admin."},
				}},
				{Type: WHITESPACE_NODE, Value: " "},
				{Type: VARIABLE_NODE, Value: "surname"},
			},
		},
		{
			name: "Simple if-elseif-else statement",
			content: `Hello, {{ name }}!
			{{ if is_admin }}
			You are an admin.
			{{ elif is_super_admin }}
			  {{ if wohoo }}
				You are a super admin.
			  {{ else }}
			  	Whooot?
			   {{ endif}}
			{{ elif is_oz }}
				You are an oz.
			{{ elif is_dobby }}
				You are a super dobby.
			{{ else }}
				You are no body.
			{{endif}}`,
			expected: []Node{
				{Type: 0, Value: "Hello, ", Children: nil},
				{Type: 1, Value: "name", Children: nil},
				{Type: 0, Value: "!\n\t\t\t", Children: nil},
				{Type: 2, Value: "is_admin", Children: []Node{
					{Type: 0, Value: "\n\t\t\tYou are an admin.\n\t\t\t", Children: nil},
					{Type: 3, Value: "is_super_admin", Children: []Node{
						{Type: 5, Value: "\n\t\t\t  ", Children: nil},
						{Type: 2, Value: "wohoo", Children: []Node{
							{Type: 0, Value: "\n\t\t\t\tYou are a super admin.\n\t\t\t  ", Children: nil},
							{Type: 0, Value: "\n\t\t\t  \tWhooot?\n\t\t\t   ", Children: nil},
						}},
						{Type: 5, Value: "\n\t\t\t", Children: nil},
					}},
					{Type: 3, Value: "is_oz", Children: []Node{
						{Type: 0, Value: "\n\t\t\t\tYou are an oz.\n\t\t\t", Children: nil},
					}},
					{Type: 3, Value: "is_dobby", Children: []Node{
						{Type: 0, Value: "\n\t\t\t\tYou are a super dobby.\n\t\t\t", Children: nil},
					}},
					{Type: 0, Value: "\n\t\t\t\tYou are no body.\n\t\t\t", Children: nil},
				}},
			},
			allowPrettyPrint: true,
		},
		{
			name:             "Malformed template starting with 'endif' without 'if'",
			content:          "Hello, {{ endif }} asdasd",
			shouldError:      true,
			allowPrettyPrint: true,
		},
		{
			name:             "Malformed template starting with 'else' without 'if'",
			content:          "Hello, {{ else }} asdasd",
			shouldError:      true,
			allowPrettyPrint: true,
		},
		{
			name:             "Malformed template 'if' without condition",
			content:          "Hello, {{ if }} asdasd",
			shouldError:      true,
			allowPrettyPrint: true,
		},
		{
			name:             "Malformed template 'if' block without 'endif'",
			content:          "Hello, {{ if is_admin }} asdasd",
			shouldError:      true,
			allowPrettyPrint: true,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := New(lexer.New(tt.content).Tokenize()).Parse()

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			require.NoError(t, err)

			if tt.allowPrettyPrint {
				PrettifyAST(ast)
			}

			require.Equal(t, tt.expected, ast)
		})
	}
}
