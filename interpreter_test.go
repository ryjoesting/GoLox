package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func interpretSource(source string) (any, error) {
	expression, err := parseExpressionSource(source)
	if err != nil {
		return nil, err
	}
	interpreter := NewInterpreter(io.Discard)
	value := interpreter.evaluate(expression)
	if interpreter.runtimeError != nil {
		return nil, interpreter.runtimeError
	}
	return value, nil
}

func TestInterpreterEvaluatesExpressions(t *testing.T) {
	tests := []struct {
		source string
		want   any
	}{
		{source: "123", want: float64(123)},
		{source: `"123"`, want: "123"},
		{source: "(1 + 2) * 3", want: float64(9)},
		{source: "-5", want: float64(-5)},
		{source: "!nil", want: true},
		{source: "1 + 2 * 3", want: float64(7)},
		{source: `"hello" + " world"`, want: "hello world"},
		{source: "1 < 2", want: true},
		{source: "nil == nil", want: true},
		{source: "false == nil", want: false},
		{source: "1, 2, 3", want: float64(3)},
		{source: "true ? 1 : 2", want: float64(1)},
		{source: "false ? 1 : 2", want: float64(2)},
		{source: `"hi" or 2`, want: "hi"},
		{source: `nil or "yes"`, want: "yes"},
		{source: `false and "no"`, want: false},
		{source: `true and "yes"`, want: "yes"},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			got, err := interpretSource(test.source)
			if err != nil {
				t.Fatalf("interpret %q: unexpected error: %v", test.source, err)
			}
			if got != test.want {
				t.Fatalf("interpret %q: got %#v, want %#v", test.source, got, test.want)
			}
		})
	}
}

func TestInterpreterLogicalOperatorsShortCircuit(t *testing.T) {
	tests := []struct {
		source string
		want   any
	}{
		{source: `true or missing`, want: true},
		{source: `false and missing`, want: false},
		{source: `nil or "fallback"`, want: "fallback"},
		{source: `true and 42`, want: float64(42)},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			got, err := interpretSource(test.source)
			if err != nil {
				t.Fatalf("interpret %q: unexpected error: %v", test.source, err)
			}
			if got != test.want {
				t.Fatalf("interpret %q: got %#v, want %#v", test.source, got, test.want)
			}
		})
	}
}

func TestInterpreterConditionalEvaluationIsLazy(t *testing.T) {
	value, err := interpretSource("true ? 1 : -\"invalid\"")
	if err != nil {
		t.Fatalf("interpret returned an unexpected error: %v", err)
	}
	if value != float64(1) {
		t.Fatalf("got %#v, want 1", value)
	}
}

func TestInterpreterExecutesFunctionsAndReturns(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
			fun add(a, b) { return a + b; }
			print add(2, 3);
			print add(4, 5);
		`)
	if err != nil {
		t.Fatalf("parse functions: %v", err)
	}
	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret functions: %v", err)
	}
	if got, want := output.String(), "5\n9\n"; got != want {
		t.Fatalf("function output: got %q, want %q", got, want)
	}
}

func TestInterpreterReturnsNilFromBareReturn(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		fun stop() { return; }
		print stop();
	`)
	if err != nil {
		t.Fatalf("parse bare return: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret bare return: %v", err)
	}
	if got, want := output.String(), "nil\n"; got != want {
		t.Fatalf("bare return output: got %q, want %q", got, want)
	}
}

func TestInterpreterStringifiesFunctions(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource("fun greet() {} print greet;")
	if err != nil {
		t.Fatalf("parse function: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret function: %v", err)
	}
	if got, want := output.String(), "<fn greet>\n"; got != want {
		t.Fatalf("function stringification: got %q, want %q", got, want)
	}
}

func TestInterpreterSupportsClosuresAndRecursion(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
			fun makeCounter() {
				var count = 0;
				fun increment() {
					count = count + 1;
					return count;
				}
				return increment;
			}
			var counter = makeCounter();
			print counter();
			print counter();
			fun factorial(n) {
				if (n <= 1) return 1;
				return n * factorial(n - 1);
			}
			print factorial(5);
		`)
	if err != nil {
		t.Fatalf("parse closure: %v", err)
	}
	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret closure: %v", err)
	}
	if got, want := output.String(), "1\n2\n120\n"; got != want {
		t.Fatalf("closure output: got %q, want %q", got, want)
	}
}

func TestInterpreterSupportsClassesFieldsMethodsAndInitializers(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		class Counter {
			init(start) { this.value = start; }
			increment() { this.value = this.value + 1; }
			current() { return this.value; }
		}
		var counter = Counter(4);
		counter.increment();
		print counter.current();
		print counter;
	`)
	if err != nil {
		t.Fatalf("parse classes: %v", err)
	}
	statementInterpreter := NewInterpreter(&output)
	if err := NewResolver(statementInterpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve classes: %v", err)
	}
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret classes: %v", err)
	}
	if got, want := output.String(), "5\nCounter instance\n"; got != want {
		t.Fatalf("class output: got %q, want %q", got, want)
	}
}

