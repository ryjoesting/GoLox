package main

import (
	"strings"
	"testing"
)

func assertParsedExpression(t *testing.T, source string, want string) {
	t.Helper()

	expression, err := parseSource(source)
	if err != nil {
		t.Fatalf("parse %q: unexpected error: %v", source, err)
	}
	if got := (AstPrinter{}).Print(expression); got != want {
		t.Fatalf("parse %q: got %q, want %q", source, got, want)
	}
}

func TestParserPrecedence(t *testing.T) {
	assertParsedExpression(t, "1 + 2 * 3", "(+ 1 (* 2 3))")
	assertParsedExpression(t, "1 == 2 < 3", "(== 1 (< 2 3))")
}

func TestParserAssociativity(t *testing.T) {
	assertParsedExpression(t, "1 - 2 - 3", "(- (- 1 2) 3)")
	assertParsedExpression(t, "--1", "(- (- 1))")
	assertParsedExpression(t, "1, 2, 3", "(, (, 1 2) 3)")
	assertParsedExpression(t, "true ? 1 : false ? 2 : 3", "(?: true 1 (?: false 2 3))")
}

func TestParserCommaAndConditionalPrecedence(t *testing.T) {
	assertParsedExpression(t, "1 + 2, 3 * 4", "(, (+ 1 2) (* 3 4))")
	assertParsedExpression(t, "true ? 1, 2 : 3", "(?: true (, 1 2) 3)")
	assertParsedExpression(t, "1 == 2 ? 3 : 4", "(?: (== 1 2) 3 4)")
}

func TestParserPrimaryExpressions(t *testing.T) {
	assertParsedExpression(t, "false", "false")
	assertParsedExpression(t, "true", "true")
	assertParsedExpression(t, "nil", "nil")
	assertParsedExpression(t, `"hello"`, "hello")
	assertParsedExpression(t, "(1 + 2)", "(group (+ 1 2))")
}

func TestParserSyntaxErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message string
		context string
	}{
		{name: "empty input", source: "", message: "Expect expression.", context: "at end"},
		{name: "missing operand", source: "1 +", message: "Expect expression.", context: "at end"},
		{name: "missing parenthesis", source: "(1 + 2", message: "Expect ')' after expression.", context: "at end"},
		{name: "invalid primary", source: ")", message: "Expect expression.", context: "at ')'"},
		{name: "trailing token", source: "1 2", message: "Expect end of expression.", context: "at '2'"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parseSource(test.source)
			if expression != nil {
				t.Fatalf("parse %q returned an expression after an error: %#v", test.source, expression)
			}
			if err == nil {
				t.Fatalf("parse %q succeeded, want an error", test.source)
			}
			if !strings.Contains(err.Error(), test.message) {
				t.Errorf("error %q does not contain %q", err, test.message)
			}
			if !strings.Contains(err.Error(), test.context) {
				t.Errorf("error %q does not contain %q", err, test.context)
			}
		})
	}
}

func TestParserLeadingBinaryOperatorErrors(t *testing.T) {
	tests := []string{"+ 1", "== 1", "< 1", "* 3", "/ 3", ", 4", "1 + * 2"}

	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			expression, err := parseSource(source)
			if expression != nil {
				t.Fatalf("parse %q returned an expression after an error: %#v", source, expression)
			}
			if err == nil {
				t.Fatalf("parse %q succeeded, want an error", source)
			}
			if !strings.Contains(err.Error(), "Expect expression before binary operator.") {
				t.Errorf("error %q does not identify a missing left operand", err)
			}
		})
	}
}

func TestParserConditionalErrors(t *testing.T) {
	tests := []struct {
		source  string
		message string
	}{
		{source: "true ? 1", message: "Expect ':' after expression."},
		{source: "true ? : 1", message: "Expect expression."},
		{source: "true ? 1 :", message: "Expect expression."},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			expression, err := parseSource(test.source)
			if expression != nil {
				t.Fatalf("parse %q returned an expression after an error: %#v", test.source, expression)
			}
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("parse %q error: got %v, want %q", test.source, err, test.message)
			}
		})
	}
}

func TestParseSourceWiresScannerAndParser(t *testing.T) {
	expression, err := parseSource("1 + 2 * 3")
	if err != nil {
		t.Fatalf("parseSource returned an unexpected error: %v", err)
	}

	got := (AstPrinter{}).Print(expression)
	want := "(+ 1 (* 2 3))"
	if got != want {
		t.Fatalf("parseSource output: got %q, want %q", got, want)
	}
}

func TestParseSourceReturnsErrorsWithoutAnExpression(t *testing.T) {
	expression, err := parseSource("1 +")
	if expression != nil {
		t.Fatalf("parseSource returned an expression after an error: %#v", expression)
	}
	if err == nil {
		t.Fatal("parseSource succeeded, want an error")
	}
}
