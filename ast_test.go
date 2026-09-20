package main

import "testing"

func TestExpressionNodesImplementExpr(t *testing.T) {
	var expressions = []Expr{
		&Binary{},
		&Grouping{},
		&Literal{},
		&Unary{},
	}

	if len(expressions) != 4 {
		t.Fatalf("got %d expression nodes, want 4", len(expressions))
	}
}

func TestExpressionNodesPreserveFields(t *testing.T) {
	operator := NewToken(PLUS, "+", nil, 1)
	left := &Literal{Value: "1"}
	right := &Literal{Value: "2"}
	binary := &Binary{Left: left, Operator: operator, Right: right}
	grouping := &Grouping{Expression: binary}
	unary := &Unary{Operator: NewToken(MINUS, "-", nil, 1), Right: grouping}

	if binary.Left != left || binary.Operator != operator || binary.Right != right {
		t.Error("binary expression did not preserve its fields")
	}
	if grouping.Expression != binary {
		t.Error("grouping expression did not preserve its child")
	}
	if unary.Right != grouping {
		t.Error("unary expression did not preserve its operand")
	}
	if unary.Operator.Type != MINUS {
		t.Errorf("unary operator type: got %s, want %s", unary.Operator.Type, MINUS)
	}
	if left.Value != "1" || right.Value != "2" {
		t.Error("literal expressions did not preserve their values")
	}
}

func TestAstPrinterPrintsNestedExpression(t *testing.T) {
	expression := &Binary{
		Left: &Unary{
			Operator: NewToken(MINUS, "-", nil, 1),
			Right:    &Literal{Value: "123"},
		},
		Operator: NewToken(STAR, "*", nil, 1),
		Right: &Grouping{
			Expression: &Literal{Value: "45.67"},
		},
	}

	got := (AstPrinter{}).Print(expression)
	want := "(* (- 123) (group 45.67))"
	if got != want {
		t.Fatalf("printed expression: got %q, want %q", got, want)
	}
}

func TestAstPrinterPrintsLiteralValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: "nil"},
		{name: "string", value: "hello", want: "hello"},
		{name: "number", value: 123.45, want: "123.45"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := (AstPrinter{}).Print(&Literal{Value: test.value})
			if got != test.want {
				t.Fatalf("printed literal: got %q, want %q", got, test.want)
			}
		})
	}
}
