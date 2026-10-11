package main

// Stmt stands for "Statement", its abbreviated since we will refer to it many times

type Stmt interface {
	isStmt()
	Accept(visitor StmtVisitor) any
}

type StmtVisitor interface {
	VisitBlockStmt(statement *BlockStmt) any
	VisitExpressionStmt(statement *ExpressionStmt) any
	VisitIfStmt(statement *IfStmt) any
	VisitPrintStmt(statement *PrintStmt) any
	VisitVarStmt(statement *VarStmt) any
}

type BlockStmt struct {
	Statements []Stmt
}

func (*BlockStmt) isStmt() {}

func (statement *BlockStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitBlockStmt(statement)
}

type ExpressionStmt struct {
	Expression Expr
}

func (*ExpressionStmt) isStmt() {}

func (statement *ExpressionStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitExpressionStmt(statement)
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (*IfStmt) isStmt() {}

func (statement *IfStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitIfStmt(statement)
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
