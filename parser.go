package main

import (
	"fmt"
	"strconv"
)

type Parser struct {
	tokens           []*Token
	current          int
	stopCommaNesting int
}

type ParseError struct {
	Token   *Token
	Message string
}

func (error ParseError) Error() string {
	if error.Token == nil {
		return error.Message
	}
	if error.Token.Type == EOF {
		return fmt.Sprintf("[line %d] Error at end: %s", error.Token.Line, error.Message)
	}
	return fmt.Sprintf("[line %d] Error at '%s': %s", error.Token.Line, error.Token.Lexeme, error.Message)
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens: tokens}
}

func (parser *Parser) Parse() ([]Stmt, error) {
	statements := []Stmt{}
	var firstError error

	for !parser.isAtEnd() {
		statement, err := parser.declaration()
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			parser.synchronize()
			continue
		}
		statements = append(statements, statement)
	}

	return statements, firstError
}

func (parser *Parser) declaration() (Stmt, error) {
	if parser.match(CLASS) {
		return parser.classDeclaration()
	}
	if parser.match(VAR) {
		return parser.varDeclaration()
	}
	if parser.match(FUN) {
		return parser.functionDeclaration()
	}
	return parser.statement()
}

func (parser *Parser) classDeclaration() (Stmt, error) {
	name, err := parser.consume(IDENTIFIER, "Expect class name.")
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(LEFT_BRACE, "Expect '{' before class body."); err != nil {
		return nil, err
	}

	methods := []*FunctionStmt{}
	for !parser.check(RIGHT_BRACE) && !parser.isAtEnd() {
		method, err := parser.functionDeclaration()
		if err != nil {
			return nil, err
		}
		methods = append(methods, method.(*FunctionStmt))
	}
	if _, err := parser.consume(RIGHT_BRACE, "Expect '}' after class body."); err != nil {
		return nil, err
	}
	return &ClassStmt{Name: name, Methods: methods}, nil
}

func (parser *Parser) functionDeclaration() (Stmt, error) {
	name, err := parser.consume(IDENTIFIER, "Expect function name.")
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(LEFT_PAREN, "Expect '(' after function name."); err != nil {
		return nil, err
	}

	parameters := []*Token{}
	var parameterLimitError error
	if !parser.check(RIGHT_PAREN) {
		for {
			parameter, err := parser.consume(IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}
			if len(parameters) >= 255 {
				if parameterLimitError == nil {
					parameterLimitError = parser.parseError(parameter, "Can't have more than 255 parameters.")
				}
			} else {
				parameters = append(parameters, parameter)
			}
			if !parser.match(COMMA) {
				break
			}
		}
	}
	if _, err := parser.consume(RIGHT_PAREN, "Expect ')' after parameters."); err != nil {
		return nil, err
	}
	if _, err := parser.consume(LEFT_BRACE, "Expect '{' before function body."); err != nil {
		return nil, err
	}
	body, err := parser.blockStatement()
	if err != nil {
		return nil, err
	}
	if parameterLimitError != nil {
		return nil, parameterLimitError
	}
	return &FunctionStmt{Name: name, Parameters: parameters, Body: body.(*BlockStmt).Statements}, nil
}

func (parser *Parser) statement() (Stmt, error) {
	if parser.match(LEFT_BRACE) {
		return parser.blockStatement()
	}
	if parser.match(IF) {
		return parser.ifStatement()
	}
	if parser.match(WHILE) {
		return parser.whileStatement()
	}
	if parser.match(FOR) {
		return parser.forStatement()
	}
	if parser.match(RETURN) {
		return parser.returnStatement()
	}
	if parser.match(PRINT) {
		return parser.printStatement()
	}
	return parser.expressionStatement()
}

func (parser *Parser) returnStatement() (Stmt, error) {
	keyword := parser.previous()
	var value Expr
	var err error
	if !parser.check(SEMICOLON) {
		value, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}
	if _, err := parser.consume(SEMICOLON, "Expect ';' after return value."); err != nil {
		return nil, err
	}
	return &ReturnStmt{Keyword: keyword, Value: value}, nil
}

