package main

import (
	"bytes"
	"strings"
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

func TestParserBuildsBlockStatements(t *testing.T) {
	statements, err := parseSource("{ var a = 1; print a; }")
	if err != nil {
		t.Fatalf("parse block: %v", err)
	}
	if len(statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(statements))
	}
	block, ok := statements[0].(*BlockStmt)
	if !ok {
		t.Fatalf("statement type: got %T, want *BlockStmt", statements[0])
	}
	if len(block.Statements) != 2 {
		t.Fatalf("block statements: got %d, want 2", len(block.Statements))
	}
}

func TestInterpreterScopesBlockVariables(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("{ var a = \"inside\"; print a; }")
	if err != nil {
		t.Fatalf("parse block: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret block: %v", err)
	}
	if got, want := output.String(), "inside\n"; got != want {
		t.Fatalf("block output: got %q, want %q", got, want)
	}
}

func TestInterpreterDoesNotExposeBlockVariables(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("{ var a = \"inside\"; } print a;")
	if err != nil {
		t.Fatalf("parse block: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	err = statementInterpreter.Interpret(statements)
	if err == nil {
		t.Fatal("interpret succeeded, want an undefined variable error")
	}
	if got, want := err.Error(), "[line 1] RuntimeError at 'a': Undefined variable 'a'."; got != want {
		t.Fatalf("runtime error: got %q, want %q", got, want)
	}
}

func TestInterpreterSupportsShadowingAndEnclosingLookup(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var a = \"global\"; { var a = \"local\"; print a; } print a;")
	if err != nil {
		t.Fatalf("parse scopes: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret scopes: %v", err)
	}
	if got, want := output.String(), "local\nglobal\n"; got != want {
		t.Fatalf("scope output: got %q, want %q", got, want)
	}

	output.Reset()
	statements, err = parseSource("var global = \"outside\"; { var local = \"inside\"; print global + local; }")
	if err != nil {
		t.Fatalf("parse enclosing lookup: %v", err)
	}
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret enclosing lookup: %v", err)
	}
	if got, want := output.String(), "outsideinside\n"; got != want {
		t.Fatalf("enclosing lookup output: got %q, want %q", got, want)
	}
}

func TestInterpreterAssignsNearestEnclosingVariable(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var a = \"global\"; { a = \"updated\"; } print a;")
	if err != nil {
		t.Fatalf("parse enclosing assignment: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	if err := statementInterpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret enclosing assignment: %v", err)
	}
	if got, want := output.String(), "updated\n"; got != want {
		t.Fatalf("enclosing assignment output: got %q, want %q", got, want)
	}
}

func TestInterpreterRestoresEnvironmentAfterBlockError(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("{ print missing; } print missing;")
	if err != nil {
		t.Fatalf("parse block error: %v", err)
	}

	statementInterpreter := NewInterpreter(&output)
	err = statementInterpreter.Interpret(statements)
	if err == nil {
		t.Fatal("interpret succeeded, want a runtime error")
	}
	if got, want := err.Error(), "[line 1] RuntimeError at 'missing': Undefined variable 'missing'."; got != want {
		t.Fatalf("runtime error: got %q, want %q", got, want)
	}
}

func TestParserBuildsIfStatement(t *testing.T) {
	statements, err := parseSource("if (true) print \"yes\"; else print \"no\";")
	if err != nil {
		t.Fatalf("parse if statement: %v", err)
	}
	if len(statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(statements))
	}
	ifStatement, ok := statements[0].(*IfStmt)
	if !ok {
		t.Fatalf("statement type: got %T, want *IfStmt", statements[0])
	}
	if ifStatement.ElseBranch == nil {
		t.Fatal("else branch is nil, want a branch")
	}
}

func TestInterpreterExecutesSelectedIfBranch(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "then", source: `if (true) print "then"; else print "else";`, want: "then\n"},
		{name: "else", source: `if (false) print "then"; else print "else";`, want: "else\n"},
		{name: "nil is falsey", source: `if (nil) print "then"; else print "else";`, want: "else\n"},
		{name: "number is truthy", source: `if (0) print "then"; else print "else";`, want: "then\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			statements, err := parseSource(test.source)
			if err != nil {
				t.Fatalf("parse if statement: %v", err)
			}
			if err := NewInterpreter(&output).Interpret(statements); err != nil {
				t.Fatalf("interpret if statement: %v", err)
			}
			if got := output.String(); got != test.want {
				t.Fatalf("if output: got %q, want %q", got, test.want)
			}
		})
	}
}

