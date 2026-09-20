package main

import (
	"fmt"
	"strconv"
)

type Parser struct {
	tokens  []*Token
	current int
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
		statement, err := parser.statement()
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

func (parser *Parser) statement() (Stmt, error) {
	if parser.match(PRINT) {
		return parser.printStatement()
	}
	return parser.expressionStatement()
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
		case CLASS, FUN, VAR, FOR, IF, WHILE, PRINT, RETURN:
			return
		}
		parser.advance()
	}
}

func (parser *Parser) expression() (Expr, error) {
	return parser.comma()
}

func (parser *Parser) comma() (Expr, error) {
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
	return parser.primary()
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
