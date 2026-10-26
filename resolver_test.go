package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolverUsesLexicalBindingForClosures(t *testing.T) {
	var output bytes.Buffer
	source := `
		var a = "global";
		{
			fun showA() { print a; }
			var a = "block";
			showA();
		}
	`
	statements, err := parseSource(source)
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}
	interpreter := NewInterpreter(&output)
	if err := NewResolver(interpreter).Resolve(statements); err != nil {
		t.Fatalf("resolve source: %v", err)
	}
	if err := interpreter.Interpret(statements); err != nil {
		t.Fatalf("interpret source: %v", err)
	}
	if got, want := output.String(), "global\n"; got != want {
		t.Fatalf("resolved closure output: got %q, want %q", got, want)
	}
}

func TestResolverRejectsLocalReadInOwnInitializer(t *testing.T) {
	statements, err := parseSource("{ var a = a; }")
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}
	err = NewResolver(NewInterpreter(nil)).Resolve(statements)
	if err == nil || !strings.Contains(err.Error(), "Can't read local variable in its own initializer.") {
		t.Fatalf("resolver error: got %v", err)
	}
}

func TestResolverRejectsDuplicateLocalDeclarations(t *testing.T) {
	statements, err := parseSource("{ var a; var a; }")
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}
	err = NewResolver(NewInterpreter(nil)).Resolve(statements)
	if err == nil || !strings.Contains(err.Error(), "Already a variable with this name in this scope.") {
		t.Fatalf("resolver error: got %v", err)
	}
}

func TestResolverRejectsTopLevelReturn(t *testing.T) {
	statements, err := parseSource("return 1;")
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}
	err = NewResolver(NewInterpreter(nil)).Resolve(statements)
	if err == nil || !strings.Contains(err.Error(), "Can't return from top-level code.") {
		t.Fatalf("resolver error: got %v", err)
	}
}

func TestResolverRejectsThisOutsideClass(t *testing.T) {
	statements, err := parseSource("print this;")
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}
	err = NewResolver(NewInterpreter(nil)).Resolve(statements)
	if err == nil || !strings.Contains(err.Error(), "Can't use 'this' outside of a class.") {
		t.Fatalf("resolver error: got %v", err)
	}
}

func TestResolverRejectsValueReturnFromInitializer(t *testing.T) {
	statements, err := parseSource("class A { init() { return 1; } }")
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}
	err = NewResolver(NewInterpreter(nil)).Resolve(statements)
	if err == nil || !strings.Contains(err.Error(), "Can't return a value from an initializer.") {
		t.Fatalf("resolver error: got %v", err)
	}
}

func TestRunSourceDoesNotExecuteAfterResolverError(t *testing.T) {
	var output bytes.Buffer
	previous := interpreter
	interpreter = NewInterpreter(&output)
	defer func() {
		interpreter = previous
	}()

	err := runSource(`print "before"; { var a = a; print "after"; }`)
	if err == nil {
		t.Fatal("runSource succeeded, want resolver error")
	}
	if got, want := output.String(), ""; got != want {
		t.Fatalf("output after resolver error: got %q, want empty output", got)
	}
}
