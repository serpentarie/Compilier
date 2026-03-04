package lexer

import "fmt"

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type     TokenType
	Value    string
	Position int
}

// NewToken creates a new token with the given type, value, and position.
func NewToken(tokenType TokenType, value string, position int) Token {
	return Token{
		Type:     tokenType,
		Value:    value,
		Position: position,
	}
}

// String returns a human-readable string representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("Token(Type: %s, Value: '%s') at %d", t.Type, t.Value, t.Position)
}
