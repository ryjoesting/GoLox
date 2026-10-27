package main

import (
	"reflect"
	"testing"
)

type expectedToken struct {
	tokenType TokenType
	lexeme    string
	literal   any
	line      int
}

func token(tokenType TokenType, lexeme string, literal any, line int) expectedToken {
	return expectedToken{
		tokenType: tokenType,
		lexeme:    lexeme,
		literal:   literal,
		line:      line,
	}
}

func assertTokens(t *testing.T, source string, expected []expectedToken) {
	t.Helper()

	actual := NewScanner(source).scanTokens()
	if len(actual) != len(expected) {
		t.Fatalf("source %q: got %d tokens, want %d: %#v", source, len(actual), len(expected), actual)
	}

	for index, want := range expected {
		got := actual[index]
		if got.Type != want.tokenType {
			t.Errorf("token %d type: got %s, want %s", index, got.Type, want.tokenType)
		}
		if got.Lexeme != want.lexeme {
			t.Errorf("token %d lexeme: got %q, want %q", index, got.Lexeme, want.lexeme)
		}
		if !reflect.DeepEqual(got.Literal, want.literal) {
			t.Errorf("token %d literal: got %#v, want %#v", index, got.Literal, want.literal)
		}
		if got.Line != want.line {
			t.Errorf("token %d line: got %d, want %d", index, got.Line, want.line)
		}
	}
}

func eof(line int) expectedToken {
	return token(EOF, "", nil, line)
}

func TestScannerSingleCharacterTokens(t *testing.T) {
	assertTokens(t, "() {},.-+;/ *", []expectedToken{
		token(LEFT_PAREN, "(", nil, 1),
		token(RIGHT_PAREN, ")", nil, 1),
		token(LEFT_BRACE, "{", nil, 1),
		token(RIGHT_BRACE, "}", nil, 1),
		token(COMMA, ",", nil, 1),
		token(DOT, ".", nil, 1),
		token(MINUS, "-", nil, 1),
		token(PLUS, "+", nil, 1),
		token(SEMICOLON, ";", nil, 1),
		token(SLASH, "/", nil, 1),
		token(STAR, "*", nil, 1),
		eof(1),
	})
}

func TestScannerCompoundOperators(t *testing.T) {
	assertTokens(t, "! != = == < <= > >=", []expectedToken{
		token(BANG, "!", nil, 1),
		token(BANG_EQUAL, "!=", nil, 1),
		token(EQUAL, "=", nil, 1),
		token(EQUAL_EQUAL, "==", nil, 1),
		token(LESS, "<", nil, 1),
		token(LESS_EQUAL, "<=", nil, 1),
		token(GREATER, ">", nil, 1),
		token(GREATER_EQUAL, ">=", nil, 1),
		eof(1),
	})
}

func TestScannerConditionalTokens(t *testing.T) {
	assertTokens(t, "? :", []expectedToken{
		token(QUESTION, "?", nil, 1),
		token(COLON, ":", nil, 1),
		eof(1),
	})
}

func TestScannerWhitespaceCommentsAndLines(t *testing.T) {
	assertTokens(t, "  // ignored\nprint\ttrue\r\n// ignored at EOF", []expectedToken{
		token(PRINT, "print", nil, 2),
		token(TRUE, "true", nil, 2),
		eof(3),
	})
}

func TestScannerBlockComments(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []expectedToken
	}{
		{
			name:   "same line",
			source: "print /* ignored */ true;",
			want: []expectedToken{
				token(PRINT, "print", nil, 1),
				token(TRUE, "true", nil, 1),
				token(SEMICOLON, ";", nil, 1),
				eof(1),
			},
		},
		{
			name:   "multiline",
			source: "var /* first line\nsecond line */ answer;",
			want: []expectedToken{
				token(VAR, "var", nil, 1),
				token(IDENTIFIER, "answer", "answer", 2),
				token(SEMICOLON, ";", nil, 2),
				eof(2),
			},
		},
		{
			name:   "empty",
			source: "/**print true;*/",
			want:   []expectedToken{eof(1)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertTokens(t, test.source, test.want)
		})
	}
}

