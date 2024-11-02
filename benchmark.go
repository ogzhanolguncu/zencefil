package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ogzhanolguncu/zencefil/lexer"
	"github.com/ogzhanolguncu/zencefil/parser"
	"github.com/ogzhanolguncu/zencefil/renderer"
)

const (
	WARMUP_ITERATIONS    = 100
	BENCHMARK_ITERATIONS = 1000
)

type TestCase struct {
	name     string
	template string
	context  map[string]interface{}
}

func runBenchmark(template string, context map[string]interface{}) (time.Duration, error) {
	// Parse template
	l := lexer.New(template)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	ast, err := p.Parse()
	if err != nil {
		return 0, fmt.Errorf("parse error: %v", err)
	}

	// Warmup
	for i := 0; i < WARMUP_ITERATIONS; i++ {
		r := renderer.New(ast, context)
		_, err := r.Render()
		if err != nil {
			return 0, fmt.Errorf("render error during warmup: %v", err)
		}
	}

	// Benchmark
	start := time.Now()
	for i := 0; i < BENCHMARK_ITERATIONS; i++ {
		r := renderer.New(ast, context)
		_, err := r.Render()
		if err != nil {
			return 0, fmt.Errorf("render error during benchmark: %v", err)
		}
	}
	elapsed := time.Since(start)

	return elapsed, nil
}

func benchmarkGo() {
	testCases := []TestCase{
		{
			name:     "simple_text",
			template: "Hello, world!",
			context:  map[string]interface{}{},
		},
		{
			name:     "variable_substitution",
			template: "Hello, {{ name }}!",
			context: map[string]interface{}{
				"name": "John",
			},
		},
		{
			name:     "nested_object_access",
			template: "Hello, {{ user['name'] }}! Your age is {{ user['age'] }}",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
		},
		{
			name:     "complex_conditions",
			template: `{{ if age >= 18 && has_license }}Can drive{{ elif age >= 16 }}Can get learner's permit{{ else }}Too young to drive{{ endif }}`,
			context: map[string]interface{}{
				"age":         17,
				"has_license": false,
			},
		},
		{
			name:     "simple_loop",
			template: `{{ for name in names }}{{ name }}, {{ endfor }}`,
			context: map[string]interface{}{
				"names": []interface{}{"John", "Jane", "Bob", "Alice"},
			},
		},
		{
			name:     "complex_loop",
			template: `{{ for item in items }}- {{ item['name'] }}: ${{ item['price'] }}{{ endfor }}`,
			context: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "Apple", "price": 0.5},
					map[string]interface{}{"name": "Banana", "price": 0.3},
					map[string]interface{}{"name": "Orange", "price": 0.6},
				},
			},
		},
	}

	fmt.Printf("\nZencefil Benchmark Results\n")
	fmt.Printf("Warmup iterations: %d\n", WARMUP_ITERATIONS)
	fmt.Printf("Benchmark iterations: %d\n\n", BENCHMARK_ITERATIONS)
	fmt.Printf("%-25s %-15s %-15s %-15s\n", "Test Case", "Total Time", "Avg per op", "ops/sec")
	fmt.Printf("%s\n", strings.Repeat("-", 70))

	for _, tc := range testCases {
		elapsed, err := runBenchmark(tc.template, tc.context)
		if err != nil {
			fmt.Printf("Error in %s: %v\n", tc.name, err)
			continue
		}

		opsPerSec := float64(BENCHMARK_ITERATIONS) / elapsed.Seconds()
		avgPerOp := elapsed.Nanoseconds() / int64(BENCHMARK_ITERATIONS)

		fmt.Printf("%-25s %-15s %-15s %-15.2f\n",
			tc.name,
			elapsed.Round(time.Microsecond),
			fmt.Sprintf("%d ns", avgPerOp),
			opsPerSec,
		)
	}
}
