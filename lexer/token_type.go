package lexer

// TokenType represents the type of a lexical token.
type TokenType int

// Token type constants.
const (
	NUMBER TokenType = iota
	ID
	STRING
	TRUE
	FALSE
	VAR
	FUN
	RETURN

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
	LBRACKET
	RBRACKET
	LBRACE
	RBRACE
	COMMA
	SEMICOLON

	EOF
)

// String returns the string representation of the token type.
func (t TokenType) String() string {
	return [...]string{
		"NUMBER", "ID", "STRING", "TRUE", "FALSE", "VAR", "FUN", "RETURN",
		"PRINT", "IF", "ELSE", "WHILE",
		"PLUS", "MINUS", "STAR", "SLASH",
		"EQ", "EQEQ", "EXCL", "NEQ",
		"LT", "GT", "LTEQ", "GTEQ",
		"AND", "OR",
		"LPAREN", "RPAREN", "LBRACKET", "RBRACKET", "LBRACE", "RBRACE", "COMMA", "SEMICOLON",
		"EOF",
	}[t]
}
