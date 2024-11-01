package renderer

import (
	"testing"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/ogzhanolguncu/zencefil/parser"
	"github.com/stretchr/testify/require"
)

func TestRenderer(t *testing.T) {
	tests := []struct {
		context             map[string]interface{}
		name                string
		content             string
		expected            string
		errorContains       string
		shouldError         bool
		allowPrettyPrintAST bool
	}{
		// Basic functionality tests
		{
			name:     "Simple text without variables",
			content:  "Hello, World!",
			context:  map[string]interface{}{},
			expected: "Hello, World!",
		},
		{
			name:    "Simple variable substitution",
			content: "Hello, {{ name }}!",
			context: map[string]interface{}{
				"name": "Oz",
			},
			expected: "Hello, Oz!",
		},

		// Variable edge cases
		{
			name:          "Missing variable",
			content:       "Hello, {{ name }}!",
			context:       map[string]interface{}{},
			shouldError:   true,
			errorContains: "not found in context",
		},
		{
			name:    "Variable with special characters",
			content: "Value: {{ special@var }}",
			context: map[string]interface{}{
				"special@var": "special value",
			},
			expected: "Value: special value",
		},
		{
			name:    "Variable with different types",
			content: "Number: {{ num }}, Bool: {{ bool }}, Nil: {{ nilVar }}",
			context: map[string]interface{}{
				"num":    42,
				"bool":   true,
				"nilVar": nil,
			},
			expected: "Number: 42, Bool: true, Nil: <nil>",
		},

		// Conditional tests
		{
			name:    "Simple if statement",
			content: "{{ if isAdmin }}Admin{{ endif }}",
			context: map[string]interface{}{
				"isAdmin": true,
			},
			expected: "Admin",
		},
		{
			name:    "If statement with false condition",
			content: "{{ if isAdmin }}Admin{{ endif }}",
			context: map[string]interface{}{
				"isAdmin": false,
			},
			expected: "",
		},
		{
			name:    "If-else with true condition",
			content: "{{ if isAdmin }}Admin{{ else }}User{{ endif }}",
			context: map[string]interface{}{
				"isAdmin": true,
			},
			expected: "Admin",
		},
		{
			name:    "If-else with false condition",
			content: "{{ if isAdmin }}Admin{{ else }}User{{ endif }}",
			context: map[string]interface{}{
				"isAdmin": false,
			},
			expected: "User",
		},

		// Complex conditional tests
		{
			name:    "Multiple elif branches with first true",
			content: "{{ if isAdmin }}Admin{{ elif isModerator }}Mod{{ elif isUser }}User{{ else }}Guest{{ endif }}",
			context: map[string]interface{}{
				"isAdmin":     true,
				"isModerator": true,
				"isUser":      true,
			},
			expected: "Admin",
		},
		{
			name:    "Multiple elif branches with middle true",
			content: "{{ if isAdmin }}Admin{{ elif isModerator }}Mod{{ elif isUser }}User{{ else }}Guest{{ endif }}",
			context: map[string]interface{}{
				"isAdmin":     false,
				"isModerator": true,
				"isUser":      true,
			},
			expected: "Mod",
		},
		{
			name:    "Multiple elif branches with none true",
			content: "{{ if isAdmin }}Admin{{ elif isModerator }}Mod{{ elif isUser }}User{{ else }}Guest{{ endif }}",
			context: map[string]interface{}{
				"isAdmin":     false,
				"isModerator": false,
				"isUser":      false,
			},
			expected: "Guest",
		},

		// Nested conditional tests
		{
			name: "Nested if statements",
			content: `{{ if isLoggedIn }}
				Welcome!
				{{ if isAdmin }}
					Admin Panel
					{{ if hasFullAccess }}Full Access{{ endif }}
				{{ endif }}
			{{ endif }}`,
			context: map[string]interface{}{
				"isLoggedIn":    true,
				"isAdmin":       true,
				"hasFullAccess": true,
			},
			expected: "\n\t\t\t\tWelcome!\n\t\t\t\t\n\t\t\t\t\tAdmin Panel\n\t\t\t\t\tFull Access\n\t\t\t\t\n\t\t\t",
		},

		// Error cases
		{
			name:          "Non-boolean if condition",
			content:       "{{ if nonBool }}Test{{ endif }}",
			context:       map[string]interface{}{"nonBool": "string"},
			shouldError:   true,
			errorContains: "is not a boolean",
		},
		{
			name:          "Non-boolean elif condition",
			content:       "{{ if isAdmin }}Admin{{ elif nonBool }}Test{{ endif }}",
			context:       map[string]interface{}{"isAdmin": false, "nonBool": 42},
			shouldError:   true,
			errorContains: "is not a boolean",
		},
		{
			name:          "Missing if condition",
			content:       "{{ if missingVar }}Test{{ endif }}",
			context:       map[string]interface{}{},
			shouldError:   true,
			errorContains: "not found in context",
		},

		// Mixed content tests
		{
			name: "Complex mixed content",
			content: `Welcome {{ name }}!
			{{ if isAdmin }}
				Admin Settings:
				{{ if hasFullAccess }}Full{{ else }}Limited{{ endif }} Access
			{{ elif isModerator }}
				Mod Tools Available
			{{ else }}
				{{ if isSubscriber }}Premium{{ else }}Basic{{ endif }} User
			{{ endif }}`,
			context: map[string]interface{}{
				"name":          "John",
				"isAdmin":       true,
				"hasFullAccess": false,
				"isModerator":   false,
				"isSubscriber":  true,
			},
			expected: "Welcome John!\n\t\t\t\n\t\t\t\tAdmin Settings:\n\t\t\t\tLimited Access\n\t\t\t",
		},

		// Whitespace handling
		{
			name: "Whitespace preservation",
			content: `Line 1
			{{ if isAdmin }}
				Admin
			{{ endif }}
			Line 2`,
			context: map[string]interface{}{
				"isAdmin": true,
			},
			expected: "Line 1\n\t\t\t\n\t\t\t\tAdmin\n\t\t\t\n\t\t\tLine 2",
		},
		// Loops
		{
			name:    "Basic loop test",
			content: "ahmet has this items:{{for item in items}}\n{{item}}{{endfor}}",
			context: map[string]interface{}{
				"items": []interface{}{"pen", "pencil", "book"},
			},
			expected: "ahmet has this items:\npen\npencil\nbook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.New(tt.content).Tokenize()
			ast, err := parser.New(tokens).Parse()
			require.NoError(t, err, "Parser should not fail")

			if tt.allowPrettyPrintAST {
				parser.PrettifyAST(ast)
			}

			template, err := New(ast, tt.context).Render()

			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, template)
		})
	}
}

// TestRendererNilCases tests nil handling
func TestRendererNilCases(t *testing.T) {
	tests := []struct {
		context       map[string]interface{}
		name          string
		errorContains string
		ast           []parser.Node
		shouldError   bool
	}{
		{
			name:    "Nil AST",
			ast:     nil,
			context: map[string]interface{}{},
		},
		{
			name:    "Nil context",
			ast:     []parser.Node{},
			context: nil,
		},
		{
			name: "Node with nil value",
			ast: []parser.Node{
				{Type: parser.TEXT_NODE, Value: nil},
			},
			context:       map[string]interface{}{},
			shouldError:   true,
			errorContains: "nil value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := New(tt.ast, tt.context)
			result, err := renderer.Render()

			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.Empty(t, result)
		})
	}
}
