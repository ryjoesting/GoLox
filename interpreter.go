package main

import (
	"fmt"
	"strconv"
)

type Interpreter struct {
	runtimeError error
}

var interpreter = &Interpreter{}

type RuntimeError struct {
	Token   *Token
	Message string
}

func (error RuntimeError) Error() string {
	if error.Token == nil {
		return error.Message
	}
	return fmt.Sprintf("[line %d] RuntimeError at '%s': %s", error.Token.Line, error.Token.Lexeme, error.Message)
}

func (interpreter *Interpreter) Interpret(expression Expr) (any, error) {
	interpreter.runtimeError = nil
	value := interpreter.evaluate(expression)
	if interpreter.runtimeError != nil {
		return nil, interpreter.runtimeError
	}
	return value, nil
}

func (interpreter *Interpreter) evaluate(expression Expr) any {
	if interpreter.runtimeError != nil {
		return nil
	}
	return expression.Accept(interpreter)
}

func (interpreter *Interpreter) VisitLiteralExpr(expression *Literal) any {
	return expression.Value
}

func (interpreter *Interpreter) VisitGroupingExpr(expression *Grouping) any {
	return interpreter.evaluate(expression.Expression)
}

func (interpreter *Interpreter) VisitUnaryExpr(expression *Unary) any {
	right := interpreter.evaluate(expression.Right)
	if interpreter.runtimeError != nil {
		return nil
	}

	switch expression.Operator.Type {
	case BANG:
		return !isTruthy(right)
	case MINUS:
		number, ok := interpreter.checkNumberOperand(expression.Operator, right)
		if !ok {
			return nil
		}
		return -number
	default:
		return interpreter.fail(expression.Operator, "Unsupported unary operator.")
	}
}

func (interpreter *Interpreter) VisitBinaryExpr(expression *Binary) any {
	left := interpreter.evaluate(expression.Left)
	if interpreter.runtimeError != nil {
		return nil
	}
	right := interpreter.evaluate(expression.Right)
	if interpreter.runtimeError != nil {
		return nil
	}

	switch expression.Operator.Type {
	case COMMA:
		return right
	case PLUS:
		switch leftValue := left.(type) {
		case float64:
			rightValue, ok := right.(float64)
			if ok {
				return leftValue + rightValue
			}
		case string:
			rightValue, ok := right.(string)
			if ok {
				return leftValue + rightValue
			}
		}
		return interpreter.fail(expression.Operator, "Operands must be two numbers or two strings.")
	case MINUS, SLASH, STAR:
		leftValue, rightValue, ok := interpreter.checkNumberOperands(expression.Operator, left, right)
		if !ok {
			return nil
		}
		switch expression.Operator.Type {
		case MINUS:
			return leftValue - rightValue
		case SLASH:
			return leftValue / rightValue
		default:
			return leftValue * rightValue
		}
	case GREATER, GREATER_EQUAL, LESS, LESS_EQUAL:
		leftValue, rightValue, ok := interpreter.checkNumberOperands(expression.Operator, left, right)
		if !ok {
			return nil
		}
		switch expression.Operator.Type {
		case GREATER:
			return leftValue > rightValue
		case GREATER_EQUAL:
			return leftValue >= rightValue
		case LESS:
			return leftValue < rightValue
		default:
			return leftValue <= rightValue
		}
	case EQUAL_EQUAL:
		return isEqual(left, right)
	case BANG_EQUAL:
		return !isEqual(left, right)
	default:
		return interpreter.fail(expression.Operator, "Unsupported binary operator.")
	}
}

func (interpreter *Interpreter) VisitConditionalExpr(expression *Conditional) any {
	condition := interpreter.evaluate(expression.Condition)
	if interpreter.runtimeError != nil {
		return nil
	}
	if isTruthy(condition) {
		return interpreter.evaluate(expression.Then)
	}
	return interpreter.evaluate(expression.Else)
}

func (interpreter *Interpreter) fail(token *Token, message string) any {
	interpreter.runtimeError = RuntimeError{Token: token, Message: message}
	return nil
}

func (interpreter *Interpreter) checkNumberOperand(operator *Token, operand any) (float64, bool) {
	number, ok := operand.(float64)
	if !ok {
		interpreter.fail(operator, "Operand must be a number.")
	}
	return number, ok
}

func (interpreter *Interpreter) checkNumberOperands(operator *Token, left, right any) (float64, float64, bool) {
	leftValue, leftOK := left.(float64)
	rightValue, rightOK := right.(float64)
	if !leftOK || !rightOK {
		interpreter.fail(operator, "Operands must be numbers.")
	}
	return leftValue, rightValue, leftOK && rightOK
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}
	if boolean, ok := value.(bool); ok {
		return boolean
	}
	return true
}

func isEqual(left, right any) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return left == right
}

func stringify(value any) string {
	switch value := value.(type) {
	case nil:
		return "nil"
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	default:
		return fmt.Sprint(value)
	}
}