func (parser *Parser) ifStatement() (Stmt, error) {
	if _, err := parser.consume(LEFT_PAREN, "Expect '(' after 'if'."); err != nil {
		return nil, err
	}
	condition, err := parser.expression()
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(RIGHT_PAREN, "Expect ')' after if condition."); err != nil {
		return nil, err
	}

	thenBranch, err := parser.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt
	if parser.match(ELSE) {
		elseBranch, err = parser.statement()
		if err != nil {
			return nil, err
		}
	}
	return &IfStmt{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

func (parser *Parser) whileStatement() (Stmt, error) {
	if _, err := parser.consume(LEFT_PAREN, "Expect '(' after 'while'."); err != nil {
		return nil, err
	}
	condition, err := parser.expression()
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(RIGHT_PAREN, "Expect ')' after condition."); err != nil {
		return nil, err
	}
	body, err := parser.statement()
	if err != nil {
		return nil, err
	}
	return &WhileStmt{Condition: condition, Body: body}, nil
}

func (parser *Parser) forStatement() (Stmt, error) {
	if _, err := parser.consume(LEFT_PAREN, "Expect '(' after 'for'."); err != nil {
		return nil, err
	}

	var initializer Stmt
	if parser.match(SEMICOLON) {
		initializer = nil
	} else if parser.match(VAR) {
		var err error
		initializer, err = parser.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		initializer, err = parser.expressionStatement()
		if err != nil {
			return nil, err
		}
	}

	var condition Expr
	if !parser.check(SEMICOLON) {
		var err error
		condition, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}
	if _, err := parser.consume(SEMICOLON, "Expect ';' after loop condition."); err != nil {
		return nil, err
	}

	var increment Expr
	if !parser.check(RIGHT_PAREN) {
		var err error
		increment, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}
	if _, err := parser.consume(RIGHT_PAREN, "Expect ')' after for clauses."); err != nil {
		return nil, err
	}

	body, err := parser.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = &BlockStmt{Statements: []Stmt{
			body,
			&ExpressionStmt{Expression: increment},
		}}
	}
	if condition == nil {
		condition = &Literal{Value: true}
	}
	body = &WhileStmt{Condition: condition, Body: body}
	if initializer != nil {
		body = &BlockStmt{Statements: []Stmt{initializer, body}}
	}

	return body, nil
}

func (parser *Parser) blockStatement() (Stmt, error) {
	statements := []Stmt{}
	var firstError error

	for !parser.check(RIGHT_BRACE) && !parser.isAtEnd() {
		statement, err := parser.declaration()
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			parser.synchronize()
			continue
		}
		statements = append(statements, statement)
	}

	if _, err := parser.consume(RIGHT_BRACE, "Expect '}' after block."); err != nil {
		if firstError != nil {
			return nil, firstError
		}
		return nil, err
	}
	if firstError != nil {
		return nil, firstError
	}
	return &BlockStmt{Statements: statements}, nil
}

func (parser *Parser) printStatement() (Stmt, error) {
	expression, err := parser.expression()
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(SEMICOLON, "Expect ';' after value."); err != nil {
		return nil, err
	}
	return &PrintStmt{Expression: expression}, nil
}

