package main

import (
	"os"
	"strings"
	"testing"
)

func interpretSource(source string) (any, error) {
	expression, err := parseSource(source)
	if err != nil {
		return nil, err
	}
	return (&Interpreter{}).Interpret(expression)
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
	interpreter := &Interpreter{}
	invalidExpression, err := parseSource(`-"text"`)
	if err != nil {
		t.Fatalf("parse invalid expression: %v", err)
	}
	if _, err := interpreter.Interpret(invalidExpression); err == nil {
		t.Fatal("invalid expression succeeded, want a runtime error")
	}

	validExpression, err := parseSource("1 + 2")
	if err != nil {
		t.Fatalf("parse valid expression: %v", err)
	}
	value, err := interpreter.Interpret(validExpression)
	if err != nil {
		t.Fatalf("interpreter remained in an error state: %v", err)
	}
	if value != float64(3) {
		t.Fatalf("got %#v, want 3", value)
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
	if err := runSource("1 + 2"); err != nil {
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
		{name: "success", source: "1 + 2", status: 0},
		{name: "parse error", source: "1 +", status: 65},
		{name: "runtime error", source: "1 + true", status: 70},
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
