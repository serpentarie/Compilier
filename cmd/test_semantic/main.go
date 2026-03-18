package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"mycompiler/lexer"
	"mycompiler/parser"
	"mycompiler/semantic"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	source := ""
	if len(os.Args) >= 2 {
		b, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("failed to read file:", err)
			os.Exit(1)
		}
		source = string(b)
	} else {
		generator := RandomProgramGenerator{}
		source = generator.Generate(12)
	}

	fmt.Println("--- Source ---")
	fmt.Println(source)
	fmt.Println("------------")

	l := lexer.NewLexer(source)
	tokens := l.Tokenize()

	p := parser.NewParser(tokens)
	statements := p.Parse()

	if p.HasErrors() {
		fmt.Println("[Parser errors]")
		for _, err := range p.Errors() {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	a := semantic.NewAnalyzer()
	errs := a.Analyze(statements)
	if len(errs) == 0 {
		fmt.Println("OK: semantic analysis passed")
		return
	}

	fmt.Println("[Semantic errors]")
	for _, err := range errs {
		fmt.Println("-", err)
	}
	os.Exit(2)
}

type RandomProgramGenerator struct{}

func (g *RandomProgramGenerator) Generate(lines int) string {
	variables := []string{"a", "b", "c", "x", "y", "z"}
	operators := []string{"+", "-", "*", "/"}

	var sb strings.Builder

	for i := 0; i < lines; i++ {
		actionType := rand.Intn(4)

		switch actionType {
		case 0: // var x = 5 + y;
			v := variables[rand.Intn(len(variables))]
			val := rand.Intn(100)
			sb.WriteString(fmt.Sprintf("var %s = %d;\n", v, val))
		case 1: // print x + 5;
			v := variables[rand.Intn(len(variables))]
			op := operators[rand.Intn(len(operators))]
			val := rand.Intn(50)
			sb.WriteString(fmt.Sprintf("print %s %s %d;\n", v, op, val))
		case 2: // if (x < 50) { print x; }
			v := variables[rand.Intn(len(variables))]
			sb.WriteString(fmt.Sprintf("if (%s < 50) { print %s; }\n", v, v))
		case 3: // x = y + 10;
			v1 := variables[rand.Intn(len(variables))]
			v2 := variables[rand.Intn(len(variables))]
			sb.WriteString(fmt.Sprintf("%s = %s + 10;\n", v1, v2))
		}
	}
	return sb.String()
}
