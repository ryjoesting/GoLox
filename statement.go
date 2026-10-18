package main

// Stmt stands for "Statement", its abbreviated since we will refer to it many times

type Stmt interface {
	isStmt()
	Accept(visitor StmtVisitor) any
}

type StmtVisitor interface {
	VisitBlockStmt(statement *BlockStmt) any
	VisitExpressionStmt(statement *ExpressionStmt) any
	VisitFunctionStmt(statement *FunctionStmt) any
	VisitIfStmt(statement *IfStmt) any
	VisitPrintStmt(statement *PrintStmt) any
	VisitReturnStmt(statement *ReturnStmt) any
	VisitVarStmt(statement *VarStmt) any
	VisitWhileStmt(statement *WhileStmt) any
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

type FunctionStmt struct {
	Name       *Token
	Parameters []*Token
	Body       []Stmt
}

func (*FunctionStmt) isStmt() {}

func (statement *FunctionStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitFunctionStmt(statement)
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

type ReturnStmt struct {
	Keyword *Token
	Value   Expr
}

func (*ReturnStmt) isStmt() {}

func (statement *ReturnStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitReturnStmt(statement)
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

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (*WhileStmt) isStmt() {}

func (statement *WhileStmt) Accept(visitor StmtVisitor) any {
	return visitor.VisitWhileStmt(statement)
}