func (parser *Parser) varDeclaration() (Stmt, error) {
	name, err := parser.consume(IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer Expr
	if parser.match(EQUAL) {
		initializer, err = parser.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := parser.consume(SEMICOLON, "Expect ';' after variable declaration."); err != nil {
		return nil, err
	}
	return &VarStmt{Name: name, Initializer: initializer}, nil
}

func (parser *Parser) expressionStatement() (Stmt, error) {
	expression, err := parser.expression()
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(SEMICOLON, "Expect ';' after expression."); err != nil {
		return nil, err
	}
	return &ExpressionStmt{Expression: expression}, nil
}

func (parser *Parser) synchronize() {
	for !parser.isAtEnd() {
		if parser.previous().Type == SEMICOLON {
			return
		}
		switch parser.peek().Type {
		case CLASS, FUN, VAR, FOR, IF, WHILE, PRINT, RETURN, RIGHT_BRACE:
			return
		}
		parser.advance()
	}
}

func (parser *Parser) expression() (Expr, error) {
	return parser.assignment()
}

func (parser *Parser) assignment() (Expr, error) {
	expression, err := parser.logicOr()
	if err != nil {
		return nil, err
	}

	if !parser.match(EQUAL) {
		return expression, nil
	}

	equals := parser.previous()
	value, err := parser.assignment()
	if err != nil {
		return nil, err
	}

	switch target := expression.(type) {
	case *Variable:
		return &Assign{Name: target.Name, Value: value}, nil
	case *Get:
		return &Set{Object: target.Object, Name: target.Name, Value: value}, nil
	default:
		return nil, parser.parseError(equals, "Invalid assignment target.")
	}
}

func (parser *Parser) logicOr() (Expr, error) {
	expression, err := parser.logicAnd()
	if err != nil {
		return nil, err
	}

	for parser.match(OR) {
		operator := parser.previous()
		right, err := parser.logicAnd()
		if err != nil {
			return nil, err
		}
		expression = &Logical{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) logicAnd() (Expr, error) {
	expression, err := parser.comma()
	if err != nil {
		return nil, err
	}

	for parser.match(AND) {
		operator := parser.previous()
		right, err := parser.comma()
		if err != nil {
			return nil, err
		}
		expression = &Logical{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) comma() (Expr, error) {
	if parser.stopCommaNesting > 0 {
		return parser.conditional()
	}
	if parser.match(COMMA) {
		operator := parser.previous()
		if _, err := parser.conditional(); err != nil {
			return nil, err
		}
		return nil, parser.parseError(operator, "Expect expression before binary operator.")
	}

	expression, err := parser.conditional()
	if err != nil {
		return nil, err
	}

	for parser.match(COMMA) {
		operator := parser.previous()
		right, err := parser.conditional()
		if err != nil {
			return nil, err
		}
		expression = &Binary{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) conditional() (Expr, error) {
	condition, err := parser.equality()
	if err != nil {
		return nil, err
	}

	if !parser.match(QUESTION) {
		return condition, nil
	}

	thenBranch, err := parser.expression()
	if err != nil {
		return nil, err
	}
	if _, err := parser.consume(COLON, "Expect ':' after expression."); err != nil {
		return nil, err
	}
	elseBranch, err := parser.conditional()
	if err != nil {
		return nil, err
	}
	return &Conditional{Condition: condition, Then: thenBranch, Else: elseBranch}, nil
}

func (parser *Parser) equality() (Expr, error) {
	if parser.match(BANG_EQUAL, EQUAL_EQUAL) {
		operator := parser.previous()
		if _, err := parser.comparison(); err != nil {
			return nil, err
		}
		return nil, parser.parseError(operator, "Expect expression before binary operator.")
	}

	expression, err := parser.comparison()
	if err != nil {
		return nil, err
	}

	for parser.match(BANG_EQUAL, EQUAL_EQUAL) {
		operator := parser.previous()
		right, err := parser.comparison()
		if err != nil {
			return nil, err
		}
		expression = &Binary{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) comparison() (Expr, error) {
	if parser.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := parser.previous()
		if _, err := parser.term(); err != nil {
			return nil, err
		}
		return nil, parser.parseError(operator, "Expect expression before binary operator.")
	}

	expression, err := parser.term()
	if err != nil {
		return nil, err
	}

	for parser.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := parser.previous()
		right, err := parser.term()
		if err != nil {
			return nil, err
		}
		expression = &Binary{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) term() (Expr, error) {
	if parser.match(PLUS) {
		operator := parser.previous()
		if _, err := parser.factor(); err != nil {
			return nil, err
		}
		return nil, parser.parseError(operator, "Expect expression before binary operator.")
	}

	expression, err := parser.factor()
	if err != nil {
		return nil, err
	}

	for parser.match(MINUS, PLUS) {
		operator := parser.previous()
		right, err := parser.factor()
		if err != nil {
			return nil, err
		}
		expression = &Binary{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) factor() (Expr, error) {
	if parser.match(SLASH, STAR) {
		operator := parser.previous()
		if _, err := parser.unary(); err != nil {
			return nil, err
		}
		return nil, parser.parseError(operator, "Expect expression before binary operator.")
	}

	expression, err := parser.unary()
	if err != nil {
		return nil, err
	}

	for parser.match(SLASH, STAR) {
		operator := parser.previous()
		right, err := parser.unary()
		if err != nil {
			return nil, err
		}
		expression = &Binary{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (parser *Parser) unary() (Expr, error) {
	if parser.match(BANG, MINUS) {
		operator := parser.previous()
		right, err := parser.unary()
		if err != nil {
			return nil, err
		}
		return &Unary{Operator: operator, Right: right}, nil
	}
	return parser.call()
}

func (parser *Parser) call() (Expr, error) {
	expression, err := parser.primary()
	if err != nil {
		return nil, err
	}

	for {
		if parser.match(LEFT_PAREN) {
			expression, err = parser.finishCall(expression)
			if err != nil {
				return nil, err
			}
		} else if parser.match(DOT) {
			name, err := parser.consume(IDENTIFIER, "Expect property name after '.'.")
			if err != nil {
				return nil, err
			}
			expression = &Get{Object: expression, Name: name}
		} else {
			break
		}
	}
	return expression, nil
}

func (parser *Parser) finishCall(callee Expr) (Expr, error) {
	arguments := []Expr{}
	if !parser.check(RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				return nil, parser.parseError(parser.peek(), "Can't have more than 255 arguments.")
			}
			parser.stopCommaNesting++
			argument, err := parser.assignment()
			parser.stopCommaNesting--
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, argument)
			if !parser.match(COMMA) {
				break
			}
		}
	}
	paren, err := parser.consume(RIGHT_PAREN, "Expect ')' after arguments.")
	if err != nil {
		return nil, err
	}
	return &Call{Callee: callee, Paren: paren, Arguments: arguments}, nil
}

func (parser *Parser) primary() (Expr, error) {
	if parser.match(FALSE) {
		return &Literal{Value: false}, nil
	}
	if parser.match(TRUE) {
		return &Literal{Value: true}, nil
	}
	if parser.match(NIL) {
		return &Literal{Value: nil}, nil
	}
	if parser.match(NUMBER) {
		number, err := strconv.ParseFloat(parser.previous().Literal.(string), 64)
		if err != nil {
			return nil, parser.parseError(parser.previous(), "Invalid number literal.")
		}
		return &Literal{Value: number}, nil
	}
	if parser.match(STRING) {
		return &Literal{Value: parser.previous().Literal}, nil
	}
	if parser.match(THIS) {
		return &This{Keyword: parser.previous()}, nil
	}
	if parser.match(LEFT_PAREN) {
		expression, err := parser.expression()
		if err != nil {
			return nil, err
		}
		if _, err := parser.consume(RIGHT_PAREN, "Expect ')' after expression."); err != nil {
			return nil, err
		}
		return &Grouping{Expression: expression}, nil
	}
	if parser.match(IDENTIFIER) {
		return &Variable{Name: parser.previous()}, nil
	}
	return nil, parser.parseError(parser.peek(), "Expect expression.")
}

func (parser *Parser) match(types ...TokenType) bool {
	for _, tokenType := range types {
		if parser.check(tokenType) {
			parser.advance()
			return true
		}
	}
	return false
}

func (parser *Parser) check(tokenType TokenType) bool {
	if parser.isAtEnd() {
		return tokenType == EOF
	}
	return parser.peek().Type == tokenType
}

func (parser *Parser) advance() *Token {
	if !parser.isAtEnd() {
		parser.current++
	}
	return parser.previous()
}

func (parser *Parser) isAtEnd() bool {
	return parser.current >= len(parser.tokens) || parser.peek().Type == EOF
}

func (parser *Parser) peek() *Token {
	if len(parser.tokens) == 0 {
		return &Token{Type: EOF}
	}
	if parser.current >= len(parser.tokens) {
		return parser.tokens[len(parser.tokens)-1]
	}
	return parser.tokens[parser.current]
}

func (parser *Parser) previous() *Token {
	if parser.current == 0 {
		return parser.peek()
	}
	return parser.tokens[parser.current-1]
}

func (parser *Parser) consume(tokenType TokenType, message string) (*Token, error) {
	if parser.check(tokenType) {
		return parser.advance(), nil
	}
	return nil, parser.parseError(parser.peek(), message)
}

func (parser *Parser) parseError(token *Token, message string) error {
	return ParseError{Token: token, Message: message}
}
