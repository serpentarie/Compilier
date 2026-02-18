package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func main() {
	codeExample := "var x = 123; print x + 5;"

	lexer := NewLexer(codeExample)
	tokens := lexer.Tokenize()

	for _, token := range tokens {
		fmt.Println(token)
	}

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}

func GenerateRandomTestProgram() string {
	variables := []string{"a", "b", "c", "x", "y", "z"}
	operators := []string{"+", "-", "*", "/"}

	var program strings.Builder

	for i := 0; i < 5; i++ {
		varName := variables[rand.Intn(len(variables))]
		number := rand.Intn(100) + 1 // 1 to 100
		program.WriteString(fmt.Sprintf("var %s = %d;\n", varName, number))
	}

	for i := 0; i < 5; i++ {
		var1 := variables[rand.Intn(len(variables))]
		var2 := variables[rand.Intn(len(variables))]
		op := operators[rand.Intn(len(operators))]
		program.WriteString(fmt.Sprintf("print %s %s %s;\n", var1, op, var2))
	}

	return program.String()
}
