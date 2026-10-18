package main

type resolverFunctionType int

const (
	resolverNone resolverFunctionType = iota
	resolverFunction
	resolverInitializer
)

type resolverClassType int

const (
	resolverClassNone resolverClassType = iota
	resolverClass
	resolverSubclass
)

type Resolver struct {
	interpreter  *Interpreter
	scopes       []map[string]bool
	currentFunc  resolverFunctionType
	currentClass resolverClassType
	firstError   error
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{interpreter: interpreter}
}

func (resolver *Resolver) Resolve(statements []Stmt) error {
	resolver.firstError = nil
	for _, statement := range statements {
		resolver.resolveStmt(statement)
	}
	return resolver.firstError
}

func (resolver *Resolver) resolveStmt(statement Stmt) {
	if statement != nil {
		statement.Accept(resolver)
	}
}

func (resolver *Resolver) resolveExpr(expression Expr) {
	if expression != nil {
		expression.Accept(resolver)
	}
}

func (resolver *Resolver) VisitBlockStmt(statement *BlockStmt) any {
	resolver.beginScope()
	for _, nested := range statement.Statements {
		resolver.resolveStmt(nested)
	}
	resolver.endScope()
	return nil
}

func (resolver *Resolver) VisitExpressionStmt(statement *ExpressionStmt) any {
	resolver.resolveExpr(statement.Expression)
	return nil
}

func (resolver *Resolver) VisitFunctionStmt(statement *FunctionStmt) any {
	resolver.declare(statement.Name)
	resolver.define(statement.Name)
	resolver.resolveFunction(statement, resolverFunction)
	return nil
}

func (resolver *Resolver) VisitIfStmt(statement *IfStmt) any {
	resolver.resolveExpr(statement.Condition)
	resolver.resolveStmt(statement.ThenBranch)
	resolver.resolveStmt(statement.ElseBranch)
	return nil
}

func (resolver *Resolver) VisitPrintStmt(statement *PrintStmt) any {
	resolver.resolveExpr(statement.Expression)
	return nil
}

func (resolver *Resolver) VisitReturnStmt(statement *ReturnStmt) any {
	if resolver.currentFunc == resolverNone {
		resolver.report(statement.Keyword, "Can't return from top-level code.")
	} else if resolver.currentFunc == resolverInitializer && statement.Value != nil {
		resolver.report(statement.Keyword, "Can't return a value from an initializer.")
	}
	resolver.resolveExpr(statement.Value)
	return nil
}

func (resolver *Resolver) VisitVarStmt(statement *VarStmt) any {
	resolver.declare(statement.Name)
	if statement.Initializer != nil {
		resolver.resolveExpr(statement.Initializer)
	}
	resolver.define(statement.Name)
	return nil
}

func (resolver *Resolver) VisitWhileStmt(statement *WhileStmt) any {
	resolver.resolveExpr(statement.Condition)
	resolver.resolveStmt(statement.Body)
	return nil
}

func (resolver *Resolver) VisitAssignExpr(expression *Assign) any {
	resolver.resolveExpr(expression.Value)
	resolver.resolveLocal(expression, expression.Name)
	return nil
}

func (resolver *Resolver) VisitBinaryExpr(expression *Binary) any {
	resolver.resolveExpr(expression.Left)
	resolver.resolveExpr(expression.Right)
	return nil
}

func (resolver *Resolver) VisitCallExpr(expression *Call) any {
	resolver.resolveExpr(expression.Callee)
	for _, argument := range expression.Arguments {
		resolver.resolveExpr(argument)
	}
	return nil
}

func (resolver *Resolver) VisitConditionalExpr(expression *Conditional) any {
	resolver.resolveExpr(expression.Condition)
	resolver.resolveExpr(expression.Then)
	resolver.resolveExpr(expression.Else)
	return nil
}

func (resolver *Resolver) VisitGroupingExpr(expression *Grouping) any {
	resolver.resolveExpr(expression.Expression)
	return nil
}

func (resolver *Resolver) VisitLiteralExpr(expression *Literal) any {
	return nil
}

func (resolver *Resolver) VisitLogicalExpr(expression *Logical) any {
	resolver.resolveExpr(expression.Left)
	resolver.resolveExpr(expression.Right)
	return nil
}

func (resolver *Resolver) VisitUnaryExpr(expression *Unary) any {
	resolver.resolveExpr(expression.Right)
	return nil
}

func (resolver *Resolver) VisitVariableExpr(expression *Variable) any {
	if len(resolver.scopes) > 0 {
		defined, ok := resolver.scopes[len(resolver.scopes)-1][expression.Name.Lexeme]
		if ok && !defined {
			resolver.report(expression.Name, "Can't read local variable in its own initializer.")
		}
	}
	resolver.resolveLocal(expression, expression.Name)
	return nil
}

func (resolver *Resolver) resolveFunction(statement *FunctionStmt, functionType resolverFunctionType) {
	enclosing := resolver.currentFunc
	resolver.currentFunc = functionType
	resolver.beginScope()
	for _, parameter := range statement.Parameters {
		resolver.declare(parameter)
		resolver.define(parameter)
	}
	for _, bodyStatement := range statement.Body {
		resolver.resolveStmt(bodyStatement)
	}
	resolver.endScope()
	resolver.currentFunc = enclosing
}

func (resolver *Resolver) beginScope() {
	resolver.scopes = append(resolver.scopes, make(map[string]bool))
}

func (resolver *Resolver) endScope() {
	resolver.scopes = resolver.scopes[:len(resolver.scopes)-1]
}

func (resolver *Resolver) declare(name *Token) {
	if len(resolver.scopes) == 0 {
		return
	}
	scope := resolver.scopes[len(resolver.scopes)-1]
	if _, exists := scope[name.Lexeme]; exists {
		resolver.report(name, "Already a variable with this name in this scope.")
	}
	scope[name.Lexeme] = false
}

func (resolver *Resolver) define(name *Token) {
	if len(resolver.scopes) > 0 {
		resolver.scopes[len(resolver.scopes)-1][name.Lexeme] = true
	}
}

func (resolver *Resolver) resolveLocal(expression Expr, name *Token) {
	for index := len(resolver.scopes) - 1; index >= 0; index-- {
		if _, ok := resolver.scopes[index][name.Lexeme]; ok {
			resolver.interpreter.Resolve(expression, len(resolver.scopes)-1-index)
			return
		}
	}
	resolver.interpreter.ResolveGlobal(expression)
}

func (resolver *Resolver) report(token *Token, message string) {
	if resolver.firstError == nil {
		resolver.firstError = ParseError{Token: token, Message: message}
	}
}
