// This file is the Scanner, covered in Chapter 4 of Crafting Interpreters
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var keywords = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func NewToken(tokenType TokenType, lexeme string, literal any, line int) *Token {
	return &Token{
		Type:    tokenType,
		Lexeme:  lexeme,
		Literal: literal,
		Line:    line,
	}
}

func (t *Token) String() string {
	return fmt.Sprintf("%s %s %v", t.Type.String(), t.Lexeme, t.Literal)
}

// TokenType represents the type of a lexical token.
type TokenType int

// Essentially, the Java Enum, represented in Go as a set of constants. The first constant is UNKNOWN, which is used to represent an uninitialized TokenType.
// The rest of the constants represent the various token types that can be encountered in the Lox programming language.
const (
	// 0 ensures an uninitialized TokenType is explicitly marked as unknown
	UNKNOWN TokenType = iota

	// Single-character tokens.
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	QUESTION
	COLON
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals.
	IDENTIFIER
	STRING
	NUMBER

	// Keywords.
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	EOF
)

// Helper function to convert a TokenType to its string representation. This is useful for debugging and logging purposes.
func (t TokenType) String() string {
	strings := [...]string{
		"UNKNOWN", "LEFT_PAREN", "RIGHT_PAREN", "LEFT_BRACE", "RIGHT_BRACE",
		"COMMA", "QUESTION", "COLON", "DOT", "MINUS", "PLUS", "SEMICOLON", "SLASH", "STAR",
		"BANG", "BANG_EQUAL", "EQUAL", "EQUAL_EQUAL", "GREATER", "GREATER_EQUAL",
		"LESS", "LESS_EQUAL", "IDENTIFIER", "STRING", "NUMBER",
		"AND", "CLASS", "ELSE", "FALSE", "FUN", "FOR", "IF", "NIL", "OR",
		"PRINT", "RETURN", "SUPER", "THIS", "TRUE", "VAR", "WHILE", "EOF",
	}

	if t < UNKNOWN || t > EOF {
		return "UNKNOWN"
	}
	return strings[t]
}

type Scanner struct {
	source  string
	tokens  []*Token
	start   int
	current int
	line    int
	err     error
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source:  source,
		tokens:  []*Token{},
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanTokens() []*Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}
	s.tokens = append(s.tokens, NewToken(EOF, "", nil, s.line))
	return s.tokens
}

func (s *Scanner) Error() error {
	return s.err
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(LEFT_PAREN, nil)
	case ')':
		s.addToken(RIGHT_PAREN, nil)
	case '{':
		s.addToken(LEFT_BRACE, nil)
	case '}':
		s.addToken(RIGHT_BRACE, nil)
	case ',':
		s.addToken(COMMA, nil)
	case '?':
		s.addToken(QUESTION, nil)
	case ':':
		s.addToken(COLON, nil)
	case '.':
		s.addToken(DOT, nil)
	case '-':
		s.addToken(MINUS, nil)
	case '+':
		s.addToken(PLUS, nil)
	case ';':
		s.addToken(SEMICOLON, nil)
	case '*':
		s.addToken(STAR, nil)
	// Behavior for these next cases depend on the next char ie != or >=
	case '/':
		if s.match('/') {
			// A comment goes until the end of the line.
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') {
			s.blockComment()
		} else {
			s.addToken(SLASH, nil)
		}

	case '!':
		if s.match('=') {
			s.addToken(BANG_EQUAL, nil)
		} else {
			s.addToken(BANG, nil)
		}
	case '=':
		if s.match('=') {
			s.addToken(EQUAL_EQUAL, nil)
		} else {
			s.addToken(EQUAL, nil)
		}
	case '<':
		if s.match('=') {
			s.addToken(LESS_EQUAL, nil)
		} else {
			s.addToken(LESS, nil)
		}
	case '>':
		if s.match('=') {
			s.addToken(GREATER_EQUAL, nil)
		} else {
			s.addToken(GREATER, nil)
		}
	// Ignores whitespaces and handles newlines
	case ' ', '\r', '\t':
		// Ignore whitespace.
	case '\n':
		s.line++
	// String literals
	case '"':
		s.string()
	// Number literals
	default:
		if isDigit(c) {
			s.number()
		} else if isAlpha(c) {
			s.identifier()
		} else {
			s.reportError(fmt.Sprintf("Unexpected character: %c.", c))
		}
	}
}