func TestInterpreterIfSupportsBlockBranches(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource(`if (true) { var message = "yes"; print message; } else { print "no"; }`)
	if err != nil {
		t.Fatalf("parse block branches: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret block branches: %v", err)
	}
	if got, want := output.String(), "yes\n"; got != want {
		t.Fatalf("block branch output: got %q, want %q", got, want)
	}
}

func TestParserBindsElseToNearestIf(t *testing.T) {
	statements, err := parseSource(`if (true) if (false) print "inner"; else print "nearest";`)
	if err != nil {
		t.Fatalf("parse nested if: %v", err)
	}
	var output bytes.Buffer
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret nested if: %v", err)
	}
	if got, want := output.String(), "nearest\n"; got != want {
		t.Fatalf("dangling else output: got %q, want %q", got, want)
	}
}

func TestInterpreterSkipsUnselectedIfBranch(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource(`if (true) print "selected"; else print missing;`)
	if err != nil {
		t.Fatalf("parse skipped branch: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret skipped branch: %v", err)
	}
	if got, want := output.String(), "selected\n"; got != want {
		t.Fatalf("skipped branch output: got %q, want %q", got, want)
	}
}

func TestParserReportsIfSyntaxErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message string
	}{
		{name: "missing opening parenthesis", source: "if true print 1;", message: "Expect '(' after 'if'."},
		{name: "missing closing parenthesis", source: "if (true print 1;", message: "Expect ')' after if condition."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseSource(test.source); err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("parse error: got %v, want %q", err, test.message)
			}
		})
	}
}

func TestParserBuildsWhileStatement(t *testing.T) {
	statements, err := parseSource("while (true) print \"loop\";")
	if err != nil {
		t.Fatalf("parse while statement: %v", err)
	}
	if len(statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(statements))
	}
	whileStatement, ok := statements[0].(*WhileStmt)
	if !ok {
		t.Fatalf("statement type: got %T, want *WhileStmt", statements[0])
	}
	if whileStatement.Body == nil {
		t.Fatal("while body is nil")
	}
}

func TestInterpreterExecutesWhileLoop(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource("var i = 0; while (i < 3) { print i; i = i + 1; }")
	if err != nil {
		t.Fatalf("parse while loop: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret while loop: %v", err)
	}
	if got, want := output.String(), "0\n1\n2\n"; got != want {
		t.Fatalf("while output: got %q, want %q", got, want)
	}
}

func TestInterpreterSkipsFalseyWhileLoop(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource(`while (false) print missing;`)
	if err != nil {
		t.Fatalf("parse falsey while loop: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret falsey while loop: %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("falsey while output: got %q, want empty", output.String())
	}
}

func TestInterpreterWhileSupportsNestedControlFlow(t *testing.T) {
	var output bytes.Buffer
	source := `var i = 0; while (i < 2) { if (i == 1) print "one"; else print "zero"; i = i + 1; }`
	statements, err := parseSource(source)
	if err != nil {
		t.Fatalf("parse nested while control flow: %v", err)
	}
	if err := NewInterpreter(&output).Interpret(statements); err != nil {
		t.Fatalf("interpret nested while control flow: %v", err)
	}
	if got, want := output.String(), "zero\none\n"; got != want {
		t.Fatalf("nested while output: got %q, want %q", got, want)
	}
}

func TestInterpreterStopsWhileLoopOnRuntimeError(t *testing.T) {
	var output bytes.Buffer
	statements, err := parseSource(`var i = 0; while (i < 2) { print i; i = i + missing; } print "after";`)
	if err != nil {
		t.Fatalf("parse while runtime error: %v", err)
	}
	err = NewInterpreter(&output).Interpret(statements)
	if err == nil {
		t.Fatal("interpret succeeded, want runtime error")
	}
	if got, want := err.Error(), "[line 1] RuntimeError at 'missing': Undefined variable 'missing'."; got != want {
		t.Fatalf("runtime error: got %q, want %q", got, want)
	}
	if got, want := output.String(), "0\n"; got != want {
		t.Fatalf("output after while runtime error: got %q, want %q", got, want)
	}
}

func TestParserReportsWhileSyntaxErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message string
	}{
		{name: "missing opening parenthesis", source: "while true print 1;", message: "Expect '(' after 'while'."},
		{name: "missing closing parenthesis", source: "while (true print 1;", message: "Expect ')' after condition."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseSource(test.source); err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("parse error: got %v, want %q", err, test.message)
			}
		})
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
