package main

type TokenType int

const (
	NUMBER TokenType = iota
	ID
	STRING
	VAR

	PRINT
	IF
	ELSE
	WHILE

	PLUS
	MINUS
	STAR
	SLASH
	EQ
	EQEQ
	EXCL
	NEQ
	LT
	GT
	LTEQ
	GTEQ
	AND
	OR

	LPAREN
	RPAREN
	LBRACE
	RBRACE
	SEMICOLON

	EOF
)

func (t TokenType) String() string {
	return [...]string{
		"NUMBER", "ID", "STRING", "VAR",
		"PRINT", "IF", "ELSE", "WHILE",
		"PLUS", "MINUS", "STAR", "SLASH",
		"EQ", "EQEQ", "EXCL", "NEQ",
		"LT", "GT", "LTEQ", "GTEQ",
		"AND", "OR",
		"LPAREN", "RPAREN", "LBRACE", "RBRACE", "SEMICOLON",
		"EOF",
	}[t]
}