func TestScannerUnterminatedBlockComment(t *testing.T) {
	assertTokens(t, "print /* ignored\nstill ignored", []expectedToken{
		token(PRINT, "print", nil, 1),
		eof(2),
	})
}

func TestScannerKeywords(t *testing.T) {
	keywordsSource := "and class else false for fun if nil or print return super this true var while"
	assertTokens(t, keywordsSource, []expectedToken{
		token(AND, "and", nil, 1),
		token(CLASS, "class", nil, 1),
		token(ELSE, "else", nil, 1),
		token(FALSE, "false", nil, 1),
		token(FOR, "for", nil, 1),
		token(FUN, "fun", nil, 1),
		token(IF, "if", nil, 1),
		token(NIL, "nil", nil, 1),
		token(OR, "or", nil, 1),
		token(PRINT, "print", nil, 1),
		token(RETURN, "return", nil, 1),
		token(SUPER, "super", nil, 1),
		token(THIS, "this", nil, 1),
		token(TRUE, "true", nil, 1),
		token(VAR, "var", nil, 1),
		token(WHILE, "while", nil, 1),
		eof(1),
	})
}

func TestScannerIdentifiersAndKeywordPrefixes(t *testing.T) {
	assertTokens(t, "_name name2 andrew className trueValue", []expectedToken{
		token(IDENTIFIER, "_name", "_name", 1),
		token(IDENTIFIER, "name2", "name2", 1),
		token(IDENTIFIER, "andrew", "andrew", 1),
		token(IDENTIFIER, "className", "className", 1),
		token(IDENTIFIER, "trueValue", "trueValue", 1),
		eof(1),
	})
}

func TestScannerStrings(t *testing.T) {
	assertTokens(t, `"" "hello, lox" "line one
line two"`, []expectedToken{
		token(STRING, `""`, "", 1),
		token(STRING, `"hello, lox"`, "hello, lox", 1),
		token(STRING, "\"line one\nline two\"", "line one\nline two", 2),
		eof(2),
	})
}

func TestScannerNumbers(t *testing.T) {
	assertTokens(t, "123 123.45 123. 1.2.3", []expectedToken{
		token(NUMBER, "123", "123", 1),
		token(NUMBER, "123.45", "123.45", 1),
		token(NUMBER, "123", "123", 1),
		token(DOT, ".", nil, 1),
		token(NUMBER, "1.2", "1.2", 1),
		token(DOT, ".", nil, 1),
		token(NUMBER, "3", "3", 1),
		eof(1),
	})
}

func TestScannerNumberFractionRequiresDigits(t *testing.T) {
	assertTokens(t, "42.foo 7.0", []expectedToken{
		token(NUMBER, "42", "42", 1),
		token(DOT, ".", nil, 1),
		token(IDENTIFIER, "foo", "foo", 1),
		token(NUMBER, "7.0", "7.0", 1),
		eof(1),
	})
}

func TestScannerUnterminatedStringTracksLines(t *testing.T) {
	assertTokens(t, "\"first line\nsecond line", []expectedToken{
		eof(2),
	})
}

func TestScannerEmptySource(t *testing.T) {
	assertTokens(t, "", []expectedToken{eof(1)})
}

func TestScannerReportsInvalidSource(t *testing.T) {
	tests := []string{"@", `"unterminated`, "/* unterminated"}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			scanner := NewScanner(source)
			scanner.scanTokens()
			if scanner.Error() == nil {
				t.Fatalf("source %q: scanner error is nil", source)
			}
		})
	}
}

func TestScannerMixedSource(t *testing.T) {
	assertTokens(t, "var answer = 42; // comment\nprint answer;", []expectedToken{
		token(VAR, "var", nil, 1),
		token(IDENTIFIER, "answer", "answer", 1),
		token(EQUAL, "=", nil, 1),
		token(NUMBER, "42", "42", 1),
		token(SEMICOLON, ";", nil, 1),
		token(PRINT, "print", nil, 2),
		token(IDENTIFIER, "answer", "answer", 2),
		token(SEMICOLON, ";", nil, 2),
		eof(2),
	})
}
