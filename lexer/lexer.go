package lexer

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

		if current == '"' {
			l.TokenizeString(&result)
			continue
		}

		if unicode.IsLetter(current) {
			l.TokenizeWord(&result)
			continue
		}

		l.TokenizeOperator(&result)
	}

	result = append(result, NewToken(EOF, "", l.position))
	return result
}

func (l *Lexer) TokenizeNumber(result *[]Token) {
	start := l.position

	for unicode.IsDigit(l.Peek()) {
		l.Next()
	}

	if l.Peek() == '.' {
		l.Next()
		for unicode.IsDigit(l.Peek()) {
			l.Next()
		}
	}

	numberStr := l.input[start:l.position]
	*result = append(*result, NewToken(NUMBER, numberStr, start))
}

func (l *Lexer) TokenizeString(result *[]Token) {
	start := l.position
	l.Next() // opening quote

	for l.Peek() != '"' && l.Peek() != 0 {
		l.Next()
	}

	if l.Peek() == 0 {
		panic(fmt.Sprintf("Unterminated string at position %d", start))
	}

	value := l.input[start+1 : l.position]
	l.Next() // closing quote
	l.AddToken(result, STRING, value, start)
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
	case "fun", "func":
		l.AddToken(result, FUN, word, start)
	case "return":
		l.AddToken(result, RETURN, word, start)
	case "true":
		l.AddToken(result, TRUE, word, start)
	case "false":
		l.AddToken(result, FALSE, word, start)
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
		if l.Peek() == '=' {
			l.Next()
			l.AddToken(result, EQEQ, "==", start)
		} else {
			l.AddToken(result, EQ, "=", start)
		}
	case '<':
		l.Next()
		if l.Peek() == '=' {
			l.Next()
			l.AddToken(result, LTEQ, "<=", start)
		} else {
			l.AddToken(result, LT, "<", start)
		}
	case '>':
		l.Next()
		if l.Peek() == '=' {
			l.Next()
			l.AddToken(result, GTEQ, ">=", start)
		} else {
			l.AddToken(result, GT, ">", start)
		}
	case '!':
		l.Next()
		if l.Peek() == '=' {
			l.Next()
			l.AddToken(result, NEQ, "!=", start)
		} else {
			l.AddToken(result, EXCL, "!", start)
		}
	case '&':
		l.Next()
		if l.Peek() == '&' {
			l.Next()
			l.AddToken(result, AND, "&&", start)
		} else {
			panic(fmt.Sprintf("Unexpected character '%c' at position %d", current, start))
		}
	case '|':
		l.Next()
		if l.Peek() == '|' {
			l.Next()
			l.AddToken(result, OR, "||", start)
		} else {
			panic(fmt.Sprintf("Unexpected character '%c' at position %d", current, start))
		}
	case ';':
		l.Next()
		l.AddToken(result, SEMICOLON, ";", start)
	case '(':
		l.Next()
		l.AddToken(result, LPAREN, "(", start)
	case ')':
		l.Next()
		l.AddToken(result, RPAREN, ")", start)
	case '{':
		l.Next()
		l.AddToken(result, LBRACE, "{", start)
	case '}':
		l.Next()
		l.AddToken(result, RBRACE, "}", start)
	case ',':
		l.Next()
		l.AddToken(result, COMMA, ",", start)
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
