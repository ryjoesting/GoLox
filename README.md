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
initializer contain `nil`. Global bindings are stored in an environment for
the lifetime of the interpreter, and redefining a global replaces its value.
Assignments such as `name = value;` update existing variables and evaluate to
the assigned value; assignment is right-associative and cannot create a new
variable. Braced blocks create lexical scopes, so variables declared inside a
block are local to it; nested scopes can read and assign enclosing variables,
and inner declarations shadow outer ones.
Function declarations use `fun name(parameters) { ... }`; functions are
first-class values, support recursion and closures, and return with `return
expression;` or `return;`. Calls evaluate arguments left to right, require an
exact parameter count, and support at most 255 arguments. Returning from
top-level code, calling a non-callable value, or passing the wrong number of
arguments reports a runtime error.
Conditional statements use `if (condition) statement` with an optional `else`
branch. Only the selected branch executes, and an `else` binds to the nearest
unmatched `if`.
Logical `or` and `and` operators short-circuit from left to right and return
operand values, with `or` having lower precedence than `and`.
While loops use `while (condition) statement`, re-evaluating the condition
before each iteration and accepting any statement as the body.
C-style `for` loops support optional initializer, condition, and increment
clauses. They are desugared into scoped blocks and `while` loops, so a
variable declared in `for (var i = ...; ...; ...)` is local to the loop.
Syntax and runtime errors are written to stderr, and the REPL continues after
an error.

When running a file, the process uses exit code `65` for file or syntax errors,
`70` for runtime errors, and `0` for success.

## Lox Language Guide

Lox is a small, dynamically typed language. Programs are made of statements,
and most statements end with a semicolon. Whitespace is insignificant, and
comments can be written with `//` or `/* ... */`.

### Values and expressions

Lox has numbers, strings, booleans, and `nil`:

```lox
var count = 3;
var message = "hello";
var ready = true;
var missing = nil;
```

Variables can be reassigned, and an assignment produces the assigned value:

```lox
count = count + 1;
print count;       // 4
print count = 10;  // 10
```

Arithmetic uses `+`, `-`, `*`, and `/`. The `+` operator adds two numbers or
concatenates two strings. Comparisons use `<`, `<=`, `>`, and `>=`; equality
uses `==` and `!=`. Parentheses group expressions, `!` negates truthiness, and
`-` negates a number:

```lox
print (2 + 3) * 4;       // 20
print "Lox" + " works";  // Lox works
print 3 < 5;             // true
print !false;            // true
```

`nil` and `false` are falsey. Every other value, including `0` and empty
strings, is truthy. The logical operators `and` and `or` short-circuit and
return one of their operands:

```lox
var name = nil;
print name or "Guest";  // Guest
print true and 42;      // 42
```

### Output and variables

Use `print` to display a value. A standalone expression is evaluated and its
result is discarded:

```lox
print "Welcome!";
2 + 2;  // evaluated, but not printed
```

`var` creates a variable. An initializer is optional; without one, the value
is `nil`. Variables declared in a block are local to that block:

```lox
var scope = "global";
{
  var scope = "local";
  print scope;  // local
}
print scope;    // global
```

### Conditions and loops

Use `if` with an optional `else` to choose which statement runs. Braces group
multiple statements into a block:

```lox
if (score >= 60) {
  print "pass";
} else {
  print "try again";
}
```

`while` repeats while its condition is truthy. A `for` loop has optional
initializer, condition, and increment clauses:

```lox
var i = 0;
while (i < 3) {
  print i;
  i = i + 1;
}

for (var j = 0; j < 3; j = j + 1) {
  print j;
}
```

### Functions

Declare a function with `fun`, list its parameters in parentheses, and put its
body in braces. Use `return` to produce a result; `return;` returns `nil`.
Functions can be stored in variables, passed as arguments, called recursively,
and can remember variables from their defining scope:

```lox
fun greet(name) {
  return "Hello, " + name + "!";
}

print greet("Ada");  // Hello, Ada!

fun factorial(n) {
  if (n <= 1) return 1;
  return n * factorial(n - 1);
}

print factorial(5);  // 120
```

Calls can have zero or more arguments, and the number of arguments must match
the number of parameters:

```lox
fun add(a, b) {
  return a + b;
}

print add(2, 3);  // 5
```

Functions support closures. An inner function can read and update variables
from an enclosing function even after the outer function returns:

```lox
fun makeCounter() {
  var value = 0;
  fun next() {
    value = value + 1;
    return value;
  }
  return next;
}

var counter = makeCounter();
print counter();  // 1
print counter();  // 2
```

### A complete small program

The following example combines variables, a function, a loop, and output:

```lox
fun sumTo(limit) {
  var total = 0;
  var number = 1;
  while (number <= limit) {
    total = total + number;
    number = number + 1;
  }
  return total;
}

print sumTo(5);  // 15
```

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
grammar. The AST uses a sealed `Expr` interface with `Binary`, `Call`,
`Grouping`, `Literal`, and `Unary` expression nodes. An `ExprVisitor` interface provides
double-dispatch for operations such as printing, and `AstPrinter` renders
expressions as parenthesized tree structures such as `(* (- 123) (group 45.67))`.

The parser uses recursive descent with precedence levels for comma, conditional,
equality, comparison, term, factor, unary, and primary expressions, then wraps
expressions in calls and declarations, expression statements, print statements,
or function declarations. The interpreter executes multiple statements in
order, evaluates literals, operators, logical operators, global variables,
assignments, lexical block scopes, conditionals, function calls, returns,
recursion, and closures through a chained environment abstraction, and reports
runtime errors with operator,
variable, and line context.
Statement-level synchronization recovers at semicolons and known statement
starters.

## Project Files

- `golox.go` - Program entry point and command-line handling.
- `scanner.go` - Token definitions, lexical scanner, parser pipeline, REPL, and file execution.
- `ast.go` - Expression AST nodes and the `ExprVisitor` interface.
- `stmt.go` - Statement AST nodes and the `StmtVisitor` interface.
- `ast_printer.go` - Visitor-based AST printer for inspecting tree structure.
- `ast_test.go` - AST structure and printer tests.
- `parser.go` - Recursive-descent program/statement parser and structured parse errors.
- `parser_test.go` - Parser precedence, associativity, primary expression, and syntax error tests.
- `environment.go` - Global name-to-value environment and undefined-variable lookup errors.
- `environment_test.go` - Environment definition, nesting, assignment, nil, and lookup-error tests.
- `interpreter.go` - Expression/statement evaluator, runtime errors, and value formatting.
- `function.go` - Callable interface and user-defined function closures.
- `interpreter_test.go` - Expression evaluation, runtime error, and pipeline tests.
- `stmt_test.go` - Statement parsing and execution tests.
- `scanner_test.go` - Built-in Go tests for scanner behavior.
- `test_src_code.lox` - Example Lox source file.
- `go.mod` - Go module definition.
