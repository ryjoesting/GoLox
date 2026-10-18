package main

type LoxCallable interface {
	Arity() int
	Call(interpreter *Interpreter, arguments []any) any
}

type LoxFunction struct {
	declaration *FunctionStmt
	closure     *Environment
}

func NewLoxFunction(declaration *FunctionStmt, closure *Environment) *LoxFunction {
	return &LoxFunction{declaration: declaration, closure: closure}
}

func (function *LoxFunction) Arity() int {
	return len(function.declaration.Parameters)
}

func (function *LoxFunction) String() string {
	return "<fn " + function.declaration.Name.Lexeme + ">"
}

func (function *LoxFunction) Call(interpreter *Interpreter, arguments []any) any {
	environment := NewEnvironment(function.closure)
	for index, parameter := range function.declaration.Parameters {
		environment.Define(parameter.Lexeme, arguments[index])
	}

	var result any
	interpreter.functionDepth++
	defer func() {
		interpreter.functionDepth--
	}()

	func() {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			returnSignal, ok := recovered.(*returnSignal)
			if !ok {
				panic(recovered)
			}
			result = returnSignal.value
		}()
		interpreter.executeBlock(function.declaration.Body, environment)
	}()
	return result
}

type returnSignal struct {
	value any
}
