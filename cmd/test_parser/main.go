package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"mycompiler/lexer"
	"mycompiler/parser"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	generator := RandomProgramGenerator{}
	randomCode := generator.Generate(10)

	fmt.Println("--- Generated Code ---")
	fmt.Println(randomCode)
	fmt.Println("----------------------")

	l := lexer.NewLexer(randomCode)
	tokens := l.Tokenize()

	p := parser.NewParser(tokens)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\n[Ошибка Парсера]:", r)
		}
	}()

	astTree := p.Parse()

	fmt.Printf("Успешно распарсено: %d инструкций на верхнем уровне.\n", len(astTree))

	printer := AstPrinter{}
	printer.Print(astTree)

	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
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
