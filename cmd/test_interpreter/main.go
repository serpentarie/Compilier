package main

import (
	"fmt"
	"os"

	"mycompiler/interpreter"
	"mycompiler/lexer"
	"mycompiler/parser"
	"mycompiler/semantic"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./cmd/test_interpreter <path-to-source-file>")
		os.Exit(1)
	}

	sourcePath := os.Args[1]
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Println("failed to read source file:", err)
		os.Exit(1)
	}

	l := lexer.NewLexer(string(content))
	tokens := l.Tokenize()

	p := parser.NewParser(tokens)
	statements := p.Parse()
	if p.HasErrors() {
		fmt.Println("[Parser errors]")
		for _, parseErr := range p.Errors() {
			fmt.Println("-", parseErr)
		}
		os.Exit(1)
	}

	statements = semantic.Optimize(statements)

	i := interpreter.NewInterpreter()
	if err := i.Interpret(statements); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	warnings := i.Warnings()
	if len(warnings) > 0 {
		fmt.Println("[Runtime warnings]")
		for _, w := range warnings {
			fmt.Println("-", w)
		}
	}
}
