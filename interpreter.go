package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type Interpreter struct {
	runtimeError  error
	Output        io.Writer
	environment   *Environment
	globalEnv     *Environment
	functionDepth int
	locals        map[Expr]int
	globals       map[Expr]struct{}
}

var interpreter = NewInterpreter(nil)

func NewInterpreter(output io.Writer) *Interpreter {
	if output == nil {
		output = os.Stdout
	}
	globalEnv := NewEnvironment()
	return &Interpreter{
		Output: output, environment: globalEnv, globalEnv: globalEnv,
		locals: make(map[Expr]int), globals: make(map[Expr]struct{}),
	}
}

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

func (interpreter *Interpreter) Interpret(statements []Stmt) error {
	interpreter.runtimeError = nil
	for _, statement := range statements {
		interpreter.execute(statement)
		if interpreter.runtimeError != nil {
			return interpreter.runtimeError
		}
	}
	return nil
}

func (interpreter *Interpreter) execute(statement Stmt) {
	if interpreter.runtimeError == nil {
		statement.Accept(interpreter)
	}
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

func (interpreter *Interpreter) VisitLogicalExpr(expression *Logical) any {
	left := interpreter.evaluate(expression.Left)
	if interpreter.runtimeError != nil {
		return nil
	}

	if expression.Operator.Type == OR {
		if isTruthy(left) {
			return left
		}
	} else if !isTruthy(left) {
		return left
	}

	return interpreter.evaluate(expression.Right)
}

func (interpreter *Interpreter) VisitBlockStmt(statement *BlockStmt) any {
	interpreter.executeBlock(statement.Statements, NewEnvironment(interpreter.environment))
	return nil
}

func (interpreter *Interpreter) VisitIfStmt(statement *IfStmt) any {
	condition := interpreter.evaluate(statement.Condition)
	if interpreter.runtimeError != nil {
		return nil
	}
	if isTruthy(condition) {
		interpreter.execute(statement.ThenBranch)
	} else if statement.ElseBranch != nil {
		interpreter.execute(statement.ElseBranch)
	}
	return nil
}

func (interpreter *Interpreter) VisitWhileStmt(statement *WhileStmt) any {
	for {
		condition := interpreter.evaluate(statement.Condition)
		if interpreter.runtimeError != nil || !isTruthy(condition) {
			return nil
		}
		interpreter.execute(statement.Body)
		if interpreter.runtimeError != nil {
			return nil
		}
	}
}

func (interpreter *Interpreter) VisitAssignExpr(expression *Assign) any {
	value := interpreter.evaluate(expression.Value)
	if interpreter.runtimeError != nil {
		return nil
	}
	var err error
	if distance, ok := interpreter.locals[expression]; ok {
		err = interpreter.environment.AssignAt(distance, expression.Name, value)
	} else if _, global := interpreter.globals[expression]; global {
		err = interpreter.globalEnv.Assign(expression.Name, value)
	} else {
		err = interpreter.environment.Assign(expression.Name, value)
	}
	if err != nil {
		interpreter.runtimeError = err
		return nil
	}
	return value
}

func (interpreter *Interpreter) VisitVariableExpr(expression *Variable) any {
	var (
		value any
		err   error
	)
	if distance, ok := interpreter.locals[expression]; ok {
		value, err = interpreter.environment.GetAt(distance, expression.Name)
	} else if _, global := interpreter.globals[expression]; global {
		value, err = interpreter.globalEnv.Get(expression.Name)
	} else {
		value, err = interpreter.environment.Get(expression.Name)
	}
	if err != nil {
		interpreter.runtimeError = err
		return nil
	}
	return value
}

func (interpreter *Interpreter) Resolve(expression Expr, depth int) {
	interpreter.locals[expression] = depth
}

func (interpreter *Interpreter) ResolveGlobal(expression Expr) {
	interpreter.globals[expression] = struct{}{}
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

func (interpreter *Interpreter) VisitCallExpr(expression *Call) any {
	callee := interpreter.evaluate(expression.Callee)
	if interpreter.runtimeError != nil {
		return nil
	}

	arguments := make([]any, 0, len(expression.Arguments))
	for _, argument := range expression.Arguments {
		arguments = append(arguments, interpreter.evaluate(argument))
		if interpreter.runtimeError != nil {
			return nil
		}
	}

	function, ok := callee.(LoxCallable)
	if !ok {
		return interpreter.fail(expression.Paren, "Can only call functions and classes.")
	}
	if len(arguments) != function.Arity() {
		return interpreter.fail(expression.Paren, fmt.Sprintf(
			"Expected %d arguments but got %d.", function.Arity(), len(arguments),
		))
	}
	return function.Call(interpreter, arguments)
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

func (interpreter *Interpreter) VisitExpressionStmt(statement *ExpressionStmt) any {
	interpreter.evaluate(statement.Expression)
	return nil
}

func (interpreter *Interpreter) VisitFunctionStmt(statement *FunctionStmt) any {
	function := NewLoxFunction(statement, interpreter.environment)
	interpreter.environment.Define(statement.Name.Lexeme, function)
	return nil
}

func (interpreter *Interpreter) VisitPrintStmt(statement *PrintStmt) any {
	value := interpreter.evaluate(statement.Expression)
	if interpreter.runtimeError == nil {
		fmt.Fprintln(interpreter.Output, stringify(value))
	}
	return nil
}

func (interpreter *Interpreter) VisitReturnStmt(statement *ReturnStmt) any {
	if interpreter.functionDepth == 0 {
		return interpreter.fail(statement.Keyword, "Can't return from top-level code.")
	}

	var value any
	if statement.Value != nil {
		value = interpreter.evaluate(statement.Value)
		if interpreter.runtimeError != nil {
			return nil
		}
	}
	panic(&returnSignal{value: value})
}

func (interpreter *Interpreter) VisitVarStmt(statement *VarStmt) any {
	var value any
	if statement.Initializer != nil {
		value = interpreter.evaluate(statement.Initializer)
		if interpreter.runtimeError != nil {
			return nil
		}
	}
	interpreter.environment.Define(statement.Name.Lexeme, value)
	return nil
}

func (interpreter *Interpreter) executeBlock(statements []Stmt, environment *Environment) {
	previous := interpreter.environment
	interpreter.environment = environment
	defer func() {
		interpreter.environment = previous
	}()

	for _, statement := range statements {
		interpreter.execute(statement)
		if interpreter.runtimeError != nil {
			return
		}
	}
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
