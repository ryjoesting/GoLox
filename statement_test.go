package main

import (
	"bytes"
	"testing"
)

func TestParserBuildsStatements(t *testing.T) {
	statements, err := parseSource("1 + 2; print \"hello\";")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}
	if len(statements) != 2 {
		t.Fatalf("got %d statements, want 2", len(statements))
	}

	expressionStatement, ok := statements[0].(*ExpressionStmt)
	if !ok {
		t.Fatalf("statement 0 type: got %T, want *ExpressionStmt", statements[0])
	}
	if got := (AstPrinter{}).Print(expressionStatement.Expression); got != "(+ 1 2)" {
		t.Fatalf("expression statement: got %q, want %q", got, "(+ 1 2)")
	}

	printStatement, ok := statements[1].(*PrintStmt)
	if !ok {
		t.Fatalf("statement 1 type: got %T, want *PrintStmt", statements[1])
	}
	if got := (AstPrinter{}).Print(printStatement.Expression); got != "hello" {
		t.Fatalf("print statement: got %q, want %q", got, "hello")
	}
}

func TestParserAcceptsEmptyProgram(t *testing.T) {
	statements, err := parseSource("")
	if err != nil {
		t.Fatalf("parse empty program: %v", err)
	}
	if len(statements) != 0 {
		t.Fatalf("got %d statements, want 0", len(statements))
	}
}

func TestParserSynchronizesAfterStatementError(t *testing.T) {
	statements, err := parseSource("1 + ; print 2;")
	if err == nil {
		t.Fatal("parse succeeded, want a syntax error")
	}
	if len(statements) != 1 {
		t.Fatalf("got %d recovered statements, want 1", len(statements))
	}
	if _, ok := statements[0].(*PrintStmt); !ok {
		t.Fatalf("recovered statement type: got %T, want *PrintStmt", statements[0])
	}
}

func TestInterpreterExecutesStatements(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("1 + 2; print 3 * 4;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret statements: %v", err)
	}
	if got, want := output.String(), "12\n"; got != want {
		t.Fatalf("statement output: got %q, want %q", got, want)
	}
}

func TestInterpreterStopsAfterStatementRuntimeError(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("print 1; print -\"bad\"; print 3;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err == nil {
		t.Fatal("interpret succeeded, want a runtime error")
	}
	if got, want := output.String(), "1\n"; got != want {
		t.Fatalf("statement output after error: got %q, want %q", got, want)
	}
}
