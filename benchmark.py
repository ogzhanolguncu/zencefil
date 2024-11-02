import time
from jinja2 import Template

WARMUP_ITERATIONS = 100
BENCHMARK_ITERATIONS = 1000

class TestCase:
    def __init__(self, name, template, context):
        self.name = name
        self.template = template
        self.context = context

def run_benchmark(template_str, context):
    template = Template(template_str)
    
    # Warmup phase
    for _ in range(WARMUP_ITERATIONS):
        template.render(**context)
    
    # Benchmark phase
    start_time = time.perf_counter()
    for _ in range(BENCHMARK_ITERATIONS):
        template.render(**context)
    total_time = time.perf_counter() - start_time
    
    ops_per_sec = BENCHMARK_ITERATIONS / total_time
    avg_per_op = (total_time * 1e9) / BENCHMARK_ITERATIONS  # Convert to nanoseconds
    
    return total_time, avg_per_op, ops_per_sec

def main():
    test_cases = [
        TestCase(
            "simple_text",
            "Hello, world!",
            {}
        ),
        TestCase(
            "variable_substitution",
            "Hello, {{ name }}!",
            {"name": "John"}
        ),
        TestCase(
            "nested_object_access",
            "Hello, {{ user.name }}! Your age is {{ user.age }}",
            {"user": {"name": "John", "age": 30}}
        ),
        TestCase(
            "complex_conditions",
            """{% if age >= 18 and has_license %}Can drive{% elif age >= 16 %}Can get learner's permit{% else %}Too young to drive{% endif %}""",
            {"age": 17, "has_license": False}
        ),
        TestCase(
            "simple_loop",
            """{% for name in names %}{{ name }}, {% endfor %}""",
            {"names": ["John", "Jane", "Bob", "Alice"]}
        ),
        TestCase(
            "complex_loop",
            """{% for item in items %}- {{ item.name }}: ${{ item.price }}{% endfor %}""",
            {"items": [
                {"name": "Apple", "price": 0.5},
                {"name": "Banana", "price": 0.3},
                {"name": "Orange", "price": 0.6}
            ]}
        ),
    ]

    print("\nJinja2 Benchmark Results")
    print(f"Warmup iterations: {WARMUP_ITERATIONS}")
    print(f"Benchmark iterations: {BENCHMARK_ITERATIONS}\n")
    print(f"{'Test Case':<25} {'Total Time':<15} {'Avg per op':<15} {'ops/sec':<15}")
    print("-" * 70)

    for test in test_cases:
        try:
            total_time, avg_per_op, ops_per_sec = run_benchmark(test.template, test.context)
            print(f"{test.name:<25} {total_time*1000:.<15.3f}ms {avg_per_op:.<15.0f}ns {ops_per_sec:<15.2f}")
        except Exception as e:
            print(f"Error in {test.name}: {str(e)}")

if __name__ == "__main__":
    main()
