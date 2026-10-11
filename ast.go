package main

type Expr interface {
	isExpr()
	Accept(visitor ExprVisitor) any
}

type ExprVisitor interface {
	VisitAssignExpr(expression *Assign) any
	VisitBinaryExpr(expression *Binary) any
	VisitConditionalExpr(expression *Conditional) any
	VisitGroupingExpr(expression *Grouping) any
	VisitLiteralExpr(expression *Literal) any
	VisitLogicalExpr(expression *Logical) any
	VisitUnaryExpr(expression *Unary) any
	VisitVariableExpr(expression *Variable) any
}

type Assign struct {
	Name  *Token
	Value Expr
}

func (*Assign) isExpr() {}

func (expression *Assign) Accept(visitor ExprVisitor) any {
	return visitor.VisitAssignExpr(expression)
}

type Binary struct {
	Left     Expr
	Operator *Token
	Right    Expr
}

func (*Binary) isExpr() {}

func (expression *Binary) Accept(visitor ExprVisitor) any {
	return visitor.VisitBinaryExpr(expression)
}

type Conditional struct {
	Condition Expr
	Then      Expr
	Else      Expr
}

func (*Conditional) isExpr() {}

func (expression *Conditional) Accept(visitor ExprVisitor) any {
	return visitor.VisitConditionalExpr(expression)
}

type Grouping struct {
	Expression Expr
}

func (*Grouping) isExpr() {}

func (expression *Grouping) Accept(visitor ExprVisitor) any {
	return visitor.VisitGroupingExpr(expression)
}

type Literal struct {
	Value any
}

func (*Literal) isExpr() {}

func (expression *Literal) Accept(visitor ExprVisitor) any {
	return visitor.VisitLiteralExpr(expression)
}

type Logical struct {
	Left     Expr
	Operator *Token
	Right    Expr
}

func (*Logical) isExpr() {}

func (expression *Logical) Accept(visitor ExprVisitor) any {
	return visitor.VisitLogicalExpr(expression)
}

type Unary struct {
	Operator *Token
	Right    Expr
}

func (*Unary) isExpr() {}

func (expression *Unary) Accept(visitor ExprVisitor) any {
	return visitor.VisitUnaryExpr(expression)
}

type Variable struct {
	Name *Token
}

func (*Variable) isExpr() {}

func (expression *Variable) Accept(visitor ExprVisitor) any {
	return visitor.VisitVariableExpr(expression)
}
