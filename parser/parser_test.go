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
				{Type: TEXT_NODE, Value: ptrStr("Hello, ")},
				{Type: VARIABLE_NODE, Value: ptrStr("name")},
				{Type: TEXT_NODE, Value: ptrStr("! ")},
				{Type: IF_NODE, Value: ptrStr("is_admin"), Children: []Node{
					{Type: THEN_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" You are an admin.")},
					}},
				}},
				{Type: WHITESPACE_NODE, Value: ptrStr(" ")},
				{Type: VARIABLE_NODE, Value: ptrStr("surname")},
			},
		},
		{
			name:    "Simple if-elseif-else statement",
			content: "Hello {{ if is_admin }}admin{{ elif is_super }}super{{ elif is_user }}user{{ else }}guest{{ endif }}!",
			expected: []Node{
				{Type: TEXT_NODE, Value: ptrStr("Hello ")},
				{Type: IF_NODE, Value: ptrStr("is_admin"), Children: []Node{
					{Type: THEN_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr("admin")},
					}},
					{Type: ELIF_BRANCH, Value: nil, Children: []Node{
						{Type: ELIF_ITEM, Value: ptrStr("is_super"), Children: []Node{
							{Type: TEXT_NODE, Value: ptrStr("super")},
						}},
						{Type: ELIF_ITEM, Value: ptrStr("is_user"), Children: []Node{
							{Type: TEXT_NODE, Value: ptrStr("user")},
						}},
					}},
					{Type: ELSE_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr("guest")},
					}},
				}},
				{Type: TEXT_NODE, Value: ptrStr("!")},
			},
			allowPrettyPrint: true,
		},
		{
			name:        "Malformed template starting with 'endif' without 'if'",
			content:     "Hello, {{ endif }} asdasd",
			shouldError: true,
		},
		{
			name:        "Malformed template starting with 'else' without 'if'",
			content:     "Hello, {{ else }} asdasd",
			shouldError: true,
		},
		{
			name:        "Malformed template 'if' without condition",
			content:     "Hello, {{ if }} asdasd",
			shouldError: true,
		},
		{
			name:        "Malformed template 'if' block without 'endif'",
			content:     "Hello, {{ if is_admin }} asdasd",
			shouldError: true,
		},
		{
			name:    "If-else statement",
			content: "Hello, {{ if is_admin }} You are an admin. {{ else }} You are not an admin. {{ endif }}",
			expected: []Node{
				{Type: TEXT_NODE, Value: ptrStr("Hello, ")},
				{Type: IF_NODE, Value: ptrStr("is_admin"), Children: []Node{
					{Type: THEN_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" You are an admin. ")},
					}},
					{Type: ELSE_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" You are not an admin. ")},
					}},
				}},
			},
		},
		{
			name:    "Nested if-else statement",
			content: `Hello, {{ if is_admin }} You are an admin. {{ if is_super_admin}} SuperAdminIsComing {{ if is_super_super_admin}} Yessssss! {{endif}} {{endif}} {{ else }} You are not an admin. {{ endif }}`,
			expected: []Node{
				{Type: TEXT_NODE, Value: ptrStr("Hello, ")},
				{Type: IF_NODE, Value: ptrStr("is_admin"), Children: []Node{
					{Type: THEN_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" You are an admin. ")},
						{Type: IF_NODE, Value: ptrStr("is_super_admin"), Children: []Node{
							{Type: THEN_BRANCH, Value: nil, Children: []Node{
								{Type: TEXT_NODE, Value: ptrStr(" SuperAdminIsComing ")},
								{Type: IF_NODE, Value: ptrStr("is_super_super_admin"), Children: []Node{
									{Type: THEN_BRANCH, Value: nil, Children: []Node{
										{Type: TEXT_NODE, Value: ptrStr(" Yessssss! ")},
									}},
								}},
								{Type: WHITESPACE_NODE, Value: ptrStr(" ")},
							}},
						}},
						{Type: WHITESPACE_NODE, Value: ptrStr(" ")},
					}},
					{Type: ELSE_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" You are not an admin. ")},
					}},
				}},
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

func ptrStr(s string) *string { return &s }
