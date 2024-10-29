package parser

import (
	"fmt"
	"testing"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name                  string
		content               string
		expected              []Node
		allowPrettyPrintAST   bool
		allowPrettyPrintToken bool
		shouldError           bool
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
				{Type: TEXT_NODE, Value: ptrStr(" ")},
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
								{Type: TEXT_NODE, Value: ptrStr(" ")}, // Changed from WHITESPACE_NODE to TEXT_NODE
							}},
						}},
						{Type: TEXT_NODE, Value: ptrStr(" ")}, // Changed from WHITESPACE_NODE to TEXT_NODE
					}},
					{Type: ELSE_BRANCH, Value: nil, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" You are not an admin. ")},
					}},
				}},
			},
		},
		{
			name:    "Simple for statement",
			content: "{{for item in items}} dobby has this item:{{item}} {{endfor}}",
			expected: []Node{
				{Type: FOR_NODE, Children: []Node{
					{Type: ITERATEE_ITEM, Value: ptrStr("item")},
					{Type: ITERATOR_ITEM, Value: ptrStr("items")},
					{Type: FOR_BODY, Children: []Node{
						{Type: TEXT_NODE, Value: ptrStr(" dobby has this item:")},
						{Type: VARIABLE_NODE, Value: ptrStr("item")},
						{Type: TEXT_NODE, Value: ptrStr(" ")},
					}},
				}},
			},
		},
		{
			name:    "variable with complex expression",
			content: "Hello, {{ name == 'dobby' && age > 18 || is_wizard }}",
			expected: []Node{
				{Type: TEXT_NODE, Value: ptrStr("Hello, ")},
				{
					Type: EXPRESSION_NODE,
					Children: []Node{
						// First comparison
						{Type: VARIABLE_NODE, Value: ptrStr("name")},
						{Type: OP_EQUALS, Value: ptrStr("==")},
						{Type: STRING_LITERAL_NODE, Value: ptrStr("dobby")},
						// AND operator
						{Type: OP_AND, Value: ptrStr("&&")},
						// Second comparison
						{Type: VARIABLE_NODE, Value: ptrStr("age")},
						{Type: OP_GT, Value: ptrStr(">")},
						{Type: NUMBER_LITERAL_NODE, Value: ptrStr("18")},
						// OR operator
						{Type: OP_OR, Value: ptrStr("||")},
						// Third condition
						{Type: VARIABLE_NODE, Value: ptrStr("is_wizard")},
					},
				},
			},
		},
		// {
		// 	name:    "if with expression",
		// 	content: "{{ if is_admin && is_active}} You are an admin and active.{{ endif }}",
		// 	expected: []Node{
		// 		{Type: TEXT_NODE, Children: []Node{
		// 			{Type: EXPRESSION_NODE, Children: []Node{
		// 				{Type: VARIABLE_NODE, Value: ptrStr("name")},
		// 				{Type: OP_NODE, Value: ptrStr("==")},
		// 				{Type: STRING_LITERAL_NODE, Value: ptrStr("dobby")},
		// 			}},
		// 		}},
		// 		{Type: TEXT_NODE, Value: ptrStr(" ")}, // Added space between expressions
		// 		{Type: IF_NODE, Children: []Node{
		// 			{Type: EXPRESSION_NODE, Children: []Node{
		// 				{Type: VARIABLE_NODE, Value: ptrStr("is_admin")},
		// 				{Type: OP_NODE, Value: ptrStr("&&")},
		// 				{Type: VARIABLE_NODE, Value: ptrStr("is_active")},
		// 			}},
		// 			{Type: THEN_BRANCH, Children: []Node{
		// 				{Type: TEXT_NODE, Value: ptrStr(" You are an admin and active.")},
		// 			}},
		// 		}},
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.New(tt.content).Tokenize()
			ast, err := New(tokens).Parse()

			if tt.allowPrettyPrintToken {
				fmt.Print(lexer.PrettyPrintTokens(tokens))
			}

			if tt.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.allowPrettyPrintAST {
				PrettifyAST(ast)
			}

			require.Equal(t, tt.expected, ast)
		})
	}
}

func ptrStr(s string) *string { return &s }
