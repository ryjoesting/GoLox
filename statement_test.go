package main

import (
	"bytes"
	"testing"
)

func TestParserBuildsStatements(t *testing.T) {
	statements, err := parseSource("var beverage = \"espresso\"; print beverage;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}
	if len(statements) != 2 {
		t.Fatalf("got %d statements, want 2", len(statements))
	}

	varStatement, ok := statements[0].(*VarStmt)
	if !ok {
		t.Fatalf("statement 0 type: got %T, want *VarStmt", statements[0])
	}
	if got := varStatement.Name.Lexeme; got != "beverage" {
		t.Fatalf("variable name: got %q, want %q", got, "beverage")
	}
	if got := (AstPrinter{}).Print(varStatement.Initializer); got != "espresso" {
		t.Fatalf("variable initializer: got %q, want %q", got, "espresso")
	}

	printStatement, ok := statements[1].(*PrintStmt)
	if !ok {
		t.Fatalf("statement 1 type: got %T, want *PrintStmt", statements[1])
	}
	if got := (AstPrinter{}).Print(printStatement.Expression); got != "beverage" {
		t.Fatalf("print statement: got %q, want %q", got, "beverage")
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

func TestParserAcceptsUninitializedVariable(t *testing.T) {
	statements, err := parseSource("var beverage;")
	if err != nil {
		t.Fatalf("parse declaration: %v", err)
	}
	varStatement, ok := statements[0].(*VarStmt)
	if !ok {
		t.Fatalf("statement type: got %T, want *VarStmt", statements[0])
	}
	if varStatement.Initializer != nil {
		t.Fatalf("initializer: got %T, want nil", varStatement.Initializer)
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

func TestInterpreterExecutesGlobalVariables(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var beverage = \"espresso\"; print beverage; var empty; print empty;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret statements: %v", err)
	}
	if got, want := output.String(), "espresso\nnil\n"; got != want {
		t.Fatalf("global variable output: got %q, want %q", got, want)
	}
}

func TestInterpreterGlobalsPersistAcrossInterpretCalls(t *testing.T) {
	var output bytes.Buffer
	statementInterpreter := NewInterpreter(&output)

	first, err := parseSource("var beverage = \"espresso\";")
	if err != nil {
		t.Fatalf("parse declaration: %v", err)
	}
	if err := statementInterpreter.Interpret(first); err != nil {
		t.Fatalf("interpret declaration: %v", err)
	}

	second, err := parseSource("print beverage;")
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if err := statementInterpreter.Interpret(second); err != nil {
		t.Fatalf("interpret access: %v", err)
	}
	if got, want := output.String(), "espresso\n"; got != want {
		t.Fatalf("persistent global output: got %q, want %q", got, want)
	}
}

func TestInterpreterAllowsGlobalRedefinition(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var a = \"before\"; print a; var a = \"after\"; print a;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret statements: %v", err)
	}
	if got, want := output.String(), "before\nafter\n"; got != want {
		t.Fatalf("redefined global output: got %q, want %q", got, want)
	}
}

func TestInterpreterAssignsGlobalVariables(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var a = 1; print a = 2; print a;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret statements: %v", err)
	}
	if got, want := output.String(), "2\n2\n"; got != want {
		t.Fatalf("assignment output: got %q, want %q", got, want)
	}
}

func TestInterpreterAssignsRightAssociatively(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var a; var b; print a = b = 3; print a; print b;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret statements: %v", err)
	}
	if got, want := output.String(), "3\n3\n3\n"; got != want {
		t.Fatalf("chained assignment output: got %q, want %q", got, want)
	}
}

func TestInterpreterRejectsUndefinedAssignment(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("a = 1;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	err = statementInterpreter.Interpret(statements)
	if err == nil {
		t.Fatal("interpret succeeded, want an undefined assignment error")
	}
	if got, want := err.Error(), "[line 1] RuntimeError at 'a': Undefined variable 'a'."; got != want {
		t.Fatalf("runtime error: got %q, want %q", got, want)
	}
}

func TestInterpreterReportsUndefinedGlobal(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("print beverage; print 2;")
	if err != nil {
		t.Fatalf("parse statements: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	err = statementInterpreter.Interpret(statements)
	if err == nil {
		t.Fatal("interpret succeeded, want an undefined variable error")
	}
	if got, want := err.Error(), "[line 1] RuntimeError at 'beverage': Undefined variable 'beverage'."; got != want {
		t.Fatalf("runtime error: got %q, want %q", got, want)
	}
	if output.Len() != 0 {
		t.Fatalf("output after undefined variable: got %q, want empty", output.String())
	}
}
