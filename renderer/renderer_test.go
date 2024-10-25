package renderer

import (
	"testing"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/ogzhanolguncu/zencefil/parser"
	"github.com/stretchr/testify/require"
)

func TestRenderer(t *testing.T) {
	tests := []struct {
		context     map[string]interface{}
		name        string
		content     string
		expected    string
		shouldError bool
	}{
		{
			name:    "Simple if statement",
			content: "Hello, {{ name }}! {{ if isAdmin }}You are an admin.{{ endif }}",
			context: map[string]interface{}{
				"name":    "Oz",
				"isAdmin": true,
			},
			expected: "Hello, Oz! You are an admin.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.New(tt.content).Tokenize()
			ast, err := parser.New(tokens).Parse()
			template, err := New(ast, tt.context).Render()

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, template)
		})
	}
}
