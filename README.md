# GoLox

GoLox is an implementation of the Lox language from *Crafting Interpreters*.

## Requirements

- Go 1.27.1 or later
- No third-party Go packages are required.

Install Go from the official instructions:

https://go.dev/doc/install

After installation, verify Go is available:

```text
go version
```

This project was built on a Windows 11, version 25H2 machine. Instructions may differ slightly for Mac or Linux but the core principle is the same. For questions, consult the Go documentation.

## Run the Project

Open PowerShell or Command Prompt (or Terminal) and change to the GoLox project directory:

```text
cd "C:\Users\{Path to Your Project}\golox"
```

Start the interactive prompt:

```text
go run .
```

At the `>` prompt, enter Lox source code. Press `Ctrl+C` to stop the program.

To scan a source file, provide its path as the first argument. The repository includes `test_src_code.lox`:

```text
go run . test_src_code.lox
```

You can also provide an absolute or relative path to another `.lox` source file:

```text
go run . path\to\your\source.lox
```

The program parses and executes complete Lox statements. Expression statements
must end with `;` and discard their result. Print statements display their
evaluated value, for example `print 1 + 2;` prints `3`. Global variables can
be declared with `var name = value;` and read by name; declarations without an
initializer contain `nil`. Syntax and runtime errors are written to stderr,
and the REPL continues after an error.

When running a file, the process uses exit code `65` for file or syntax errors,
`70` for runtime errors, and `0` for success.

## Run the Tests

Run all tests for this GoLox module from the project directory:

```text
go test
```

A successful test run ends with output similar to:

```text
PASS
ok      golox
```

To see each scanner test while it runs:

```text
go test -v
```

To run only scanner tests:

```text
go test -run TestScanner -v
```

The expression AST tests can be run with:

```text
go test -run 'TestExpression|TestAstPrinter' -v
```

The parser and statement tests can be run with:

```text
go test -run TestParser -v
```

Statement tests can also be run with:

```text
go test -run 'TestParserBuildsStatements|TestInterpreterExecutesStatements' -v
```

The interpreter tests can be run with:

```text
go test -run TestInterpreter -v
```

## Current Implementation

The project currently includes the scanner and the expression AST from the Lox
grammar. The AST uses a sealed `Expr` interface with `Binary`, `Grouping`,
`Literal`, and `Unary` expression nodes. An `ExprVisitor` interface provides
double-dispatch for operations such as printing, and `AstPrinter` renders
expressions as parenthesized tree structures such as `(* (- 123) (group 45.67))`.

The parser uses recursive descent with precedence levels for comma, conditional,
equality, comparison, term, factor, unary, and primary expressions, then wraps
expressions in declarations, expression statements, or print statements. The
interpreter executes multiple statements in order, evaluates literals,
operators, and global variables, and reports runtime errors with operator,
variable, and line context. Statement-level synchronization recovers at
semicolons and known statement starters.

## Project Files

- `golox.go` - Program entry point and command-line handling.
- `scanner.go` - Token definitions, lexical scanner, parser pipeline, REPL, and file execution.
- `ast.go` - Expression AST nodes and the `ExprVisitor` interface.
- `stmt.go` - Statement AST nodes and the `StmtVisitor` interface.
- `ast_printer.go` - Visitor-based AST printer for inspecting tree structure.
- `ast_test.go` - AST structure and printer tests.
- `parser.go` - Recursive-descent program/statement parser and structured parse errors.
- `parser_test.go` - Parser precedence, associativity, primary expression, and syntax error tests.
- `interpreter.go` - Expression/statement evaluator, runtime errors, and value formatting.
- `interpreter_test.go` - Expression evaluation, runtime error, and pipeline tests.
- `stmt_test.go` - Statement parsing and execution tests.
- `scanner_test.go` - Built-in Go tests for scanner behavior.
- `test_src_code.lox` - Example Lox source file.
- `go.mod` - Go module definition.
