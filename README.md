# Zencefil

Zencefil is a sophisticated template engine written in Go, designed to provide flexible text templating with a familiar syntax.

## About

Zencefil (Turkish for "ginger") is a template engine implemented in Go that supports a rich set of features for text manipulation and dynamic content generation. While created as a learning project, it offers robust functionality similar to established template engines like Jinja2 or Django templates.

## Performance

Zencefil is designed for high performance, showing significant speed advantages over similar template engines. Here's a comparison with Jinja2 (Python) across different operations:

| Operation             | Zencefil    | Jinja2      | Performance Difference |
| --------------------- | ----------- | ----------- | ---------------------- |
| Simple text           | 75 ns/op    | 4,881 ns/op | ~65x faster            |
| Variable substitution | 187 ns/op   | 5,286 ns/op | ~28x faster            |
| Nested object access  | 356 ns/op   | 6,162 ns/op | ~17x faster            |
| Complex conditions    | 1,013 ns/op | 5,338 ns/op | ~5x faster             |
| Simple loop           | 940 ns/op   | 5,664 ns/op | ~6x faster             |
| Complex loop          | 1,614 ns/op | 8,445 ns/op | ~5x faster             |

_Benchmark details:_

- Each operation tested with 1000 iterations after 100 warmup iterations
- Tests run on the same machine under identical conditions
- Zencefil maintains sub-millisecond performance even for complex operations
- Performance advantage is most pronounced in simple operations while remaining significant in complex scenarios

## Features

### Variable Substitution

- Basic variable output: `{{ variable }}`
- Object/map access: `{{ person['address'] }}`
- Null coalescing operator: `{{ value ?? 'default' }}`

### Control Structures

#### Conditionals

- Basic if statements: `{{ if condition }}...{{ endif }}`
- If-else blocks: `{{ if condition }}...{{ else }}...{{ endif }}`
- Multiple elif branches: `{{ if condition }}...{{ elif condition2 }}...{{ elif condition3 }}...{{ else }}...{{ endif }}`
- Nested conditionals supported

#### Loops

- For loops with iterables: `{{ for item in items }}...{{ endfor }}`
- Access to loop variables within the loop body
- Nested loops supported

### Expressions

#### Logical Operators

- AND: `&&`
- OR: `||`
- NOT: `!`

#### Comparison Operators

- Equals: `==`
- Not equals: `!=`
- Greater than: `>`
- Less than: `<`
- Greater than or equal: `>=`
- Less than or equal: `<=`

#### Special Features

- Parenthesized expressions: `{{ (age >= 18 && (role == 'admin' || role == 'moderator')) }}`
- Complex nested expressions
- Type-aware comparisons (strings, numbers, booleans)
- Whitespace preservation

### Data Types Support

- Strings (with single quotes): `'string value'`
- Numbers: `42`, `3.14`
- Booleans: `true`, `false`
- Arrays/Slices
- Maps/Objects

### Debugging and Development Tools

#### AST Visualization

Zencefil includes a built-in AST (Abstract Syntax Tree) visualizer that helps you understand how your templates are parsed. This is particularly useful for debugging complex templates or during development.

```go
import (
    "github.com/ogzhanolguncu/zencefil/lexer"
    "github.com/ogzhanolguncu/zencefil/parser"
)

func main() {
    // Your template
    content := `
        {{ if isAdmin && isActive }}
            Welcome, {{ name }}!
            {{ if hasFullAccess }}
                Full access granted.
            {{ endif }}
        {{ endif }}
    `

    // Generate tokens
    tokens := lexer.New(content).Tokenize()

    // Parse tokens into AST
    ast, _ := parser.New(tokens).Parse()

    // Visualize the AST
    parser.PrettifyAST(ast)
}
```

This will output a color-coded visualization of your template's AST structure:

```
IF_NODE:
  EXPRESSION_NODE:
    VARIABLE_NODE: isAdmin
    OP_AND: &&
    VARIABLE_NODE: isActive
  THEN_BRANCH:
    TEXT_NODE: Welcome,
    VARIABLE_NODE: name
    TEXT_NODE: !
    IF_NODE:
      VARIABLE_NODE: hasFullAccess
      THEN_BRANCH:
        TEXT_NODE: Full access granted.
```

The visualization uses different colors to distinguish between:

- Control structures (if, for, etc.)
- Variables
- Operators
- Text nodes
- Expressions

This visualization helps you:

- Debug complex templates
- Understand how the parser interprets your template
- Verify that nested structures are correctly parsed
- Identify potential issues in template syntax

## Example Usage

```go
// Initialize the template engine
content := `
<h1>Welcome, {{ name }}!</h1>
{{ if isAdmin }}
    <div class="admin-panel">
        {{ if hasFullAccess }}
            <p>Full admin access granted</p>
        {{ else }}
            <p>Limited admin access</p>
        {{ endif }}
    </div>
{{ elif isModerator }}
    <div class="mod-panel">
        <p>Moderator tools available</p>
    </div>
{{ else }}
    <p>Welcome, regular user!</p>
{{ endif }}

<h2>Your Items:</h2>
<ul>
{{ for item in items }}
    <li>{{ item['name'] }}: {{ item['quantity'] }}</li>
{{ endfor }}
</ul>
`

context := map[string]interface{}{
    "name": "John",
    "isAdmin": true,
    "hasFullAccess": false,
    "items": []interface{}{
        map[string]interface{}{"name": "Item 1", "quantity": 5},
        map[string]interface{}{"name": "Item 2", "quantity": 3},
    },
}

// Parse and render the template
tokens := lexer.New(content).Tokenize()
ast, _ := parser.New(tokens).Parse()
result, _ := renderer.New(ast, context).Render()
```

## Implementation Details

The template engine follows a three-phase process:

1. **Lexing**: Converts raw template text into tokens
2. **Parsing**: Transforms tokens into an Abstract Syntax Tree (AST)
3. **Rendering**: Evaluates the AST with provided context to produce final output

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests.

## License

This project is open source and available under the [MIT License](LICENSE).
