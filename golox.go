package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:] // Skip the first argument, which is the program name
	if len(args) > 1 {
		fmt.Println("Usage: golox [script]")
		os.Exit(64)
	} else if len(args) == 1 {
		os.Exit(runFile(args[0])) // Read in and run the source Lox file
	} else {
		runPrompt() // Start the REPL loop for Lox interpreter
	}
}
