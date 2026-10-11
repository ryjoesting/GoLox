package main

// Stmt stands for "Statement", its abbreviated since we will refer to it many times

type Stmt interface {
	isStmt()
	Accept(visitor StmtVisitor) any
}

type StmtVisitor interface {
	VisitExpressionStmt(statement *ExpressionStmt) any
	VisitPrintStmt(statement *PrintStmt) any
	VisitVarStmt(statement *VarStmt) any
}

type ExpressionStmt struct {
	Expression Expr
}

func (*ExpressionStmt) isStmt() {}

func (statement *ExpressionStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitExpressionStmt(statement)
}

type PrintStmt struct {
	Expression Expr
}

func (*PrintStmt) isStmt() {}

func (statement *PrintStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitPrintStmt(statement)
}

type VarStmt struct {
	Name        *Token
	Initializer Expr
}

func (*VarStmt) isStmt() {}

func (statement *VarStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitVarStmt(statement)
}
