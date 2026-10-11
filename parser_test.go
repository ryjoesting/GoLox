package main

import (
	"strings"
	"testing"
)

func assertParsedExpression(t *testing.T, source string, want string) {
	t.Helper()

	expression, err := parseExpressionSource(source)
	if err != nil {
		t.Fatalf("parse %q: unexpected error: %v", source, err)
	}
	if got := (AstPrinter{}).Print(expression); got != want {
		t.Fatalf("parse %q: got %q, want %q", source, got, want)
	}
}

func parseExpressionSource(source string) (Expr, error) {
	statements, err := parseSource(source + ";")
	if err != nil {
		return nil, err
	}
	if len(statements) != 1 {
		return nil, ParseError{Message: "Expected one expression statement."}
	}
	statement, ok := statements[0].(*ExpressionStmt)
	if !ok {
		return nil, ParseError{Message: "Expected an expression statement."}
	}
	return statement.Expression, nil
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
	assertParsedExpression(t, "a = b = 3", "(= a (= b 3))")
}

func TestParserCommaAndConditionalPrecedence(t *testing.T) {
	assertParsedExpression(t, "1 + 2, 3 * 4", "(, (+ 1 2) (* 3 4))")
	assertParsedExpression(t, "true ? 1, 2 : 3", "(?: true (, 1 2) 3)")
	assertParsedExpression(t, "1 == 2 ? 3 : 4", "(?: (== 1 2) 3 4)")
	assertParsedExpression(t, "a = 1, 2", "(= a (, 1 2))")
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
		{name: "missing operand", source: "1 +", message: "Expect expression.", context: "at end"},
		{name: "missing parenthesis", source: "(1 + 2", message: "Expect ')' after expression.", context: "at end"},
		{name: "invalid primary", source: ")", message: "Expect expression.", context: "at ')'"},
		{name: "missing semicolon", source: "1", message: "Expect ';' after expression.", context: "at end"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parseSource(test.source)
			if len(expression) != 0 {
				t.Fatalf("parse %q returned statements after an error: %#v", test.source, expression)
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
			if len(expression) != 0 {
				t.Fatalf("parse %q returned statements after an error: %#v", source, expression)
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
			if len(expression) != 0 {
				t.Fatalf("parse %q returned statements after an error: %#v", test.source, expression)
			}
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("parse %q error: got %v, want %q", test.source, err, test.message)
			}
		})
	}
}

func TestParserAssignmentErrors(t *testing.T) {
	tests := []string{"1 = 2", "(a) = 3", "a + b = c"}

	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			statements, err := parseSource(source + ";")
			if len(statements) != 0 {
				t.Fatalf("parse %q returned statements after an error: %#v", source, statements)
			}
			if err == nil || !strings.Contains(err.Error(), "Invalid assignment target.") {
				t.Fatalf("parse %q error: got %v, want invalid assignment target", source, err)
			}
		})
	}
}

func TestParseSourceWiresScannerAndParser(t *testing.T) {
	expression, err := parseExpressionSource("1 + 2 * 3")
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
	statements, err := parseSource("1 +")
	if len(statements) != 0 {
		t.Fatalf("parseSource returned statements after an error: %#v", statements)
	}
	if err == nil {
		t.Fatal("parseSource succeeded, want an error")
	}
}
