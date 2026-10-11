package main

import (
	"strings"
	"testing"
)

func TestEnvironmentDefinesAndGetsValues(t *testing.T) {
	environment := NewEnvironment()
	name := NewToken(IDENTIFIER, "answer", "answer", 1)
	environment.Define(name.Lexeme, 42.0)

	value, err := environment.Get(name)
	if err != nil {
		t.Fatalf("get defined value: %v", err)
	}
	if got, want := value, 42.0; got != want {
		t.Fatalf("defined value: got %v, want %v", got, want)
	}
}

func TestEnvironmentAllowsRedefinition(t *testing.T) {
	environment := NewEnvironment()
	name := NewToken(IDENTIFIER, "answer", "answer", 1)
	environment.Define(name.Lexeme, "before")
	environment.Define(name.Lexeme, "after")

	value, err := environment.Get(name)
	if err != nil {
		t.Fatalf("get redefined value: %v", err)
	}
	if got, want := value, "after"; got != want {
		t.Fatalf("redefined value: got %v, want %v", got, want)
	}
}

func TestEnvironmentAssignsExistingValues(t *testing.T) {
	environment := NewEnvironment()
	name := NewToken(IDENTIFIER, "answer", "answer", 1)
	environment.Define(name.Lexeme, 1.0)

	if err := environment.Assign(name, 2.0); err != nil {
		t.Fatalf("assign existing value: %v", err)
	}
	value, err := environment.Get(name)
	if err != nil {
		t.Fatalf("get assigned value: %v", err)
	}
	if got, want := value, 2.0; got != want {
		t.Fatalf("assigned value: got %v, want %v", got, want)
	}
}

func TestEnvironmentRejectsAssignmentToUndefinedValue(t *testing.T) {
	environment := NewEnvironment()
	name := NewToken(IDENTIFIER, "missing", "missing", 1)

	err := environment.Assign(name, 2.0)
	if err == nil {
		t.Fatal("assign undefined value succeeded, want an error")
	}
	if !strings.Contains(err.Error(), "Undefined variable 'missing'.") {
		t.Fatalf("undefined assignment error: got %q", err)
	}
}

func TestEnvironmentFindsEnclosingValues(t *testing.T) {
	global := NewEnvironment()
	local := NewEnvironment(global)
	name := NewToken(IDENTIFIER, "answer", "answer", 1)
	global.Define(name.Lexeme, 42.0)

	value, err := local.Get(name)
	if err != nil {
		t.Fatalf("get enclosing value: %v", err)
	}
	if got, want := value, 42.0; got != want {
		t.Fatalf("enclosing value: got %v, want %v", got, want)
	}
}

func TestEnvironmentAssignsEnclosingValues(t *testing.T) {
	global := NewEnvironment()
	local := NewEnvironment(global)
	name := NewToken(IDENTIFIER, "answer", "answer", 1)
	global.Define(name.Lexeme, 1.0)

	if err := local.Assign(name, 2.0); err != nil {
		t.Fatalf("assign enclosing value: %v", err)
	}
	value, err := global.Get(name)
	if err != nil {
		t.Fatalf("get assigned enclosing value: %v", err)
	}
	if got, want := value, 2.0; got != want {
		t.Fatalf("assigned enclosing value: got %v, want %v", got, want)
	}
}

func TestEnvironmentStoresNil(t *testing.T) {
	environment := NewEnvironment()
	name := NewToken(IDENTIFIER, "empty", "empty", 1)
	environment.Define(name.Lexeme, nil)

	value, err := environment.Get(name)
	if err != nil {
		t.Fatalf("get nil value: %v", err)
	}
	if value != nil {
		t.Fatalf("nil value: got %v, want nil", value)
	}
}

func TestEnvironmentReportsUndefinedValues(t *testing.T) {
	environment := NewEnvironment()
	name := NewToken(IDENTIFIER, "missing", "missing", 7)

	_, err := environment.Get(name)
	if err == nil {
		t.Fatal("get undefined value succeeded, want an error")
	}
	if !strings.Contains(err.Error(), "Undefined variable 'missing'.") {
		t.Fatalf("undefined value error: got %q", err)
	}
}
