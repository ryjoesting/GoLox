package main

type Expr interface {
	isExpr()
	Accept(visitor ExprVisitor) any
}

type ExprVisitor interface {
	VisitBinaryExpr(expression *Binary) any
	VisitGroupingExpr(expression *Grouping) any
	VisitLiteralExpr(expression *Literal) any
	VisitUnaryExpr(expression *Unary) any
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

type Unary struct {
	Operator *Token
	Right    Expr
}

func (*Unary) isExpr() {}

func (expression *Unary) Accept(visitor ExprVisitor) any {
	return visitor.VisitUnaryExpr(expression)
}
