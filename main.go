package main

import (
	"fmt"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/ogzhanolguncu/zencefil/parser"
)

func main() {
	// content := `Hello, {{ name }}!
	// 			{{ if is_admin }}
	// 			You are an admin.
	// 				{{ if is_super_admin }}
	// 					You are a super admin!
	// 				{{ else }}
	// 					But not a super admin.
	// 				{{ endif }}
	// 			{{ else }}
	// 			You are not an admin.
	// 			{{ endif }}`
	content := `Hello, {{ name }}!
				{{ if is_admin }}
				You are an admin.
				{{ else }}
				You are not an admin.
				{{ endif }}`

	fmt.Printf("%+v", parser.New(lexer.New(content).Tokenize()).Parse(""))
}
