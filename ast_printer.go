package main

import (
	"fmt"
	"strings"
)

// This file is a helper module which implements the Visitor pattern for the Expr interface.
// It provides a way to traverse and print the abstract syntax tree (AST) of expressions in a human-readable format.
// This will help in debugging and understanding the structure of the parsed expressions

type AstPrinter struct{}

func (printer AstPrinter) Print(expression Expr) string {
	return expression.Accept(printer).(string)
}

func (printer AstPrinter) VisitAssignExpr(expression *Assign) any {
	return printer.parenthesize("=", &Variable{Name: expression.Name}, expression.Value)
}

func (printer AstPrinter) VisitBinaryExpr(expression *Binary) any {
	return printer.parenthesize(expression.Operator.Lexeme, expression.Left, expression.Right)
}

func (printer AstPrinter) VisitConditionalExpr(expression *Conditional) any {
	return printer.parenthesize("?:", expression.Condition, expression.Then, expression.Else)
}

func (printer AstPrinter) VisitGroupingExpr(expression *Grouping) any {
	return printer.parenthesize("group", expression.Expression)
}

func (AstPrinter) VisitLiteralExpr(expression *Literal) any {
	if expression.Value == nil {
		return "nil"
	}
	return fmt.Sprint(expression.Value)
}

func (printer AstPrinter) VisitLogicalExpr(expression *Logical) any {
	return printer.parenthesize(expression.Operator.Lexeme, expression.Left, expression.Right)
}

func (AstPrinter) VisitVariableExpr(expression *Variable) any {
	return expression.Name.Lexeme
}

func (printer AstPrinter) VisitUnaryExpr(expression *Unary) any {
	return printer.parenthesize(expression.Operator.Lexeme, expression.Right)
}

func (printer AstPrinter) parenthesize(name string, expressions ...Expr) string {
	var output strings.Builder
	output.WriteByte('(')
	output.WriteString(name)

	for _, expression := range expressions {
		output.WriteByte(' ')
		output.WriteString(printer.Print(expression))
	}

	output.WriteByte(')')
	return output.String()
}
