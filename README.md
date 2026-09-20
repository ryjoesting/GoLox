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

The program parses each input expression and prints its parenthesized AST.
For example, `1 + 2 * 3` prints `(+ 1 (* 2 3))`. Syntax errors are written to
stderr, and invalid expressions are not printed as ASTs.

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

The expression parser tests can be run with:

```text
go test -run TestParser -v
```

## Current Implementation

The project currently includes the scanner and the expression AST from the Lox
grammar. The AST uses a sealed `Expr` interface with `Binary`, `Grouping`,
`Literal`, and `Unary` expression nodes. An `ExprVisitor` interface provides
double-dispatch for operations such as printing, and `AstPrinter` renders
expressions as parenthesized tree structures such as `(* (- 123) (group 45.67))`.

The expression parser uses recursive descent with precedence levels for comma,
conditional, equality, comparison, term, factor, unary, and primary
expressions. It constructs the AST nodes from scanned tokens and returns
structured syntax errors with line and token context. Statement parsing and
statement-level error recovery are not implemented yet.

## Project Files

- `golox.go` - Program entry point and command-line handling.
- `scanner.go` - Token definitions, lexical scanner, parser pipeline, REPL, and file execution.
- `ast.go` - Expression AST nodes and the `ExprVisitor` interface.
- `ast_printer.go` - Visitor-based AST printer for inspecting tree structure.
- `ast_test.go` - AST structure and printer tests.
- `parser.go` - Recursive-descent expression parser and structured parse errors.
- `parser_test.go` - Parser precedence, associativity, primary expression, and syntax error tests.
- `scanner_test.go` - Built-in Go tests for scanner behavior.
- `test_src_code.lox` - Example Lox source file.
- `go.mod` - Go module definition.
