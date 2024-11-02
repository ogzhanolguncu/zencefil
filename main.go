package main

import (
	"fmt"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/ogzhanolguncu/zencefil/parser"
	"github.com/ogzhanolguncu/zencefil/renderer"
)

func main() {
	// // Example 1: Simple Template
	simpleExample()
	//
	// // Example 2: Complex Nested Conditionals
	nestedConditionalsExample()
	//
	// // Example 3: Loops and Object Access
	loopsAndObjectsExample()

	// Example 4: Complex Expressions
	complexExpressionsExample()
}

func simpleExample() {
	fmt.Println("\n=== Simple Template Example ===")

	content := `Hello, {{ name }}! {{ if isAdmin }}You are an admin.{{ else }}You are a regular user.{{ endif }}`

	context := map[string]interface{}{
		"name":    "John",
		"isAdmin": true,
	}

	result := renderTemplate(content, context)
	fmt.Println(result)
}

func nestedConditionalsExample() {
	fmt.Println("\n=== Nested Conditionals Example ===")

	content := `
Welcome, {{ name }}!
{{ if isAdmin }}
    Admin Panel:
    {{ if hasFullAccess }}
        Full administrative access granted.
        {{ if canManageUsers }}
            User management enabled.
        {{ endif }}
    {{ else }}
        Limited administrative access.
    {{ endif }}
{{ elif isModerator }}
    Moderator Tools Available
{{ else }}
    Regular User Interface
{{ endif }}
`

	context := map[string]interface{}{
		"name":           "Alice",
		"isAdmin":        true,
		"hasFullAccess":  true,
		"canManageUsers": true,
		"isModerator":    false,
	}

	result := renderTemplate(content, context)
	fmt.Println(result)
}

func loopsAndObjectsExample() {
	fmt.Println("\n=== Loops and Object Access Example ===")

	content := `
Inventory Report:
{{ for item in inventory }}
    - {{ item['name'] }}: {{ item['quantity'] }} units at ${{ item['price'] }}
    {{ if item['quantity'] < 5 }}
        [LOW STOCK ALERT]
    {{ endif }}
{{ endfor }}

Total Items: {{ totalItems }}
`

	context := map[string]interface{}{
		"inventory": []interface{}{
			map[string]interface{}{
				"name":     "Widget",
				"quantity": 3,
				"price":    19.99,
			},
			map[string]interface{}{
				"name":     "Gadget",
				"quantity": 8,
				"price":    24.99,
			},
			map[string]interface{}{
				"name":     "Tool",
				"quantity": 2,
				"price":    15.99,
			},
		},
		"totalItems": 3,
	}

	result := renderTemplate(content, context)
	fmt.Println(result)
}

func complexExpressionsExample() {
	fmt.Println("\n=== Complex Expressions Example ===")

	content := `
User Status Report:
{{ if (age >= 18 && (role == 'admin' || role == 'moderator') && !isBlocked) }}
    Full Access Granted
{{ elif (age >= 16 && role == 'junior-mod' && postCount > 100) || (isPremium && trustScore > 8.5) }}
    Limited Access Granted
{{ else }}
    Basic Access Only
{{ endif }}

Account Type: {{ accountType ?? 'Standard' }}
Verification: {{ isVerified && hasMFA && 'Fully Verified' || 'Incomplete' }}
`

	context := map[string]interface{}{
		"age":         20,
		"role":        "admin",
		"isBlocked":   false,
		"postCount":   150,
		"isPremium":   true,
		"trustScore":  9.0,
		"isVerified":  true,
		"hasMFA":      true,
		"accountType": "Full",
	}

	result := renderTemplate(content, context)
	fmt.Println(result)
}

// Helper function to handle the template rendering process
func renderTemplate(content string, context map[string]interface{}) string {
	// Lexical analysis
	tokens := lexer.New(content).Tokenize()

	// Parsing
	ast, err := parser.New(tokens).Parse()
	if err != nil {
		return fmt.Sprintf("Parse error: %v", err)
	}

	// Rendering
	result, err := renderer.New(ast, context).Render()
	if err != nil {
		return fmt.Sprintf("Render error: %v", err)
	}

	return result
}