func (s *Scanner) match(expected byte) bool {
	// Checks if the next char matches the expected char. If it does, consumes it and returns true. Otherwise, returns false.
	if s.isAtEnd() {
		return false
	}
	if s.source[s.current] == expected {
		s.current++
		return true
	}
	return false
}

func (s *Scanner) peek() byte {
	// Returns the next char in the source without consuming it. If at the end of the source, returns 0.
	if s.isAtEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) advance() byte {
	// Consumes next char in source and returns it
	c := s.source[s.current]
	s.current++
	return c
}

func (s *Scanner) string() {
	// Handles string literals. Consumes characters until it finds a closing quote or reaches the end of the source.
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	if s.isAtEnd() {
		s.reportError("Unterminated string.")
		return
	}
	s.advance()
	s.addToken(STRING, s.source[s.start+1:s.current-1]) // trims the surrounding quotes using slice
}

func (s *Scanner) blockComment() {
	for !s.isAtEnd() {
		if s.peek() == '*' && s.peekNext() == '/' {
			s.advance()
			s.advance()
			return
		}

		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	s.reportError("Unterminated block comment.")
}

func isDigit(c byte) bool {
	// Checks if a character is a digit (0-9).
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	// Checks if a character is a letter (a-z or A-Z).
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func (s *Scanner) identifier() {
	// Handles identifiers and reserved keywords. Consumes characters until it finds a non-alphanumeric character.
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}
	// Check if the identifier is a reserved keyword.
	if tokenType, ok := keywords[string(s.source[s.start:s.current])]; ok {
		s.addToken(tokenType, nil)
	} else {
		s.addToken(IDENTIFIER, s.source[s.start:s.current])
	}
}

func (s *Scanner) number() {
	// Handles number literals. Consumes characters until it finds a non-digit character or reaches the end of the source.
	for isDigit(s.peek()) {
		s.advance()
	}
	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// Consume the "."
		s.advance()
		// Consume the digits after the "."
		for isDigit(s.peek()) {
			s.advance()
		}
	}
	s.addToken(NUMBER, s.source[s.start:s.current]) // trims the surrounding quotes using slice
}

func (s *Scanner) peekNext() byte {
	// Returns the character after the next character without consuming it. If at the end of the source, returns 0.
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func (s *Scanner) addToken(tokenType TokenType, literal any) {
	// Grabs the text of current lexeme and creates a new token for it
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, NewToken(tokenType, text, literal, s.line))
}

func runFile(path string) int {
	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		return 65
	}
	if err := runSource(string(content)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		switch err.(type) {
		case ParseError, ScanError:
			return 65
		}
		return 70
	}
	return 0
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSuffix(input, "\r")

		run(input)
	}
}

func run(src string) {
	if err := runSource(src); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func runSource(src string) error {
	statements, err := parseSource(src)
	if err != nil {
		return err
	}

	resolver := NewResolver(interpreter)
	if err := resolver.Resolve(statements); err != nil {
		return err
	}
	return interpreter.Interpret(statements)
}

func parseSource(src string) ([]Stmt, error) {
	scanner := NewScanner(src)
	tokens := scanner.scanTokens()
	if err := scanner.Error(); err != nil {
		return nil, err
	}
	return NewParser(tokens).Parse()
}

type ScanError struct {
	Line    int
	Message string
}

func (error ScanError) Error() string {
	return fmt.Sprintf("[line %d] Error: %s", error.Line, error.Message)
}

func (s *Scanner) reportError(message string) {
	if s.err == nil {
		s.err = ScanError{Line: s.line, Message: message}
	}
}
