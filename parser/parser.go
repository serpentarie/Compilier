package parser

import (
	"errors"
	"fmt"
	"mycompiler/lexer"
	"mycompiler/parser/ast"
	"strconv"
)

// Parser represents a recursive descent parser for the language.
type Parser struct {
	tokens   []lexer.Token
	position int
	errors   []error
}

// NewParser creates a new parser instance with the given tokens.
func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
		errors:   make([]error, 0),
	}
}

// Parse parses the tokens and returns a slice of statements.
func (p *Parser) Parse() []ast.Statement {
	var statements []ast.Statement
	for !p.isAtEnd() {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	return statements
}

// Errors returns all parsing errors encountered.
func (p *Parser) Errors() []error {
	return p.errors
}

// HasErrors returns true if any parsing errors were encountered.
func (p *Parser) HasErrors() bool {
	return len(p.errors) > 0
}

// --- Statements ---

func (p *Parser) parseDeclaration() ast.Statement {
	if p.match(lexer.VAR) {
		return p.parseVarDeclaration()
	}
	return p.parseStatement()
}

func (p *Parser) parseStatement() ast.Statement {
	if p.match(lexer.IF) {
		return p.parseIfStatement()
	}
	if p.match(lexer.WHILE) {
		return p.parseWhileStatement()
	}
	if p.match(lexer.PRINT) {
		return p.parsePrintStatement()
	}
	if p.match(lexer.LBRACE) {
		return &ast.BlockStatement{Statements: p.parseBlock()}
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseVarDeclaration() ast.Statement {
	name := p.consume(lexer.ID, "expected variable name")

	var initializer ast.Expression
	if p.match(lexer.EQ) {
		initializer = p.parseExpression()
	}

	p.consume(lexer.SEMICOLON, "expected ';' after variable declaration")
	return &ast.VarStatement{Name: name.Value, Initializer: initializer}
}

func (p *Parser) parseIfStatement() ast.Statement {
	p.consume(lexer.LPAREN, "expected '(' after 'if'")
	condition := p.parseExpression()
	p.consume(lexer.RPAREN, "expected ')' after if condition")

	thenBranch := p.parseStatement()
	var elseBranch ast.Statement

	if p.match(lexer.ELSE) {
		elseBranch = p.parseStatement()
	}

	return &ast.IfStatement{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}
}

func (p *Parser) parseWhileStatement() ast.Statement {
	p.consume(lexer.LPAREN, "expected '(' after 'while'")
	condition := p.parseExpression()
	p.consume(lexer.RPAREN, "expected ')' after while condition")

	body := p.parseStatement()
	return &ast.WhileStatement{Condition: condition, Body: body}
}

func (p *Parser) parsePrintStatement() ast.Statement {
	value := p.parseExpression()
	p.consume(lexer.SEMICOLON, "expected ';' after print expression")
	return &ast.PrintStatement{Expression: value}
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	expr := p.parseExpression()
	p.consume(lexer.SEMICOLON, "expected ';' after expression")
	return &ast.ExpressionStatement{Expression: expr}
}

func (p *Parser) parseBlock() []ast.Statement {
	var statements []ast.Statement

	for !p.check(lexer.RBRACE) && !p.isAtEnd() {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	p.consume(lexer.RBRACE, "expected '}' after block")
	return statements
}

// --- Expressions ---

func (p *Parser) parseExpression() ast.Expression {
	return p.parseAssignment()
}

// 1. Присваивание
func (p *Parser) parseAssignment() ast.Expression {
	expr := p.parseLogicalOr()

	if p.match(lexer.EQ) {
		value := p.parseAssignment()

		if variable, ok := expr.(*ast.VariableExpression); ok {
			return &ast.AssignExpression{Name: variable.Name, Value: value}
		}

		p.error(fmt.Sprintf("invalid assignment target at position %d", p.previous().Position))
		return nil
	}

	return expr
}

// 2. Логическое ИЛИ (||)
func (p *Parser) parseLogicalOr() ast.Expression {
	expr := p.parseLogicalAnd()

	for p.match(lexer.OR) {
		op := p.previous().Type
		right := p.parseLogicalAnd()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// 3. Логическое И (&&)
func (p *Parser) parseLogicalAnd() ast.Expression {
	expr := p.parseEquality()

	for p.match(lexer.AND) {
		op := p.previous().Type
		right := p.parseEquality()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// 4. Равенство (==, !=)
func (p *Parser) parseEquality() ast.Expression {
	expr := p.parseComparison()

	for p.match(lexer.EQEQ, lexer.NEQ) {
		op := p.previous().Type
		right := p.parseComparison()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// 5. Сравнение (<, >, <=, >=)
func (p *Parser) parseComparison() ast.Expression {
	expr := p.parseTerm()

	for p.match(lexer.LT, lexer.LTEQ, lexer.GT, lexer.GTEQ) {
		op := p.previous().Type
		right := p.parseTerm()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// 6. Сложение/Вычитание (+, -)
func (p *Parser) parseTerm() ast.Expression {
	expr := p.parseFactor()

	for p.match(lexer.PLUS, lexer.MINUS) {
		op := p.previous().Type
		right := p.parseFactor()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// 7. Умножение/Деление (*, /)
func (p *Parser) parseFactor() ast.Expression {
	expr := p.parseUnary()

	for p.match(lexer.STAR, lexer.SLASH) {
		op := p.previous().Type
		right := p.parseUnary()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// 8. Унарные (!, -)
func (p *Parser) parseUnary() ast.Expression {
	if p.match(lexer.EXCL, lexer.MINUS) {
		op := p.previous().Type
		right := p.parseUnary()
		return &ast.UnaryExpression{Operator: op, Right: right}
	}
	return p.parsePrimary()
}

// 9. Primary
func (p *Parser) parsePrimary() ast.Expression {
	if p.match(lexer.NUMBER) {
		valStr := p.previous().Value
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			p.error(fmt.Sprintf("failed to parse number: %s", valStr))
			return nil
		}
		return &ast.NumberExpression{Value: val}
	}

	if p.match(lexer.ID) {
		return &ast.VariableExpression{Name: p.previous().Value}
	}

	if p.match(lexer.LPAREN) {
		expr := p.parseExpression()
		p.consume(lexer.RPAREN, "expected ')' after expression")
		return expr
	}

	token := p.peek()
	p.error(fmt.Sprintf("expected expression at position %d, found: %s", token.Position, token.Type))
	return nil
}

// --- Helpers ---

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.position++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.EOF
}

func (p *Parser) peek() lexer.Token {
	if p.position >= len(p.tokens) {
		// Возвращаем фиктивный EOF токен, если вышли за границы
		return lexer.NewToken(lexer.EOF, "", -1)
	}
	return p.tokens[p.position]
}

func (p *Parser) previous() lexer.Token {
	return p.tokens[p.position-1]
}

func (p *Parser) consume(t lexer.TokenType, message string) lexer.Token {
	if p.check(t) {
		return p.advance()
	}
	token := p.peek()
	p.error(fmt.Sprintf("at position %d: %s, found: %s", token.Position, message, token.Type))
	return token
}

// error records a parsing error.
func (p *Parser) error(message string) {
	p.errors = append(p.errors, errors.New(message))
}
