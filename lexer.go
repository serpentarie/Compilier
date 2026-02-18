package main

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	input    string
	length   int
	position int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    input,
		length:   len(input),
		position: 0,
	}
}

func (l *Lexer) Tokenize() []Token {
	var result []Token

	for l.position < l.length {
		current := l.Peek()

		if unicode.IsSpace(current) {
			l.Next()
			continue
		}

		if unicode.IsDigit(current) {
			l.TokenizeNumber(&result)
			continue
		}

		if unicode.IsLetter(current) {
			l.TokenizeWord(&result)
			continue
		}

		l.TokenizeOperator(&result)
	}

	return result
}

func (l *Lexer) TokenizeNumber(result *[]Token) {
	start := l.position

	for unicode.IsDigit(l.Peek()) {
		l.Next()
	}

	numberStr := l.input[start:l.position]
	*result = append(*result, NewToken(NUMBER, numberStr, start))
}

func (l *Lexer) TokenizeWord(result *[]Token) {
	start := l.position

	for {
		c := l.Peek()
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			break
		}
		l.Next()
	}

	word := l.input[start:l.position]

	switch word {
	case "var":
		l.AddToken(result, VAR, word, start)
	case "print":
		l.AddToken(result, PRINT, word, start)
	case "if":
		l.AddToken(result, IF, word, start)
	case "else":
		l.AddToken(result, ELSE, word, start)
	case "while":
		l.AddToken(result, WHILE, word, start)
	default:
		l.AddToken(result, ID, word, start)
	}
}

func (l *Lexer) TokenizeOperator(result *[]Token) {
	current := l.Peek()
	start := l.position

	switch current {
	case '+':
		l.Next()
		l.AddToken(result, PLUS, "+", start)
	case '-':
		l.Next()
		l.AddToken(result, MINUS, "-", start)
	case '*':
		l.Next()
		l.AddToken(result, STAR, "*", start)
	case '/':
		l.Next()
		l.AddToken(result, SLASH, "/", start)
	case '=':
		l.Next()
		l.AddToken(result, EQ, "=", start)
	case ';':
		l.Next()
		l.AddToken(result, SEMICOLON, ";", start)
	default:
		panic(fmt.Sprintf("Unexpected character '%c' at position %d", current, l.position))
	}
}

func (l *Lexer) Peek() rune {
	if l.position >= l.length {
		return 0
	}
	return rune(l.input[l.position])
}

func (l *Lexer) Next() rune {
	if l.position >= l.length {
		return 0
	}
	char := rune(l.input[l.position])
	l.position++
	return char
}

func (l *Lexer) AddToken(result *[]Token, tokenType TokenType, value string, start int) {
	*result = append(*result, NewToken(tokenType, value, start))
}