func TestInterpreterBindsMethodsWhenExtracted(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		class Person { sayName() { print this.name; } }
		var jane = Person();
		jane.name = "Jane";
		var method = jane.sayName;
		method();
	`)
	if err != nil {
		t.Fatalf("parse bound method: %v", err)
	}
	statementInterpreter := NewInterpreter(&output)
	if err := NewResolver(statementInterpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve bound method: %v", err)
	}
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret bound method: %v", err)
	}
	if got, want := output.String(), "Jane\n"; got != want {
		t.Fatalf("bound method output: got %q, want %q", got, want)
	}
}

func TestInterpreterFieldsOverrideMethods(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		class Example { value() { return "method"; } }
		var example = Example();
		example.value = "field";
		print example.value;
	`)
	if err != nil {
		t.Fatalf("parse field override: %v", err)
	}
	interpreter := NewInterpreter(&output)
	if err := NewResolver(interpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve field override: %v", err)
	}
	if err := interpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret field override: %v", err)
	}
	if got, want := output.String(), "field\n"; got != want {
		t.Fatalf("field override output: got %q, want %q", got, want)
	}
}

func TestInterpreterSupportsInheritanceAndSuper(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		class A {
			method() { print "A"; }
			value() { return "A value"; }
		}
		class B < A {
			method() {
				super.method();
				print "B";
			}
			value() { return super.value() + " plus B"; }
		}
		class C < B {}
		C().method();
		print C().value();
	`)
	if err != nil {
		t.Fatalf("parse inheritance: %v", err)
	}
	interpreter := NewInterpreter(&output)
	if err := NewResolver(interpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve inheritance: %v", err)
	}
	if err := interpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret inheritance: %v", err)
	}
	if got, want := output.String(), "A\nB\nA value plus B\n"; got != want {
		t.Fatalf("inheritance output: got %q, want %q", got, want)
	}
}

func TestInterpreterBindsExtractedSuperMethod(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		class A { method() { print this.name; } }
		class B < A { getMethod() { return super.method; } }
		var b = B();
		b.name = "B instance";
		var method = b.getMethod();
		method();
	`)
	if err != nil {
		t.Fatalf("parse extracted super method: %v", err)
	}
	interpreter := NewInterpreter(&output)
	if err := NewResolver(interpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve extracted super method: %v", err)
	}
	if err := interpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret extracted super method: %v", err)
	}
	if got, want := output.String(), "B instance\n"; got != want {
		t.Fatalf("extracted super method output: got %q, want %q", got, want)
	}
}

func TestInterpreterInheritsInitializer(t *testing.T) {
	var output strings.Builder
	statements, err := parseSource(`
		class A { init(value) { this.value = value; } }
		class B < A {}
		print B("inherited").value;
	`)
	if err != nil {
		t.Fatalf("parse inherited initializer: %v", err)
	}
	interpreter := NewInterpreter(&output)
	if err := NewResolver(interpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve inherited initializer: %v", err)
	}
	if err := interpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret inherited initializer: %v", err)
	}
	if got, want := output.String(), "inherited\n"; got != want {
		t.Fatalf("inherited initializer output: got %q, want %q", got, want)
	}
}

func TestInterpreterClassErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message string
	}{
		{name: "non-instance get", source: `1.foo;`, message: "Only instances have properties."},
		{name: "non-instance set", source: `1.foo = 2;`, message: "Only instances have fields."},
		{name: "missing property", source: `class A {} print A().missing;`, message: "Undefined property 'missing'."},
		{name: "initializer arity", source: `class A { init(value) {} } A();`, message: "Expected 1 arguments but got 0."},
		{name: "non-class superclass", source: `var value = "not a class"; class A < value {}`, message: "Superclass must be a class."},
		{name: "missing superclass method", source: `class A {} class B < A { method() { super.missing(); } } B().method();`, message: "Undefined property 'missing'."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statements, err := parseSource(test.source)
			if err != nil {
				t.Fatalf("parse %s: %v", test.name, err)
			}
			interpreter := NewInterpreter(io.Discard)
			if err := NewResolver(interpreter).Resolve(statements); err != nil {
				t.Fatalf("resolve %s: %v", test.name, err)
			}
			err = interpreter.Interpret(statements)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error: got %v, want %q", err, test.message)
			}
		})
	}
}

func TestInterpreterFunctionErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message string
	}{
		{name: "non-callable", source: `1();`, message: "Can only call functions and classes."},
		{name: "too few arguments", source: `fun one(a) {} one();`, message: "Expected 1 arguments but got 0."},
		{name: "too many arguments", source: `fun one(a) {} one(1, 2);`, message: "Expected 1 arguments but got 2."},
		{name: "top-level return", source: `return 1;`, message: "Can't return from top-level code."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statements, err := parseSource(test.source)
			if err != nil {
				t.Fatalf("parse %s: %v", test.name, err)
			}
			err = NewInterpreter(io.Discard).Interpret(statements)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error: got %v, want %q", err, test.message)
			}
		})
	}
}

func TestInterpreterRuntimeErrors(t *testing.T) {
	tests := []string{"-\"text\"", "1 + true", "true < false"}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			value, err := interpretSource(source)
			if value != nil {
				t.Fatalf("interpret returned %#v after an error", value)
			}
			if err == nil {
				t.Fatal("interpret succeeded, want a runtime error")
			}
			if !strings.Contains(err.Error(), "RuntimeError") {
				t.Fatalf("error %q is not a runtime error", err)
			}
		})
	}
}

func TestInterpreterReportsNumericOperandErrors(t *testing.T) {
	tests := []struct {
		source  string
		message string
	}{
		{source: `-"text"`, message: "Operand must be a number."},
		{source: `"text" - 1`, message: "Operands must be numbers."},
		{source: `"text" / 1`, message: "Operands must be numbers."},
		{source: `"text" * 1`, message: "Operands must be numbers."},
		{source: `"text" > 1`, message: "Operands must be numbers."},
		{source: `"text" >= 1`, message: "Operands must be numbers."},
		{source: `"text" < 1`, message: "Operands must be numbers."},
		{source: `"text" <= 1`, message: "Operands must be numbers."},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			value, err := interpretSource(test.source)
			if value != nil {
				t.Fatalf("interpret returned %#v after an error", value)
			}
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("interpret error: got %v, want %q", err, test.message)
			}
		})
	}
}

func TestInterpreterRuntimeErrorIncludesTokenContext(t *testing.T) {
	value, err := interpretSource("1 + true")
	if value != nil {
		t.Fatalf("interpret returned %#v after an error", value)
	}
	if err == nil {
		t.Fatal("interpret succeeded, want a runtime error")
	}
	if !strings.Contains(err.Error(), "[line 1]") || !strings.Contains(err.Error(), "at '+'") {
		t.Fatalf("runtime error %q lacks token context", err)
	}
}

func TestInterpreterCanBeReusedAfterRuntimeError(t *testing.T) {
	interpreter := NewInterpreter(io.Discard)
	invalidStatements, err := parseSource(`-"text";`)
	if err != nil {
		t.Fatalf("parse invalid expression: %v", err)
	}
	if err := interpreter.Interpret(invalidStatements); err == nil {
		t.Fatal("invalid expression succeeded, want a runtime error")
	}

	validStatements, err := parseSource("1 + 2;")
	if err != nil {
		t.Fatalf("parse valid expression: %v", err)
	}
	err = interpreter.Interpret(validStatements)
	if err != nil {
		t.Fatalf("interpreter remained in an error state: %v", err)
	}
}

func TestInterpreterRuntimeErrorsDoNotPanic(t *testing.T) {
	sources := []string{`-"text"`, "true * 2", `"text" < "other"`}
	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("interpret panicked: %v", recovered)
				}
			}()

			if _, err := interpretSource(source); err == nil {
				t.Fatal("interpret succeeded, want a runtime error")
			}
		})
	}
}

func TestStringify(t *testing.T) {
	tests := []struct {
		value any
		want  string
	}{
		{value: nil, want: "nil"},
		{value: true, want: "true"},
		{value: float64(3.5), want: "3.5"},
		{value: "text", want: "text"},
	}

	for _, test := range tests {
		if got := stringify(test.value); got != test.want {
			t.Errorf("stringify(%#v): got %q, want %q", test.value, got, test.want)
		}
	}
}

func TestRunSourceReturnsPipelineErrors(t *testing.T) {
	if err := runSource("1 + 2;"); err != nil {
		t.Fatalf("runSource returned an unexpected error: %v", err)
	}
	if err := runSource("1 + true"); err == nil {
		t.Fatal("runSource succeeded, want a runtime error")
	}
	if err := runSource("1 +"); err == nil {
		t.Fatal("runSource succeeded, want a parse error")
	}
}

func TestRunFileStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		status int
	}{
		{name: "success", source: "1 + 2;", status: 0},
		{name: "parse error", source: "1 +", status: 65},
		{name: "runtime error", source: "1 + true;", status: 70},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "golox-*.lox")
			if err != nil {
				t.Fatal(err)
			}
			path := file.Name()
			defer os.Remove(path)
			if _, err := file.WriteString(test.source); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}

			if got := runFile(path); got != test.status {
				t.Fatalf("runFile status: got %d, want %d", got, test.status)
			}
		})
	}

	if got := runFile("missing-file.lox"); got != 65 {
		t.Fatalf("missing file status: got %d, want 65", got)
	}
}
